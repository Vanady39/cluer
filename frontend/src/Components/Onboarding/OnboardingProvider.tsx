import { useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import { useLoadTour } from "../../Hooks/useLoadTour";
import { useOnboardingGoals } from "../../Hooks/useOnboardingGoals";
import { useManualStart } from "../../Hooks/useManualStart";

interface Props {
  appKey: string;
}

export function OnboardingProvider({ appKey } : Props) {
  const [isOpen, setIsOpen] = useState(false);
  const params = new URLSearchParams(window.location.search);
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";
  const previewTourId = params.get("tourId");
  const { tour } = useLoadTour(appKey, isPreview, previewTourId, isBuilder, setIsOpen);

  useOnboardingGoals(tour, appKey, isPreview, isBuilder);
  useManualStart(tour, isBuilder, setIsOpen);

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
      appKey={appKey}
      isPreview={isPreview}
      onClose={() => setIsOpen(false)}
    />
  );
}