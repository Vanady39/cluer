import { useEffect, useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import type { Tour } from "../../types/sdk";
import { resolveTour } from "./client";

const API_URL = "http://localhost:8080";
const APP_KEY = "pk_4e17b539-07c4-429a-9b30-12a34b2059f5";

export function OnboardingProvider() {
  const [tour, setTour] = useState<Tour | null>(null);
  const [isOpen, setIsOpen] = useState(false);

  const params = new URLSearchParams(window.location.search);

  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";

  useEffect(() => {
    if (isBuilder || isPreview) {
      return;
    }

    let cancelled = false;

    async function start() {
      try {
        const response = await resolveTour({
          apiUrl: API_URL,
          appKey: APP_KEY,

          props: {
            isNewUser: true,
          },
        });

        console.log(
          "[Onboarding] RESOLVE RESPONSE",
          response,
        );

        if (cancelled) {
          return;
        }

        if (!response) {
          setTour(null);
          setIsOpen(false);
          return;
        }

        const raw: any = response;
        const source = raw.tour ?? raw;

        const resolvedTour: Tour = {
          id:
            source.id ??
            raw.tour_id,

          title:
            source.title ??
            "",

          description:
            source.description ??
            "",

          target_path:
            source.target_path ??
            "/",

          priority:
            source.priority ??
            0,

          trigger_type:
            source.trigger_type ??
            "on_load",

          audience:
            source.audience ?? {
              show_once: false,
              max_shows: 0,
              only_new: false,
            },

          hints:
            source.hints ??
            raw.hints ??
            [],

          current_hint_id:
            raw.current_hint_id ??
            source.current_hint_id,

          version_id:
            raw.version_id ??
            raw.tour_version_id ??
            source.version_id,
        };

        console.log(
          "[Onboarding] NORMALIZED TOUR",
          resolvedTour,
        );

        setTour(resolvedTour);
        setIsOpen(true);
      } catch (error) {
        console.error(
          "[Onboarding] RESOLVE ERROR",
          error,
        );

        setTour(null);
        setIsOpen(false);
      }
    }

    start();

    return () => {
      cancelled = true;
    };
  }, [isBuilder, isPreview]);

  // =========================
  // BUILDER
  // =========================

  if (isBuilder) {
    return (
      <Builder
        onSelect={(selector) => {
          localStorage.setItem(
            "selected_element",
            selector,
          );

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

  // =========================
  // PREVIEW
  // =========================

  if (isPreview) {
    return null;
  }

  // =========================
  // RUNTIME
  // =========================

  if (!tour || !isOpen) {
    return null;
  }

  return (
    <TourRunner
      tour={tour}
      onClose={() => {
        setIsOpen(false);
      }}
    />
  );
}