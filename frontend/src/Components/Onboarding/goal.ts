import type { OnboardingGoalDetail } from "../../types/events";

const GOAL_EVENT = "onboarding:goal-reached";
export function trackOnboardingGoal(name: string, payload?: Record<string, unknown>,
) {
  window.dispatchEvent(
    new CustomEvent<OnboardingGoalDetail>(GOAL_EVENT, {
      detail: {
        name,
        payload,
      },
    }),
  );
}

export function getOnboardingGoalEventName() {
  return GOAL_EVENT;
}
