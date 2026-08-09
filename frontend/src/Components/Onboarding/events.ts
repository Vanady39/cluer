import { getSessionId, getSubjectId } from "./storage";

export type EventType =
  | "tour_started"
  | "hint_shown"
  | "hint_completed"
  | "hint_skipped"
  | "selector_missing"
  | "tour_completed"
  | "tour_dismissed"
  | "goal_reached";

export interface RuntimeEvent {
  type: EventType;
  tourId: string;
  tourVersionId: string;
  hintId?: string | null;
  payload?: Record<string, unknown>;
}

interface SendEventConfig {
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

export async function sendOnboardingEvents(
  config: SendEventConfig,
  events: RuntimeEvent[],
): Promise<EventBatchResult> {
  const body = {
    subject_id: getSubjectId(config.subjectId),
    session_id: getSessionId(),

    events: events.map((event) => ({
      event_key: crypto.randomUUID(),
      type: event.type,
      tour_id: event.tourId,
      tour_version_id: event.tourVersionId,
      hint_id: event.hintId ?? null,
      occurred_at: new Date().toISOString(),
      payload: event.payload ?? {},
    })),
  };

  console.log(
    "[Onboarding SDK] EVENT REQUEST",
    body,
  );

  const response = await fetch(
    `${config.apiUrl}/v1/events`,
    {
      method: "POST",

      headers: {
        "Content-Type": "application/json",
        "X-App-Key": config.appKey,
      },

      body: JSON.stringify(body),

      // Важно, если клик по элементу
      // сразу уводит пользователя на другую страницу.
      keepalive: true,
    },
  );

  if (!response.ok) {
    const text = await response.text();

    throw new Error(
      `Events request failed: ${response.status} ${text}`,
    );
  }

  const result: EventBatchResult =
    await response.json();

  console.log(
    "[Onboarding SDK] EVENT RESPONSE",
    result,
  );

  return result;
}

export function sendOnboardingEvent(
  config: SendEventConfig,
  event: RuntimeEvent,
) {
  return sendOnboardingEvents(config, [event]);
}