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

  const closeTour = () => {
    setIsOpen(false);
    if (!isPreview && tour.audience?.show_once) {
      localStorage.setItem(completedKey, "true");
    }
  };

  return (
    <>
      <AutoStartTrigger tour={tour} isPreview={isPreview} showsKey={showsKey} />
      {isOpen ? <TourRunner tour={tour} onClose={closeTour} /> : null}
    </>
  );
}

// Its own component so the effect is reached by rendering rather than by a hook
// call placed after the guards above: hook order stays the same on every
// render, and the guards keep deciding whether the trigger exists at all.
function AutoStartTrigger({
  tour,
  isPreview,
  showsKey,
}: {
  tour: Tour;
  isPreview: boolean;
  showsKey: string;
}) {
  useEffect(() => {
    if (isPreview) return;

    const startTour = () => {
      const shows = Number(localStorage.getItem(showsKey) || 0);
      localStorage.setItem(showsKey, String(shows + 1));
      localStorage.setItem("user_seen", "true");
    };

    if (tour.trigger_type === "delay") {
      // Was `clearTimeout(setTimeout(startTour, 2000))` inside the cleanup,
      // which created a timer and cancelled it in the same breath, so a delay
      // tour never actually started.
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
  }, [tour, isPreview, showsKey]);

  return null;
}
