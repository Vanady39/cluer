import { useEffect, useRef } from "react";
import { sendOnboardingEvent } from "../Components/Onboarding/events";
import { getOnboardingGoalEventName } from "../Components/Onboarding/goal";
import type { OnboardingGoalDetail, Tour } from "../types";
import { API_URL } from "../Config/env";

export function useOnboardingGoals(
  tour: Tour | null,
  appKey: string | null,
  isPreview: boolean,
  isBuilder: boolean,
) {
  const sentGoalsRef = useRef<Set<string>>(new Set());

  useEffect(() => {
    if (isBuilder || isPreview) return;
    const handleGoal = (event: Event) => {
      if (!tour?.version_id) {
        console.warn("[Onboarding] goal ignored: active tour is missing");
        return;
      }

      if (!appKey) {
        console.warn("[Onboarding] goal ignored: app key is missing");
        return;
      }

      const goal = (event as CustomEvent<OnboardingGoalDetail>).detail;
      if (!goal?.name) return;

      const key = `${tour.id}:${tour.version_id}:${goal.name}`;
      if (sentGoalsRef.current.has(key)) return;
      sentGoalsRef.current.add(key);

      void sendOnboardingEvent(
        { apiUrl: API_URL, appKey },
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
  }, [appKey, isBuilder, isPreview, tour?.id, tour?.version_id]);
}