import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Tour } from "../../types/sdk";
import { Hint } from "./Hint";
import { sendOnboardingEvents, type EventType } from "./events";

interface Props {
  tour: Tour;
  onClose: () => void;
  isPreview?: boolean;
}

const API_URL = "http://localhost:8080";
const APP_KEY = "pk_4e17b539-07c4-429a-9b30-12a34b2059f5";
const SELECTOR_MISSING_TIMEOUT = 5000;

interface EventToSend {
  type: EventType;
  hintId?: string | null;
  payload?: Record<string, unknown>;
}

export function TourRunner({ tour, onClose, isPreview = false }: Props) {
  const sortedHints = useMemo(() => {
    return [...(tour.hints || [])].sort((a, b) => a.step - b.step);
  }, [tour.hints]);

  const initialStep = useMemo(() => {
    if (!tour.current_hint_id) return 0;
    const index = sortedHints.findIndex((hint) => hint.id === tour.current_hint_id);
    return index >= 0 ? index : 0;
  }, [tour.current_hint_id, sortedHints]);

  const [step, setStep] = useState(initialStep);
  const [element, setElement] = useState<HTMLElement | null>(null);
  const hint = sortedHints[step];
  const shownHintsRef = useRef<Set<string>>(new Set());
  const missingSelectorsRef = useRef<Set<string>>(new Set());
  const tourStartedRef = useRef(false);

  const sendEvents = useCallback(
    (events: EventToSend[]) => {
      if (isPreview) return;

      const versionId = tour.version_id;
      if (!versionId) {
        console.warn("[Onboarding] tour_version_id is missing. Events were not sent.", tour);
        return;
      }

      void sendOnboardingEvents(
        { apiUrl: API_URL, appKey: APP_KEY },
        events.map((event) => ({
          type: event.type,
          tourId: tour.id,
          tourVersionId: versionId,
          hintId: event.hintId ?? null,
          payload: event.payload ?? {},
        })),
      ).catch((error) => {
        console.error("[Onboarding] event sending failed", error);
      });
    },
    [isPreview, tour],
  );

  useEffect(() => {
    if (tourStartedRef.current) return;
    tourStartedRef.current = true;
    sendEvents([{ type: "tour_started" }]);
  }, [tour.id, sendEvents]);

  useEffect(() => {
    if (!hint) return;
    if (hint.placement === "center") return;
    if (!hint.selector) {
      console.warn("[Onboarding] selector is missing", hint);
      if (!missingSelectorsRef.current.has(hint.id)) {
        missingSelectorsRef.current.add(hint.id);
        sendEvents([
          {
            type: "selector_missing",
            hintId: hint.id,
            payload: { selector: null, reason: "selector_empty" },
          },
        ]);
      }
      return;
    }

    let _intervalRef: number | undefined;
    let _missingTimeoutRef: number | undefined;

    const reportSelectorMissing = (reason: string) => {
      if (missingSelectorsRef.current.has(hint.id)) return;
      missingSelectorsRef.current.add(hint.id);
      console.warn("[Onboarding] selector missing", hint.selector);
      sendEvents([
        {
          type: "selector_missing",
          hintId: hint.id,
          payload: { selector: hint.selector, reason },
        },
      ]);
    };

    const findElement = () => {
      if (!hint.selector) return;
      try {
        const target = document.querySelector(hint.selector) as HTMLElement | null;
        if (target) {
          console.log("[Onboarding] element found", target);
          setElement(target);
          if (_intervalRef !== undefined) window.clearInterval(_intervalRef);
          if (_missingTimeoutRef !== undefined) window.clearTimeout(_missingTimeoutRef);
        }
      } catch (error) {
        console.error("[Onboarding] invalid selector", hint.selector, error);
        reportSelectorMissing("invalid_selector");
        if (_intervalRef !== undefined) window.clearInterval(_intervalRef);
        if (_missingTimeoutRef !== undefined) window.clearTimeout(_missingTimeoutRef);
      }
    };

    findElement();
    _intervalRef = window.setInterval(findElement, 200);
    _missingTimeoutRef = window.setTimeout(() => {
      reportSelectorMissing("element_not_found");
    }, SELECTOR_MISSING_TIMEOUT);

    return () => {
      if (_intervalRef !== undefined) window.clearInterval(_intervalRef);
      if (_missingTimeoutRef !== undefined) window.clearTimeout(_missingTimeoutRef);
    };
  }, [step, hint?.id, hint?.selector, hint?.placement, sendEvents]);

  useEffect(() => {
    if (!hint) return;
    const canShow = hint.placement === "center" || element !== null;
    if (!canShow) return;
    if (shownHintsRef.current.has(hint.id)) return;

    shownHintsRef.current.add(hint.id);
    sendEvents([{ type: "hint_shown", hintId: hint.id }]);
  }, [hint, element, sendEvents]);

  useEffect(() => {
    if (!element || !hint) return;
    if (hint.placement === "center") return;

    const rect = element.getBoundingClientRect();
    const isVisible =
      rect.top >= 0 && rect.left >= 0 && rect.bottom <= window.innerHeight && rect.right <= window.innerWidth;

    if (!isVisible) {
      element.scrollIntoView({ behavior: "smooth", block: "center", inline: "nearest" });
    }
  }, [element, hint]);

  const goToNextStep = useCallback(() => {
    if (!hint) return;

    const nextStep = step + 1;
    if (nextStep >= sortedHints.length) {
      console.log("[Onboarding] tour completed", tour.id);
      sendEvents([
        { type: "hint_completed", hintId: hint.id },
        { type: "tour_completed" },
      ]);
      setElement(null);
      onClose();
      return;
    }

    sendEvents([{ type: "hint_completed", hintId: hint.id }]);
    setElement(null);
    setStep(nextStep);
  }, [hint, step, sortedHints, tour.id, sendEvents, onClose]);

  useEffect(() => {
    if (!element || !hint) return;
    const handleElementClick = () => {
      goToNextStep();
    };
    element.addEventListener("click", handleElementClick);
    return () => {
      element.removeEventListener("click", handleElementClick);
    };
  }, [element, hint, goToNextStep]);

  const next = useCallback(() => {
    goToNextStep();
  }, [goToNextStep]);

  const skip = useCallback(() => {
    if (!hint) return;
    sendEvents([{ type: "tour_dismissed" }]);
    setElement(null);
    onClose();
  }, [hint, sendEvents, onClose]);

  if (!hint) return null;
  if (hint.placement !== "center" && !element) return null;

  return (
    <Hint
      hint={hint}
      element={element}
      step={step}
      total={sortedHints.length}
      next={next}
      skip={skip}
    />
  );
}