import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import type { Tour } from "../types/tour";
import type { EventToSend, PreviewState } from "../types";
import { sendOnboardingEvents } from "../Components/Onboarding/events";
import { SELECTOR_MISSING_TIMEOUT } from "../Utils/constants";
import { API_URL } from "../Config/env";

export function useTourRunner(
  tour: Tour,
  appKey: string,
  isPreview: boolean,
  onClose: () => void,
) {
  const navigate = useNavigate();
  const location = useLocation();
  const advancingRef = useRef(false);
  const shownRef = useRef(new Set<string>());
  const missingRef = useRef(new Set<string>());
  const startedRef = useRef(false);

  const hints = useMemo(
    () => [...(tour.hints || [])].sort((a, b) => a.step - b.step),
    [tour.hints],
  );

  const initialStep = useMemo(() => {
    const state = location.state as PreviewState | null;
    const hintId =
      isPreview &&
      state?.onboardingPreviewTourId === tour.id &&
      state.onboardingPreviewHintId
        ? state.onboardingPreviewHintId
        : tour.current_hint_id;

    if (!hintId) return 0;
    const index = hints.findIndex((hint) => hint.id === hintId);
    return index >= 0 ? index : 0;
  }, [isPreview, location.state, tour.id, tour.current_hint_id, hints]);

  const [step, setStep] = useState(initialStep);
  const [element, setElement] = useState<HTMLElement | null>(null);
  const hint = hints[step];

  const normalizePath = (path?: string) =>
    path ? path.split(/[?#]/)[0].replace(/\/+$/, "") || "/" : "";

  const hintPage = normalizePath(hint?.page_path || tour.target_path);
  const currentPage = normalizePath(location.pathname);
  const isHintPage = !hintPage || hintPage === currentPage;

  const sendEvents = useCallback(
    async (events: EventToSend[]) => {
      if (isPreview || !tour.version_id) return;

      if (!appKey) {
        console.error("[Onboarding] appKey is not configured");
        return;
      }

      try {
        await sendOnboardingEvents(
          {
            apiUrl: API_URL,
            appKey,
          },
          events.map((event) => ({
            type: event.type,
            tourId: tour.id,
            tourVersionId: tour.version_id!,
            hintId: event.hintId ?? null,
            payload: event.payload ?? {},
          })),
        );
      } catch (error) {
        console.error("[Onboarding] event error", error);
      }
    },
    [appKey, isPreview, tour],
  );

  const skipCurrentStep = useCallback(
    async (reason: string) => {
      if (!hint || advancingRef.current) return;

      advancingRef.current = true;

      try {
        await sendEvents([
          {
            type: "hint_skipped",
            hintId: hint.id,
            payload: {
              reason,
            },
          },
        ]);

        const nextStep = step + 1;

        if (nextStep >= hints.length) {
          await sendEvents([
            {
              type: "tour_completed",
            },
          ]);

          setElement(null);
          onClose();
          return;
        }

        const nextHint = hints[nextStep];
        const nextPage = nextHint.page_path || tour.target_path;

        setElement(null);
        setStep(nextStep);

        if (
          nextPage &&
          normalizePath(nextPage) !== normalizePath(location.pathname)
        ) {
          navigate(
            {
              pathname: nextPage,
              search: isPreview ? location.search : "",
            },
            isPreview
              ? {
                  state: {
                    onboardingPreviewTourId: tour.id,
                    onboardingPreviewHintId: nextHint.id,
                  },
                }
              : undefined,
          );
        }
      } finally {
        advancingRef.current = false;
      }
    },
    [
      hint,
      step,
      hints,
      tour.id,
      tour.target_path,
      location.pathname,
      location.search,
      isPreview,
      navigate,
      sendEvents,
      onClose,
    ],
  );

  useEffect(() => {
    if (startedRef.current) return;
    startedRef.current = true;
    void sendEvents([{ type: "tour_started" }]);
  }, [sendEvents]);

  useEffect(() => {
    if (!hint || !isHintPage) return;
    if (hint.placement !== "center" && !element) return;
    if (shownRef.current.has(hint.id)) return;

    shownRef.current.add(hint.id);
    void sendEvents([{ type: "hint_shown", hintId: hint.id }]);
  }, [hint, element, isHintPage, sendEvents]);

  useEffect(() => {
    if (!hint || !isHintPage || hint.placement === "center") return;

    if (!hint.selector) {
      if (!missingRef.current.has(hint.id)) {
        missingRef.current.add(hint.id);
        void sendEvents([
          {
            type: "selector_missing",
            hintId: hint.id,
            payload: {
              selector: null,
              reason: "selector_empty",
            },
          },
        ]).then(() => {
          void skipCurrentStep("selector_empty");
        });
      }
      return;
    }

    const reportMissing = (reason: string) => {
      if (missingRef.current.has(hint.id)) return;
      missingRef.current.add(hint.id);
      void sendEvents([
        {
          type: "selector_missing",
          hintId: hint.id,
          payload: {
            selector: hint.selector,
            reason,
          },
        },
      ]).then(() => {
        void skipCurrentStep(reason);
      });
    };

    if (!hint.wait_for_selector) {
      try {
        const target = document.querySelector(
          hint.selector,
        ) as HTMLElement | null;
        if (!target) {
          reportMissing("element_not_found");
          return;
        }
        const frameId = requestAnimationFrame(() => setElement(target));
        return () => cancelAnimationFrame(frameId);
      } catch {
        reportMissing("invalid_selector");
      }
      return;
    }

    const findElement = () => {
      try {
        const target = document.querySelector(
          hint.selector!,
        ) as HTMLElement | null;
        if (!target) return false;
        setElement(target);
        return true;
      } catch {
        reportMissing("invalid_selector");
        return true;
      }
    };

    const interval = window.setInterval(() => {
      if (findElement()) {
        window.clearInterval(interval);
        window.clearTimeout(timeout);
      }
    }, 200);

    const timeout = window.setTimeout(() => {
      window.clearInterval(interval);
      if (!findElement()) {
        reportMissing("element_not_found");
      }
    }, SELECTOR_MISSING_TIMEOUT);

    if (findElement()) {
      window.clearInterval(interval);
      window.clearTimeout(timeout);
    }

    return () => {
      window.clearInterval(interval);
      window.clearTimeout(timeout);
    };
  }, [hint, isHintPage, sendEvents, skipCurrentStep, hint.wait_for_selector]);

  const goToNextStep = useCallback(async () => {
    if (!hint || advancingRef.current) return;

    advancingRef.current = true;

    try {
      await sendEvents([{ type: "hint_completed", hintId: hint.id }]);

      const nextStep = step + 1;

      if (nextStep >= hints.length) {
        await sendEvents([{ type: "tour_completed" }]);
        setElement(null);
        onClose();
        return;
      }

      const nextHint = hints[nextStep];
      const nextPage = nextHint.page_path || tour.target_path;

      setElement(null);
      setStep(nextStep);

      if (
        nextPage &&
        normalizePath(nextPage) !== normalizePath(location.pathname)
      ) {
        navigate(
          {
            pathname: nextPage,
            search: isPreview ? location.search : "",
          },
          isPreview
            ? {
                state: {
                  onboardingPreviewTourId: tour.id,
                  onboardingPreviewHintId: nextHint.id,
                },
              }
            : undefined,
        );
      }
    } finally {
      advancingRef.current = false;
    }
  }, [
    hint,
    step,
    hints,
    tour.id,
    tour.target_path,
    location.pathname,
    location.search,
    isPreview,
    navigate,
    sendEvents,
    onClose,
  ]);

  useEffect(() => {
    if (!element || !hint || !isHintPage) return;

    const handler = () => void goToNextStep();

    element.addEventListener("click", handler);
    return () => element.removeEventListener("click", handler);
  }, [element, hint, isHintPage, goToNextStep]);

  const skip = useCallback(() => {
    void sendEvents([{ type: "tour_dismissed" }]);
    setElement(null);
    onClose();
  }, [sendEvents, onClose]);

  return {
    step,
    setStep,
    element,
    setElement,
    hint,
    hints,
    isHintPage,
    next: () => void goToNextStep(),
    skip,
    hasHint:
      !!hint &&
      (!hint.placement || hint.placement !== "center" ? !!element : true),
  };
}
