import { useEffect, useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import type { Tour } from "../../types/sdk";

export function OnboardingProvider() {
  const [isOpen, setIsOpen] = useState(true);
  const [manualTour, setManualTour] = useState<Tour | null>(null);

  const params = new URLSearchParams(window.location.search);

  const tourId = params.get("tour");
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";

  useEffect(() => {
    window.startOnboarding = (id: string) => {
      const tours = JSON.parse(localStorage.getItem("tours") || "[]");

      console.log("ALL TOURS", tours);
      console.log("SEARCH ID", id);

      const tour = tours.find(
        (item: Tour) => item.id === id && item.trigger_type === "manual",
      );

      console.log("FOUND TOUR", tour);

      if (!tour) return;

      setManualTour(tour);
    };
  }, []);

  if (isBuilder) {
    return (
      <Builder
        onSelect={(selector) => {
          console.log("Выбран selector:", selector);

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

  const getTour = () => {
    const tours = JSON.parse(localStorage.getItem("tours") || "[]");

    if (tourId) {
      const tour = tours.find((tour: Tour) => tour.id === tourId);

      if (isPreview) return tour;

      return tour?.status === "published" ? tour : null;
    }

    return tours.find(
      (tour: Tour) =>
        tour.status === "published" &&
        tour.trigger_type !== "manual" &&
        (!tour.target_path || tour.target_path === window.location.pathname),
    );
  };

  const tour = getTour();

  if (!tour) {
    return null;
  }

  const completedKey = `onboarding_completed_${tour.id}`;
  const showsKey = `onboarding_shows_${tour.id}`;

  const completed = localStorage.getItem(completedKey);
  const shows = Number(localStorage.getItem(showsKey) || 0);

  if (!isPreview) {
    if (tour.audience?.show_once && completed) {
      return null;
    }

    if (tour.audience?.max_shows > 0 && shows >= tour.audience.max_shows) {
      return null;
    }

    if (tour.audience?.only_new && localStorage.getItem("user_seen")) {
      return null;
    }
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
      const timer = setTimeout(startTour, 2000);

      return () => clearTimeout(timer);
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

  if (!isOpen) {
    return null;
  }

  return tour ? <TourRunner tour={tour} onClose={closeTour} /> : null;
}
