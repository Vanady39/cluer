import { getSessionId, getSubjectId } from "./storage";
import type { RuntimeEvent, SendEventConfig, EventBatchResult } from "../../types/events";


function buildUrl(apiUrl: string): string {
  return `${apiUrl.replace(/\/$/, "")}/events`;
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

class NonRetryableEventError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "NonRetryableEventError";
  }
}

function wait(ms: number) {
  return new Promise<void>((resolve) => {
    window.setTimeout(resolve, ms);
  });
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

  const retryDelays = [0, 500, 1500];

  let lastError: unknown;

  for (let attempt = 0; attempt < retryDelays.length; attempt++) {
    const delay = retryDelays[attempt];

    if (delay > 0) {
      await wait(delay);
    }

    try {
      const response = await fetch(buildUrl(config.apiUrl), {
        method: "POST",
        headers: buildHeaders(config.appKey),
        body: JSON.stringify(body),
        keepalive: true,
      });

      if (response.status >= 500) {
        throw new Error(
          `Events server error: ${response.status}`,
        );
      }

      if (response.status === 429) {
        throw new Error(
          "Events request rate limited: 429",
        );
      }

      if (!response.ok) {
        const errorText = await response.text();

        throw new NonRetryableEventError(
          `Events request failed: ${response.status} ${errorText}`,
        );
      }

      const result: EventBatchResult =
        await response.json();

      if (result.rejected > 0) {
        console.error(
          "[Onboarding] events rejected",
          {
            rejected: result.rejected,
            accepted: result.accepted,
            duplicates: result.duplicates,
            errors: result.errors,
          },
        );
      }

      return result;
    } catch (error) {
      lastError = error;

      if (error instanceof NonRetryableEventError) {
        throw error;
      }

      const isLastAttempt =
        attempt === retryDelays.length - 1;

      if (isLastAttempt) {
        throw error;
      }

      console.warn(
        `[Onboarding] events send failed, retry ${attempt + 1}/${retryDelays.length - 1}`,
        error,
      );
    }
  }

  throw lastError;
}

export function sendOnboardingEvent(
  config: SendEventConfig,
  event: RuntimeEvent,
) {
  return sendOnboardingEvents(config, [event]);
}