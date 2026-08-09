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
    if (!tour.current_hint_id) {
      return 0;
    }

    const index = sortedHints.findIndex(
      (hint) => hint.id === tour.current_hint_id,
    );

    return index >= 0 ? index : 0;
  }, [tour.current_hint_id, sortedHints]);

  const [step, setStep] = useState(initialStep);

  const [element, setElement] = useState<HTMLElement | null>(null);

  const hint = sortedHints[step];

  /*
   * Нужны, чтобы внутри одного запуска
   * случайно не отправить hint_shown /
   * selector_missing несколько раз.
   */
  const shownHintsRef = useRef<Set<string>>(new Set());

  const missingSelectorsRef = useRef<Set<string>>(new Set());

  const tourStartedRef = useRef(false);

  // =========================
  // EVENTS
  // =========================

  const sendEvents = useCallback(
    (events: EventToSend[]) => {
      // В режиме предпросмотра ничего
      // не отправляем в backend analytics.
      if (isPreview) {
        console.log("[Onboarding Preview] events skipped", events);

        return;
      }

      const versionId = tour.version_id;

      if (!versionId) {
        console.warn(
          "[Onboarding] tour_version_id is missing. Events were not sent.",
          tour,
        );

        return;
      }

      void sendOnboardingEvents(
        {
          apiUrl: API_URL,
          appKey: APP_KEY,
        },
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
    [isPreview, tour.id, tour.version_id],
  );

  // =========================
  // TOUR STARTED
  // =========================

  useEffect(() => {
    if (tourStartedRef.current) {
      return;
    }

    tourStartedRef.current = true;

    console.log("[Onboarding] tour started", tour.id);

    sendEvents([
      {
        type: "tour_started",
      },
    ]);
  }, [tour.id, sendEvents]);

  // =========================
  // СИНХРОНИЗАЦИЯ С BACKEND
  // =========================

  useEffect(() => {
    setStep(initialStep);
  }, [initialStep]);

  // =========================
  // ПОИСК ELEMENT
  // =========================

  useEffect(() => {
    setElement(null);

    if (!hint) {
      return;
    }

    // center не требует selector
    if (hint.placement === "center") {
      return;
    }

    // Selector вообще не указан
    if (!hint.selector) {
      console.warn("[Onboarding] selector is missing", hint);

      if (!missingSelectorsRef.current.has(hint.id)) {
        missingSelectorsRef.current.add(hint.id);

        sendEvents([
          {
            type: "selector_missing",
            hintId: hint.id,
            payload: {
              selector: null,
              reason: "selector_empty",
            },
          },
        ]);
      }

      return;
    }

    let interval: number | undefined;

    let missingTimeout: number | undefined;

    const reportSelectorMissing = (reason: string) => {
      if (missingSelectorsRef.current.has(hint.id)) {
        return;
      }

      missingSelectorsRef.current.add(hint.id);

      console.warn("[Onboarding] selector missing", hint.selector);

      sendEvents([
        {
          type: "selector_missing",
          hintId: hint.id,
          payload: {
            selector: hint.selector,
            reason,
          },
        },
      ]);
    };

    const findElement = () => {
      if (!hint.selector) {
        return;
      }

      let target: HTMLElement | null = null;

      try {
        target = document.querySelector(hint.selector) as HTMLElement | null;
      } catch (error) {
        console.error("[Onboarding] invalid selector", hint.selector, error);

        reportSelectorMissing("invalid_selector");

        if (interval !== undefined) {
          window.clearInterval(interval);
        }

        if (missingTimeout !== undefined) {
          window.clearTimeout(missingTimeout);
        }

        return;
      }

      if (!target) {
        return;
      }

      console.log("[Onboarding] element found", target);

      setElement(target);

      if (interval !== undefined) {
        window.clearInterval(interval);
      }

      if (missingTimeout !== undefined) {
        window.clearTimeout(missingTimeout);
      }
    };

    findElement();

    interval = window.setInterval(findElement, 200);

    missingTimeout = window.setTimeout(() => {
      reportSelectorMissing("element_not_found");
    }, SELECTOR_MISSING_TIMEOUT);

    return () => {
      if (interval !== undefined) {
        window.clearInterval(interval);
      }

      if (missingTimeout !== undefined) {
        window.clearTimeout(missingTimeout);
      }
    };
  }, [step, hint?.id, hint?.selector, hint?.placement, sendEvents]);

  // =========================
  // HINT SHOWN
  // =========================

  useEffect(() => {
    if (!hint) {
      return;
    }

    const canShow = hint.placement === "center" || element !== null;

    if (!canShow) {
      return;
    }

    if (shownHintsRef.current.has(hint.id)) {
      return;
    }

    shownHintsRef.current.add(hint.id);

    console.log("[Onboarding] hint shown", hint.id);

    sendEvents([
      {
        type: "hint_shown",
        hintId: hint.id,
      },
    ]);
  }, [hint, element, sendEvents]);

  // =========================
  // ПРОКРУТКА К ELEMENT
  // =========================

  useEffect(() => {
    if (!element || !hint) {
      return;
    }

    if (hint.placement === "center") {
      return;
    }

    const rect = element.getBoundingClientRect();

    const isVisible =
      rect.top >= 0 &&
      rect.left >= 0 &&
      rect.bottom <= window.innerHeight &&
      rect.right <= window.innerWidth;

    if (!isVisible) {
      element.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "nearest",
      });
    }
  }, [element, hint]);

  // =========================
  // ПЕРЕХОД К СЛЕДУЮЩЕМУ ШАГУ
  // =========================

  const goToNextStep = useCallback(() => {
    if (!hint) {
      return;
    }

    console.log("[Onboarding] hint completed", hint.id);

    const nextStep = step + 1;

    // Последний шаг
    if (nextStep >= sortedHints.length) {
      console.log("[Onboarding] tour completed", tour.id);

      /*
       * Оба события отправляем одним batch.
       *
       * Так backend получает завершение
       * последней подсказки и всего тура
       * одним запросом.
       */
      sendEvents([
        {
          type: "hint_completed",
          hintId: hint.id,
        },
        {
          type: "tour_completed",
        },
      ]);

      setElement(null);
      onClose();

      return;
    }

    sendEvents([
      {
        type: "hint_completed",
        hintId: hint.id,
      },
    ]);

    console.log("[Onboarding] move to step", nextStep, sortedHints[nextStep]);

    setElement(null);
    setStep(nextStep);
  }, [hint, step, sortedHints, tour.id, sendEvents, onClose]);

  // =========================
  // КЛИК ПО ЦЕЛЕВОМУ ELEMENT
  // =========================

  useEffect(() => {
    if (!element || !hint) {
      return;
    }

    const handleElementClick = () => {
      console.log("[Onboarding] target element clicked", hint.id);

      /*
       * preventDefault НЕ делаем.
       *
       * Если элемент хоста выполняет
       * переход или другое действие,
       * оно продолжает работать.
       *
       * fetch событий использует
       * keepalive: true.
       */
      goToNextStep();
    };

    element.addEventListener("click", handleElementClick);

    return () => {
      element.removeEventListener("click", handleElementClick);
    };
  }, [element, hint, goToNextStep]);

  // =========================
  // ДАЛЕЕ
  // =========================

  const next = () => {
    goToNextStep();
  };

  // =========================
  // ЗАКРЫТЬ / ПРОПУСТИТЬ
  // =========================

  const skip = () => {
    if (!hint) {
      return;
    }

    console.log("[Onboarding] tour dismissed", tour.id);

    sendEvents([
      {
        type: "tour_dismissed",
      },
    ]);

    setElement(null);
    onClose();
  };

  if (!hint) {
    return null;
  }

  // Обычная подсказка ждёт DOM element
  if (hint.placement !== "center" && !element) {
    return null;
  }

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
