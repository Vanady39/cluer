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

function buildUrl(apiUrl: string): string {
  return `${apiUrl.replace(/\/$/, "")}/v1/events`;
}

function buildHeaders(appKey: string): HeadersInit {
  return {
    "Content-Type": "application/json",
    "X-App-Key": appKey,
  };
}

function buildEventPayload(event: RuntimeEvent) {
  return {
    event_key: crypto.randomUUID(),
    type: event.type,
    tour_id: event.tourId,
    tour_version_id: event.tourVersionId,
    hint_id: event.hintId ?? null,
    occurred_at: new Date().toISOString(),
    payload: event.payload ?? {},
  };
}

export async function sendOnboardingEvents(
  config: SendEventConfig,
  events: RuntimeEvent[],
): Promise<EventBatchResult> {
  const body = {
    subject_id: getSubjectId(config.subjectId),
    session_id: getSessionId(),
    events: events.map(buildEventPayload),
  };

  const response = await fetch(buildUrl(config.apiUrl), {
    method: "POST",
    headers: buildHeaders(config.appKey),
    body: JSON.stringify(body),
    keepalive: true,
  });

  if (!response.ok) throw new Error(`Events request failed: ${response.status} ${await response.text()}`);

  const result: EventBatchResult = await response.json();
  return result;
}

export function sendOnboardingEvent(
  config: SendEventConfig,
  event: RuntimeEvent,
) {
  return sendOnboardingEvents(config, [event]);
}