export type EventType = "tour_started" | "hint_shown" | "hint_completed" | "hint_skipped" | "selector_missing"
                        | "tour_completed" | "tour_dismissed" | "goal_reached";

export interface RuntimeEvent {
  type: EventType;
  tourId: string;
  tourVersionId: string;
  hintId?: string | null;
  payload?: Record<string, unknown>;
}

export interface SendEventConfig {
  apiUrl: string;
  appKey: string;
  subjectId?: string;
}

export interface EventBatchResult {
  accepted: number;
  duplicates: number;
  rejected: number;
  errors?: string[];
}

export interface OnboardingGoalDetail {
  name: string;
  payload?: Record<string, unknown>;
}

export interface EventToSend {
  type: EventType;
  hintId?: string | null;
  payload?: Record<string, unknown>;
}