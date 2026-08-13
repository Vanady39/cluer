import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useLoadTour } from "./useLoadTour";
import { useManualStart } from "./useManualStart";
import { resolveTour } from "../Components/Onboarding/client";
import type { TriggerConfig, TriggerType } from "../types/tour";

vi.mock("../Components/Onboarding/client", () => ({
  resolveTour: vi.fn(),
}));

const TEST_APP_KEY = "test-key";

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
    localStorage.clear();
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("on_load запускает onboarding автоматически", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("on_load"));

    const setIsOpen = vi.fn();
    renderHook(() => useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen));

    await waitFor(() => {
      expect(setIsOpen).toHaveBeenCalledWith(true);
    });
  });

  it("manual не запускается автоматически", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
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
      const loadedTour = useLoadTour(
        TEST_APP_KEY,
        false,
        null,
        false,
        setIsOpen,
      );

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

  it("manual запускается через window.startOnboarding()", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() => {
      const loadedTour = useLoadTour(
        TEST_APP_KEY,
        false,
        null,
        false,
        setIsOpen,
      );

      useManualStart(loadedTour.tour, false, setIsOpen);
      return loadedTour;
    });

    await waitFor(() => {
      expect(result.current.tour).not.toBeNull();
    });
    expect(typeof window.startOnboarding).toBe("function");

    act(() => {
      window.startOnboarding?.();
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("exit_intent запускается при уходе курсора вверх", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("exit_intent"));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
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
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
    );

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

  it("delay не запускается без delay_ms", async () => {
    vi.useFakeTimers();
    vi.mocked(resolveTour).mockResolvedValue(createTour("delay", {}));

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
    );

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(result.current.tour).not.toBeNull();
    setIsOpen.mockClear();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(10000);
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);
  });

  it("scroll_depth запускается только после достижения нужной глубины", async () => {
    vi.mocked(resolveTour).mockResolvedValue(
      createTour("scroll_depth", {
        scroll_depth: 75,
      }),
    );

    Object.defineProperty(window, "scrollY", {
      value: 0,
      writable: true,
      configurable: true,
    });

    Object.defineProperty(window, "innerHeight", {
      value: 500,
      writable: true,
      configurable: true,
    });

    Object.defineProperty(document.documentElement, "scrollHeight", {
      value: 2000,
      configurable: true,
    });

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
    );

    await waitFor(() => {
      expect(result.current.tour).not.toBeNull();
    });
    setIsOpen.mockClear();

    // 25%
    act(() => {
      window.dispatchEvent(new Event("scroll"));
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    // 75%
    act(() => {
      Object.defineProperty(window, "scrollY", {
        value: 1000,
        writable: true,
        configurable: true,
      });
      window.dispatchEvent(new Event("scroll"));
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("inactivity запускается после заданного времени бездействия", async () => {
    vi.useFakeTimers();
    vi.mocked(resolveTour).mockResolvedValue(
      createTour("inactivity", {
        inactivity_secs: 3,
      }),
    );

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
    );

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(result.current.tour).not.toBeNull();
    setIsOpen.mockClear();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2999);
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1);
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("inactivity сбрасывает таймер при активности пользователя", async () => {
    vi.useFakeTimers();
    vi.mocked(resolveTour).mockResolvedValue(
      createTour("inactivity", {
        inactivity_secs: 3,
      }),
    );

    const setIsOpen = vi.fn();
    const { result } = renderHook(() =>
      useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen),
    );

    await act(async () => {
      await Promise.resolve();
      await Promise.resolve();
      await Promise.resolve();
    });
    expect(result.current.tour).not.toBeNull();
    setIsOpen.mockClear();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    act(() => {
      window.dispatchEvent(new Event("mousemove"));
    });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });
    expect(setIsOpen).not.toHaveBeenCalledWith(true);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1000);
    });
    expect(setIsOpen).toHaveBeenCalledWith(true);
  });

  it("передаёт isNewUser true для нового пользователя", async () => {
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();
    renderHook(() => useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen));

    await waitFor(() => {
      expect(resolveTour).toHaveBeenCalled();
    });

    expect(resolveTour).toHaveBeenCalledWith(
      expect.objectContaining({
        props: {
          isNewUser: true,
        },
      }),
    );
  });

  it("передаёт isNewUser false для существующего пользователя", async () => {
    localStorage.setItem("onboarding_subject_id", "anon-test");
    vi.mocked(resolveTour).mockResolvedValue(createTour("manual"));

    const setIsOpen = vi.fn();
    renderHook(() => useLoadTour(TEST_APP_KEY, false, null, false, setIsOpen));

    await waitFor(() => {
      expect(resolveTour).toHaveBeenCalled();
    });

    expect(resolveTour).toHaveBeenCalledWith(
      expect.objectContaining({
        props: {
          isNewUser: false,
        },
      }),
    );
  });
});
