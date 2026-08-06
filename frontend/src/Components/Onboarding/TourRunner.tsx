import { useEffect, useState } from "react";
import type { Tour } from "../../types/sdk";
import { Hint } from "./Hint";
import { useNavigate } from "react-router-dom";

interface Props {
  tour: Tour;
  onClose: () => void;
}

export function TourRunner({ tour, onClose }: Props) {
  const storageKey = `tour_step_${tour.id}`;
  const navigate = useNavigate();

  const [step, setStep] = useState(() => {
    const saved = sessionStorage.getItem(storageKey);
    return saved ? Number(saved) : 0;
  });

  const hint = tour.hints[step];
  const [element, setElement] = useState<HTMLElement | null>(null);

  const changeStep = (newStep: number) => {
    sessionStorage.setItem(`tour_step_${tour.id}`, String(newStep));
    setStep(newStep);
  };

  const next = () => {
    const nextStep = step + 1;
    if (nextStep >= tour.hints.length) {
      sessionStorage.removeItem(storageKey);
      onClose();
      return;
    }

    const nextHint = tour.hints[nextStep];
    changeStep(nextStep);

    if (nextHint.path && nextHint.path !== window.location.pathname) {
      navigate(nextHint.path);
    }
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
      console.log("Шаг выполнен", step);

      const nextStep = step + 1;

      if (nextStep >= tour.hints.length) {
        sessionStorage.removeItem(`tour_step_${tour.id}`);

        onClose();

        return;
      }

      const nextHint = tour.hints[nextStep];

      changeStep(nextStep);

      if (nextHint.path && nextHint.path !== window.location.pathname) {
        navigate(nextHint.path);
      }
    };

    element.addEventListener("click", handleAction);

    return () => {
      element.removeEventListener("click", handleAction);
    };
  }, [element, step]);

  useEffect(() => {
    if (!element) return;

    if (hint.placement === "center") {
      return;
    }

    const rect = element.getBoundingClientRect();

    const isVisible = rect.top >= 0 && rect.bottom <= window.innerHeight;

    if (!isVisible) {
      element.scrollIntoView({
        behavior: "smooth",

        block: "center",

        inline: "nearest",
      });
    }
  }, [element, hint.placement]);

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
