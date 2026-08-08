import { useEffect, useMemo, useState } from "react";
import type { Tour } from "../../types/sdk";
import { Hint } from "./Hint";

interface Props {
  tour: Tour;
  onClose: () => void;
}

export function TourRunner({
  tour,
  onClose,
}: Props) {
  const sortedHints = useMemo(() => {
    return [...(tour.hints || [])].sort(
      (a, b) => a.step - b.step,
    );
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

  const [element, setElement] =
    useState<HTMLElement | null>(null);

  const hint = sortedHints[step];

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

    if (!hint.selector) {
      console.warn(
        "[Onboarding] selector is missing",
        hint,
      );

      return;
    }

    let interval: number | undefined;

    const findElement = () => {
      if (!hint.selector) {
        return;
      }

      const target = document.querySelector(
        hint.selector,
      ) as HTMLElement | null;

      if (!target) {
        return;
      }

      console.log(
        "[Onboarding] element found",
        target,
      );

      setElement(target);

      if (interval !== undefined) {
        window.clearInterval(interval);
      }
    };

    findElement();

    interval = window.setInterval(
      findElement,
      200,
    );

    return () => {
      if (interval !== undefined) {
        window.clearInterval(interval);
      }
    };
  }, [
    step,
    hint?.id,
    hint?.selector,
    hint?.placement,
  ]);

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

  const goToNextStep = () => {
    if (!hint) {
      return;
    }

    console.log(
      "[Onboarding] hint completed",
      hint.id,
    );

    const nextStep = step + 1;

    if (nextStep >= sortedHints.length) {
      console.log(
        "[Onboarding] tour completed",
        tour.id,
      );

      setElement(null);
      onClose();

      return;
    }

    console.log(
      "[Onboarding] move to step",
      nextStep,
      sortedHints[nextStep],
    );

    setElement(null);
    setStep(nextStep);
  };

  // =========================
  // КЛИК ПО ЦЕЛЕВОМУ ELEMENT
  // =========================

  useEffect(() => {
    if (!element || !hint) {
      return;
    }

    const handleElementClick = () => {
      console.log(
        "[Onboarding] target element clicked",
        hint.id,
      );

      /*
       * НЕ делаем preventDefault.
       *
       * Если кнопка хоста должна открыть страницу,
       * React Router / сайт продолжит работать сам.
       *
       * SDK только переводит onboarding
       * на следующий шаг.
       */
      goToNextStep();
    };

    element.addEventListener(
      "click",
      handleElementClick,
    );

    return () => {
      element.removeEventListener(
        "click",
        handleElementClick,
      );
    };
  }, [element, step, hint?.id]);

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

    console.log(
      "[Onboarding] tour dismissed",
      tour.id,
    );

    setElement(null);
    onClose();
  };

  if (!hint) {
    return null;
  }

  // Обычная подсказка ждёт DOM element
  if (
    hint.placement !== "center" &&
    !element
  ) {
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