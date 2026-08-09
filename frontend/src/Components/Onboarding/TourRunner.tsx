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
    sessionStorage.setItem(storageKey, String(newStep));
    setStep(newStep);
  };

  const next = () => {
    const nextStep = step + 1;
    if (nextStep >= tour.hints.length) {
      sessionStorage.removeItem(`tour_step_${tour.id}`);
      setElement(null);
      onClose();
      return;
    }

    const nextHint = tour.hints[nextStep];
    changeStep(nextStep);

    if (nextHint.page_path && nextHint.page_path !== window.location.pathname) {
      navigate(nextHint.page_path);
    }
  };

  useEffect(() => {
    if (!hint) return;

    const selector = hint.selector;
    const findElement = () => {
      const target = selector
        ? (document.querySelector(selector) as HTMLElement | null)
        : null;

      if (target) {
        console.log("Элемент найден:", target);
        setElement(target);
        clearInterval(interval);
      }
    };

    // Started before the first lookup so that finding the target straight away
    // stops the poll: previously that first hit cleared an interval that did
    // not exist yet, and the timer kept running.
    const interval = window.setInterval(findElement, 200);
    findElement();

    return () => {
      clearInterval(interval);
    };
  }, [step, hint?.selector]);

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

      if (
        nextHint.page_path &&
        nextHint.page_path !== window.location.pathname
      ) {
        navigate(nextHint.page_path);
      }
    };

    element.addEventListener("click", handleAction);

    return () => {
      element.removeEventListener("click", handleAction);
    };
  }, [element, step]);

  useEffect(() => {
    if (!element || !hint) return;

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
  }, [element, hint?.placement]);

  // Both guards live below the hooks so the hook order stays identical on every
  // render; the early return used to sit above them.
  if (!hint || !element) return null;

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
