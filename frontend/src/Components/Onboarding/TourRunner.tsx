import { useEffect, useState } from "react";
import type { Tour } from "../../types/sdk";
import { Hint } from "./Hint";

interface Props {
  tour: Tour;
  onClose: () => void;
}

export function TourRunner({ tour, onClose }: Props) {
  const [step, setStep] = useState(0);
  const hint = tour.hints[step];
  const [element, setElement] = useState<HTMLElement | null>(null);

  const next = () => {
    console.log("Нажали далее");

    if (step < tour.hints.length - 1) {
      setStep((prev) => prev + 1);
      return;
    }

    console.log("Тур завершен");
    onClose();
  };

  useEffect(() => {
    let interval: number;
    const findElement = () => {
      const target = document.querySelector(
        hint.selector,
      ) as HTMLElement | null;

      if (target) {
        console.log("Элемент найден:", target);

        setElement(target);
        clearInterval(interval);
      }
    };
    findElement();
    interval = window.setInterval(findElement, 200);
    return () => {
      clearInterval(interval);
    };
  }, [step, hint.selector]);

  useEffect(() => {
    if (!element) return;

    const handleAction = () => {
      console.log("Шаг выполнен");

      if (step < tour.hints.length - 1) {
        setStep((prev) => prev + 1);
      } else {
        onClose();
      }
    };

    element.addEventListener("click", handleAction);

    return () => {
      element.removeEventListener("click", handleAction);
    };
  }, [element, step]);

  useEffect(() => {
    if (!element) return;
    if (!hint) return;

    const handleClick = () => {
      setTimeout(() => {
        setStep((prev) => {
          if (prev < tour.hints.length - 1) return prev + 1;
          return prev;
        });
      }, 300);
    };

    element.addEventListener("click", handleClick);

    return () => {
      element.removeEventListener("click", handleClick);
    };
  }, [element, hint, tour.hints.length]);

  if (!element) return null;

  return (
    <Hint
      hint={hint}
      element={element}
      step={step}
      total={tour.hints.length}
      next={next}
      skip={onClose}
    />
  );
}
