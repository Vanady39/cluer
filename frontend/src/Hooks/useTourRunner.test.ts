import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTourRunner } from "./useTourRunner";
import { getCurrentApp } from "../Api/Helpers/Helpers";
import { sendOnboardingEvents } from "../Components/Onboarding/events";
import type { Tour } from "../types/tour";

vi.mock("react-router-dom", () => ({
  useLocation: () => ({
    pathname: "/",
    search: "",
    state: null,
  }),

  useNavigate: () => vi.fn(),
}));

vi.mock("../Api/Helpers/Helpers", () => ({
  getCurrentApp: vi.fn(),
}));

vi.mock("../Components/Onboarding/events", () => ({
  sendOnboardingEvents: vi.fn(),
}));

const tour: Tour = {
  id: "tour-1",
  title: "Тестовый сценарий",
  description: "",
  target_path: "/",
  priority: 0,
  trigger_type: "on_load",
  version_id: "version-1",

  audience: {
    show_once: false,
    max_shows: 0,
    only_new: false,
  },

  hints: [
    {
      id: "hint-1",
      step: 1,
      title: "Первый шаг",
      content: "Подсказка",
      selector: "#missing-element",
      placement: "top",
      spotlight: false,
      wait_for_selector: false,
    },
    {
      id: "hint-2",
      step: 2,
      title: "Второй шаг",
      content: "Следующая подсказка",
      placement: "center",
      spotlight: false,
      wait_for_selector: false,
    },
  ],
};

describe("useTourRunner", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getCurrentApp).mockResolvedValue({
      id: "app-1",
      name: "Test app",
      public_key: "test-key",
      allowed_origins: [],
    });
    vi.mocked(sendOnboardingEvents).mockResolvedValue({
      accepted: 1,
      duplicates: 0,
      rejected: 0,
    });
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("пропускает шаг, если DOM-элемент не найден", async () => {
    const onClose = vi.fn();

    const { result } = renderHook(() => useTourRunner(tour, false, onClose));
    await waitFor(() => {
      expect(result.current.step).toBe(1);
    });

    const sentEvents = vi
      .mocked(sendOnboardingEvents)
      .mock.calls.flatMap((call) => call[1]);

    expect(sentEvents).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          type: "selector_missing",
          hintId: "hint-1",
        }),
      ]),
    );

    expect(sentEvents).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          type: "hint_skipped",
          hintId: "hint-1",
        }),
      ]),
    );

    // Один проблемный шаг не должен закрывать весь onboarding
    expect(onClose).not.toHaveBeenCalled();
    expect(result.current.hint?.id).toBe("hint-2");
  });

  it("продолжает onboarding при некорректном селекторе", async () => {
    const onClose = vi.fn();

    const tourWithBrokenSelector: Tour = {
      ...tour,
      hints: [
        {
          ...tour.hints[0],
          selector: "[broken",
        },
        tour.hints[1],
      ],
    };

    const { result } = renderHook(() =>
      useTourRunner(tourWithBrokenSelector, false, onClose),
    );

    await waitFor(() => {
      expect(result.current.step).toBe(1);
    });

    const sentEvents = vi
      .mocked(sendOnboardingEvents)
      .mock.calls.flatMap((call) => call[1]);

    expect(sentEvents).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          type: "selector_missing",
          hintId: "hint-1",
        }),
      ]),
    );

    expect(sentEvents).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          type: "hint_skipped",
          hintId: "hint-1",
        }),
      ]),
    );

    expect(onClose).not.toHaveBeenCalled();
    expect(result.current.hint?.id).toBe("hint-2");
  });
});
