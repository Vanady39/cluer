import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useLoadTour } from "./useLoadTour";
import { useManualStart } from "./useManualStart";
import { resolveTour } from "../Components/Onboarding/client";
import { getCurrentApp } from "../Api/Helpers/Helpers";
import type { TriggerConfig, TriggerType } from "../types/tour";

vi.mock("../Components/Onboarding/client", () => ({
  resolveTour: vi.fn(),
}));

vi.mock("../Api/Helpers/Helpers", () => ({
  getCurrentApp: vi.fn(),
}));

function createTour(triggerType: TriggerType, triggerConfig?: TriggerConfig) {
  return {
    id: "tour-1",
    title: "Тестовый сценарий",
    description: "",
    target_path: "/",
    priority: 0,
    trigger_type: triggerType,
    trigger_config: triggerConfig,

    audience: {
      show_once: false,
      max_shows: 0,
      only_new: false,
    },

    hints: [],
    version_id: "version-1",
  };
}

describe("триггеры onboarding", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getCurrentApp).mockResolvedValue({
      id: "app-1",
      name: "Test app",
      public_key: "test-key",
      allowed_origins: [],
    });
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("on_load запускает onboarding автоматически", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("on_load"));

    const setIsOpen = vi.fn();
    renderHook(() => useLoadTour(false, null, false, setIsOpen));

    await waitFor(() => {
      expect(setIsOpen).toHaveBeenCalledWith(true);
    });
  });

  it("manual не запускается автоматически", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(false, null, false, setIsOpen),
    );

    await waitFor(() => {
      expect(result.current.tour).not.toBeNull();
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);
  });

  it("manual запускается после события start-onboarding", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();

    const { result } = renderHook(() => {
      const loadedTour = useLoadTour(false, null, false, setIsOpen);
      useManualStart(loadedTour.tour, false, setIsOpen);
      return loadedTour;
    });

    await waitFor(() => {
      expect(result.current.tour).not.toBeNull();
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    act(() => {
      window.dispatchEvent(new Event("start-onboarding"));
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("exit_intent запускается при уходе курсора вверх", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("exit_intent"));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(false, null, false, setIsOpen),
    );

    await waitFor(() => {
      expect(result.current.tour).not.toBeNull();
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    act(() => {
      document.dispatchEvent(
        new MouseEvent("mouseleave", {
          clientY: -1,
        }),
      );
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("delay запускает onboarding через заданное время", async () => {
    vi.useFakeTimers();
    vi.mocked(resolveTour).mockResolvedValue(
      createTour("delay", {
        delay_ms: 1000,
      }),
    );

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(false, null, false, setIsOpen),
    );

    // Завершаем асинхронную загрузку тура, не двигая виртуальный таймер
    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(result.current.tour).not.toBeNull();
    setIsOpen.mockClear();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(999);
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });
});
