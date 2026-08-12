import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { sendOnboardingEvents } from "./events";
import type { EventBatchResult, RuntimeEvent } from "../../types/events";

const config = {
  apiUrl: "/v1",
  appKey: "test-key",
  subjectId: "user-1",
};

const event: RuntimeEvent = {
  type: "tour_started",
  tourId: "tour-1",
  tourVersionId: "version-1",
};

const successResult: EventBatchResult = {
  accepted: 1,
  duplicates: 0,
  rejected: 0,
};

describe("sendOnboardingEvents", () => {
  beforeEach(() => {
    // Retry на 500 и 429 ожидаем, поэтому не выводим warning в тестах
    vi.spyOn(console, "warn").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("повторяет запрос после ошибки 500", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response("", {
          status: 500,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(successResult), {
          status: 200,
        }),
      );

    vi.stubGlobal("fetch", fetchMock);

    const result = await sendOnboardingEvents(config, [event]);

    expect(result).toEqual(successResult);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("повторяет запрос после ошибки 429", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response("", {
          status: 429,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(successResult), {
          status: 200,
        }),
      );

    vi.stubGlobal("fetch", fetchMock);

    const result = await sendOnboardingEvents(config, [event]);

    expect(result).toEqual(successResult);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("не повторяет запрос после ошибки 400", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response("Bad request", {
        status: 400,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);
    await expect(sendOnboardingEvents(config, [event])).rejects.toThrow();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("сохраняет одинаковый event_key при retry", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response("", {
          status: 500,
        }),
      )
      .mockResolvedValueOnce(
        new Response(JSON.stringify(successResult), {
          status: 200,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    await sendOnboardingEvents(config, [event]);
    const firstRequest = fetchMock.mock.calls[0][1] as RequestInit;
    const secondRequest = fetchMock.mock.calls[1][1] as RequestInit;
    const firstBody = JSON.parse(String(firstRequest.body));
    const secondBody = JSON.parse(String(secondRequest.body));
    const firstKey = firstBody.events[0].event_key;
    const secondKey = secondBody.events[0].event_key;

    expect(firstKey).toBeDefined();
    expect(secondKey).toBe(firstKey);
  });

  it("сообщает об отклоненном событии в консоль", async () => {
    const rejectedResult: EventBatchResult = {
      accepted: 0,
      duplicates: 0,
      rejected: 1,
      errors: ["invalid event"],
    };

    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(rejectedResult), {
        status: 200,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const result = await sendOnboardingEvents(config, [event]);

    expect(result).toEqual(rejectedResult);
    expect(consoleError).toHaveBeenCalled();
  });
});
