import { useCallback, useEffect, useRef, useState } from "react";
import { TourRunner } from "./TourRunner";
import { Builder } from "./Builder";
import type { Tour } from "../../types/sdk";
import { resolveTour } from "./client";
import { sendOnboardingEvent } from "./events";
import { getOnboardingGoalEventName, type OnboardingGoalDetail } from "./goal";

const API_URL = "http://localhost:8080";
const APP_KEY = "pk_4e17b539-07c4-429a-9b30-12a34b2059f5";

type ResolvePayload = Partial<Tour> & {
  tour_id?: string;
  tour_version_id?: string;
  current_hint_id?: string;
  hints?: Tour["hints"];
  tour?: Partial<Tour>;
};

interface PreviewVersion {
  id: string;
  trigger_type?: Tour["trigger_type"];
  target_path?: string;
  audience?: Tour["audience"];
  hints?: Tour["hints"];
}

interface PreviewTourCard {
  tour: {
    id: string;
    title?: string;
    description?: string;
    priority?: number;
  };
  draft?: PreviewVersion | null;
  published?: PreviewVersion | null;
}

export function OnboardingProvider() {
  const [tour, setTour] = useState<Tour | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const sentGoalsRef = useRef<Set<string>>(new Set());
  const params = new URLSearchParams(window.location.search);
  const isPreview = params.get("preview") === "true";
  const isBuilder = params.get("builder") === "true";
  const previewTourId = params.get("tourId");

  const loadRuntimeTour = async (): Promise<Tour | null> => {
    const response = await resolveTour({
      apiUrl: API_URL,
      appKey: APP_KEY,
      props: { isNewUser: true },
    });

    if (!response) return null;

    const raw = response as ResolvePayload;
    const source = raw.tour ?? raw;
    const tourId = source.id ?? raw.tour_id;

    if (!tourId) throw new Error("Resolve response does not contain tour id");
    return {
      id: tourId,
      title: source.title ?? "",
      description: source.description ?? "",
      target_path: source.target_path ?? "/",
      priority: source.priority ?? 0,
      trigger_type: source.trigger_type ?? "on_load",
      audience: source.audience ?? {
        show_once: false,
        max_shows: 0,
        only_new: false,
      },
      hints: source.hints ?? raw.hints ?? [],
      current_hint_id: raw.current_hint_id ?? source.current_hint_id,
      version_id: raw.tour_version_id ?? source.version_id,
    };
  };

  const loadPreviewTour = async (
    tourId: string | null,
  ): Promise<Tour | null> => {
    if (!tourId) return null;

    const response = await fetch(`${API_URL}/v1/tours/${tourId}`);
    if (!response.ok)
      throw new Error(`Preview request failed: ${response.status}`);

    const card = (await response.json()) as PreviewTourCard;
    const source = card.draft ?? card.published;

    if (!source) {
      console.warn(
        "[Onboarding] PREVIEW: tour has no draft or published version",
      );
      return null;
    }

    const hints = [...(source.hints ?? [])].sort((a, b) => a.step - b.step);
    return {
      id: card.tour.id,
      title: card.tour.title ?? "",
      description: card.tour.description ?? "",
      target_path: source.target_path ?? "/",
      priority: card.tour.priority ?? 0,
      trigger_type: source.trigger_type ?? "on_load",
      audience: source.audience ?? {
        show_once: false,
        max_shows: 0,
        only_new: false,
      },
      hints,
      current_hint_id: hints[0]?.id,
      version_id: source.id,
    };
  };

  const loadTour = useCallback(
    async (isPreview: boolean, previewTourId: string | null) => {
      if (isPreview) return await loadPreviewTour(previewTourId);
      return await loadRuntimeTour();
    },
    [loadRuntimeTour, loadPreviewTour],
  );

  useEffect(() => {
    let cancelled = false;
    async function start() {
      try {
        if (isBuilder) return;
        const resolvedTour = await loadTour(isPreview, previewTourId);

        if (cancelled) return;

        if (!resolvedTour) {
          setTour(null);
          setIsOpen(false);
          return;
        }
        setTour(resolvedTour);
        setIsOpen(true);
      } catch (error) {
        console.error("[Onboarding] LOAD ERROR", error);
        if (!cancelled) {
          setTour(null);
          setIsOpen(false);
        }
      }
    }
    void start();
    return () => {
      cancelled = true;
    };
  }, [isBuilder, isPreview, previewTourId, loadTour]);

  useEffect(() => {
    if (isBuilder || isPreview) return;
    const handleGoal = (event: Event) => {
      if (!tour?.version_id) {
        console.warn("[Onboarding] goal ignored: active tour is missing");
        return;
      }

      const goal = (event as CustomEvent<OnboardingGoalDetail>).detail;
      if (!goal?.name) return;

      const key = `${tour.id}:${tour.version_id}:${goal.name}`;
      if (sentGoalsRef.current.has(key)) return;
      sentGoalsRef.current.add(key);

      void sendOnboardingEvent(
        { apiUrl: API_URL, appKey: APP_KEY },
        {
          type: "goal_reached",
          tourId: tour.id,
          tourVersionId: tour.version_id,
          hintId: null,
          payload: { goal: goal.name, ...goal.payload },
        },
      ).catch((error) => {
        sentGoalsRef.current.delete(key);
        console.error("[Onboarding] goal event failed", error);
      });
    };

    const eventName = getOnboardingGoalEventName();
    window.addEventListener(eventName, handleGoal);
    return () => window.removeEventListener(eventName, handleGoal);
  }, [isBuilder, isPreview, tour?.id, tour?.version_id]);

  if (isBuilder) {
    return (
      <Builder
        onSelect={(selector) => {
          localStorage.setItem("selected_element", selector);
          if (window.opener) {
            window.opener.postMessage(
              { type: "SELECTOR_SELECTED", selector },
              "*",
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
      onClose={() => setIsOpen(false)}
    />
  );
}