import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import type { Tour } from "../../types/sdk";
import { Hint } from "./Hint";
import { sendOnboardingEvents, type EventType } from "./events";
import { API_URL, APP_KEY, SELECTOR_MISSING_TIMEOUT } from "../../Utils/constants";

interface Props {
  tour: Tour;
  onClose: () => void;
  isPreview?: boolean;
}

interface EventToSend {
  type: EventType;
  hintId?: string | null;
  payload?: Record<string, unknown>;
}

interface PreviewState {
  onboardingPreviewTourId?: string;
  onboardingPreviewHintId?: string;
}

const normalizePath = (path?: string) =>
  path ? path.split(/[?#]/)[0].replace(/\/+$/, "") || "/" : "";

export function TourRunner({ tour, onClose, isPreview = false }: Props) {
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
  const hintPage = normalizePath(hint?.page_path || tour.target_path);
  const currentPage = normalizePath(location.pathname);
  const isHintPage = !hintPage || hintPage === currentPage;

  const sendEvents = useCallback(
    async (events: EventToSend[]) => {
      if (isPreview || !tour.version_id) return;

      try {
        await sendOnboardingEvents(
          { apiUrl: API_URL, appKey: APP_KEY },
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
    [isPreview, tour],
  );

  useEffect(() => {
    if (startedRef.current) return;
    startedRef.current = true;
    void sendEvents([{ type: "tour_started" }]);
  }, [sendEvents]);

  useEffect(() => {
    setElement(null);

    if (!hint || !isHintPage || hint.placement === "center") return;

    if (!hint.selector) {
      if (!missingRef.current.has(hint.id)) {
        missingRef.current.add(hint.id);
        void sendEvents([
          {
            type: "selector_missing",
            hintId: hint.id,
            payload: { selector: null, reason: "selector_empty" },
          },
        ]);
      }
      return;
    }

    let interval: number | undefined;
    let timeout: number | undefined;

    const cleanup = () => {
      if (interval !== undefined) clearInterval(interval);
      if (timeout !== undefined) clearTimeout(timeout);
    };

    const reportMissing = (reason: string) => {
      if (missingRef.current.has(hint.id)) return;
      missingRef.current.add(hint.id);

      void sendEvents([
        {
          type: "selector_missing",
          hintId: hint.id,
          payload: { selector: hint.selector, reason },
        },
      ]);
    };

    const findElement = () => {
      try {
        const target = document.querySelector(
          hint.selector!,
        ) as HTMLElement | null;

        if (!target) return;

        setElement(target);
        cleanup();
      } catch {
        reportMissing("invalid_selector");
        cleanup();
      }
    };

    findElement();

    interval = window.setInterval(findElement, 200);
    timeout = window.setTimeout(
      () => reportMissing("element_not_found"),
      SELECTOR_MISSING_TIMEOUT,
    );

    return cleanup;
  }, [hint, isHintPage, sendEvents]);

  useEffect(() => {
    if (!hint || !isHintPage) return;
    if (hint.placement !== "center" && !element) return;
    if (shownRef.current.has(hint.id)) return;

    shownRef.current.add(hint.id);
    void sendEvents([{ type: "hint_shown", hintId: hint.id }]);
  }, [hint, element, isHintPage, sendEvents]);

  useEffect(() => {
    if (!element || !hint || !isHintPage || hint.placement === "center") return;

    const rect = element.getBoundingClientRect();

    const visible =
      rect.top >= 0 &&
      rect.left >= 0 &&
      rect.bottom <= window.innerHeight &&
      rect.right <= window.innerWidth;

    if (!visible) {
      element.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "nearest",
      });
    }
  }, [element, hint, isHintPage]);

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

  if (!hint || !isHintPage) return null;
  if (hint.placement !== "center" && !element) return null;

  return (
    <Hint
      hint={hint}
      element={element}
      step={step}
      total={hints.length}
      next={() => void goToNextStep()}
      skip={skip}
    />
  );
}