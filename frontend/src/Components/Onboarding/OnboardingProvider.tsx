import { useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import { useLoadTour } from "../../Hooks/useLoadTour";
import { useOnboardingGoals } from "../../Hooks/useOnboardingGoals";
import { useManualStart } from "../../Hooks/useManualStart";

export function OnboardingProvider() {
  const [isOpen, setIsOpen] = useState(false);
  const params = new URLSearchParams(window.location.search);
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";
  const previewTourId = params.get("tourId");
  const [hasSeenOnboarding, setHasSeenOnboarding] = useState(() => {
    return localStorage.getItem('onboarding_seen') === 'true';
  });
  const { tour, appKey } = useLoadTour(isPreview, previewTourId, isBuilder, setIsOpen, hasSeenOnboarding);

  useOnboardingGoals(tour, appKey, isPreview, isBuilder);
  useManualStart(tour, isBuilder, setIsOpen);

  const handleClose = () => {
    setIsOpen(false);
    localStorage.setItem('onboarding_seen', 'true');
    setHasSeenOnboarding(true);
  };

  if (isBuilder) {
    return (
      <Builder
        onSelect={(selector) => {
          localStorage.setItem("selected_element", selector);
          if (window.opener) {
            window.opener.postMessage(
              { type: "SELECTOR_SELECTED", selector },
              "*"
            );
            window.close();
          }
        }}
      />
    );
  }

  if (!tour || !isOpen) return null;

  return (
    <TourRunner
      tour={tour}
      isPreview={isPreview}
      onClose={handleClose}
    />
  );
}