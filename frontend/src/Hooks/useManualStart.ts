import { useEffect } from "react";
import type { Tour } from "../types";

declare global {
  interface Window {
    startOnboarding?: () => void;
  }
}

export function useManualStart(tour: Tour | null, isBuilder: boolean, setIsOpen: (open: boolean) => void) {
  useEffect(() => {
    if (isBuilder) return;

    const handleManualStart = () => {
      if (!tour) {
        console.warn("[Onboarding] manual start: tour is not loaded");
        return;
      }

      if (tour.trigger_type !== "manual") {
        console.warn("[Onboarding] manual start: active tour is not manual");
        return;
      }
      setIsOpen(true);
    };
    window.startOnboarding = () => {
      window.dispatchEvent(new CustomEvent("start-onboarding"));
    };
    
    window.addEventListener("start-onboarding", handleManualStart);
    return () => {
      window.removeEventListener("start-onboarding", handleManualStart);
    };
  }, [isBuilder, tour, setIsOpen]);
}