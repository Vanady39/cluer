import { useEffect, useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import type { Tour } from "../../types/sdk";
import { usePublishedTour } from "../../Hooks/usePublishedTour";

export function OnboardingProvider() {
  const [isOpen, setIsOpen] = useState(true);
  const [manualTour, setManualTour] = useState<Tour | null>(null);
  const params = new URLSearchParams(window.location.search);
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";
  const { data: tour, isLoading } = usePublishedTour(window.location.pathname);

  useEffect(() => {
    window.startOnboarding = (id: string) => {
      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      const tour = tours.find(
        (item: Tour) => item.id === id && item.trigger_type === "manual",
      );
      if (!tour) return;
      setManualTour(tour);
    };
  }, []);

  if (isBuilder) {
    return (
      <Builder
        onSelect={(selector) => {
          localStorage.setItem("selected_element", selector);
          if (window.opener) {
            window.opener.postMessage(
              {
                type: "SELECTOR_SELECTED",
                selector,
              },
              "*",
            );
            window.close();
          }
        }}
      />
    );
  }

  if (manualTour) {
    return (
      <>
        <TourRunner
          tour={manualTour}
          onClose={() => {
            setManualTour(null);
          }}
        />
      </>
    );
  }

  if (isLoading) return null;
  if (!tour) return null;

  const completedKey = `onboarding_completed_${tour.id}`;
  const showsKey = `onboarding_shows_${tour.id}`;
  const shows = Number(localStorage.getItem(showsKey) || 0);

  if (!isPreview) {
    if (tour.audience?.show_once && localStorage.getItem(completedKey)) return null;
    if (tour.audience?.max_shows > 0 && shows >= tour.audience.max_shows) return null;
    if (tour.audience?.only_new && localStorage.getItem("user_seen")) return null;
  }

  const startTour = () => {
    if (!isPreview) {
      localStorage.setItem(showsKey, String(shows + 1));
      localStorage.setItem("user_seen", "true");
    }
  };

  const closeTour = () => {
    setIsOpen(false);
    if (!isPreview && tour.audience?.show_once) {
      localStorage.setItem(completedKey, "true");
    }
  };

  useEffect(() => {
    if (isPreview) return;

    if (tour.trigger_type === "delay") {
      return () => clearTimeout(setTimeout(startTour, 2000));
    }

    if (tour.trigger_type === "exit_intent") {
      const handleExit = (event: MouseEvent) => {
        if (event.clientY <= 0) {
          startTour();
        }
      };

      document.addEventListener("mouseleave", handleExit);
      return () => {
        document.removeEventListener("mouseleave", handleExit);
      };
    }

    if (tour.trigger_type === "on_load") {
      startTour();
    }
  }, [tour]);

  if (!isOpen) return null;

  return tour ? <TourRunner tour={tour} onClose={closeTour} /> : null;
}
