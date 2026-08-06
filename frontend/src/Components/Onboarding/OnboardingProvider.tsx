import { useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";

export function OnboardingProvider() {
  const [isOpen, setIsOpen] = useState(true);
  const params = new URLSearchParams(window.location.search);
  const tourId = params.get("tour");
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";

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

  const getTour = () => {
    const tours = JSON.parse(localStorage.getItem("tours") || "[]");

    if (tourId) {
      return tours.find((tour: any) => tour.id === tourId);
    }

    return tours.find(
      (tour: any) =>
        tour.status === "published" &&
        tour.target_path === window.location.pathname,
    );
  };

  const tour = getTour();
  const closeTour = () => {
    setIsOpen(false);

    if (!isPreview) {
      localStorage.setItem("onboarding_completed", "true");
    }
  };

  const completed = localStorage.getItem("onboarding_completed");

  if (!tour || !isOpen || (completed && !isPreview)) {
    return null;
  }

  return <TourRunner tour={tour} onClose={closeTour} />;
}
