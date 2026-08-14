import { cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useTourRunner } from "./useTourRunner";
import { sendOnboardingEvents } from "../Components/Onboarding/events";
import type { Tour } from "../types/tour";
import { SELECTOR_MISSING_TIMEOUT } from "../Utils/constants";

vi.mock("react-router-dom", () => ({
  useLocation: () => ({
    pathname: "/",
    search: "",
    state: null,
  }),
  useNavigate: () => vi.fn(),
}));

vi.mock("../Components/Onboarding/events", () => ({
  sendOnboardingEvents: vi.fn(),
}));

const TEST_APP_KEY = "test-key";
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
    const { result } = renderHook(() =>
      useTourRunner(tour, TEST_APP_KEY, false, onClose),
    );

    // Шаг пропускается только после того, как истечёт окно ожидания якоря:
    // до этого рантайм обязан ждать, иначе он проигрывает гонку с рендером
    // хост-страницы.
    await waitFor(
      () => {
        expect(result.current.step).toBe(1);
      },
      { timeout: SELECTOR_MISSING_TIMEOUT + 1000 },
    );

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
  }, SELECTOR_MISSING_TIMEOUT + 5000);

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
      useTourRunner(tourWithBrokenSelector, TEST_APP_KEY, false, onClose),
    );

    // Шаг пропускается только после того, как истечёт окно ожидания якоря:
    // до этого рантайм обязан ждать, иначе он проигрывает гонку с рендером
    // хост-страницы.
    await waitFor(
      () => {
        expect(result.current.step).toBe(1);
      },
      { timeout: SELECTOR_MISSING_TIMEOUT + 1000 },
    );

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
  }, SELECTOR_MISSING_TIMEOUT + 5000);

  it("передаёт appKey при отправке событий", async () => {
    const onClose = vi.fn();
    renderHook(() => useTourRunner(tour, "test-key", false, onClose));

    await waitFor(() => {
      expect(sendOnboardingEvents).toHaveBeenCalled();
    });

    expect(sendOnboardingEvents).toHaveBeenCalledWith(
      expect.objectContaining({
        appKey: "test-key",
      }),
      expect.any(Array),
    );
  });

  it("отправляет tour_started при запуске", async () => {
    const onClose = vi.fn();
    renderHook(() => useTourRunner(tour, "test-key", false, onClose));

    await waitFor(() => {
      const events = vi
        .mocked(sendOnboardingEvents)
        .mock.calls.flatMap((call) => call[1]);

      expect(events).toEqual(
        expect.arrayContaining([
          expect.objectContaining({
            type: "tour_started",
          }),
        ]),
      );
    });
  });

  it("сортирует подсказки по step независимо от порядка в массиве", async () => {
    const onClose = vi.fn();
    const reorderedTour: Tour = {
      ...tour,
      hints: [
        {
          id: "hint-3",
          step: 3,
          title: "Третий шаг",
          content: "Шаг 3",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
        {
          id: "hint-1",
          step: 1,
          title: "Первый шаг",
          content: "Шаг 1",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
        {
          id: "hint-2",
          step: 2,
          title: "Второй шаг",
          content: "Шаг 2",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
      ],
    };

    const { result } = renderHook(() =>
      useTourRunner(reorderedTour, TEST_APP_KEY, false, onClose),
    );

    expect(result.current.hints.map((hint) => hint.id)).toEqual([
      "hint-1",
      "hint-2",
      "hint-3",
    ]);
    expect(result.current.step).toBe(0);
    expect(result.current.hint?.id).toBe("hint-1");
  });

  it("правильно определяет текущий шаг по current_hint_id после перестановки", () => {
    const onClose = vi.fn();
    const reorderedTour: Tour = {
      ...tour,
      current_hint_id: "hint-2",

      hints: [
        {
          id: "hint-3",
          step: 3,
          title: "Третий шаг",
          content: "Шаг 3",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
        {
          id: "hint-2",
          step: 2,
          title: "Второй шаг",
          content: "Шаг 2",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
        {
          id: "hint-1",
          step: 1,
          title: "Первый шаг",
          content: "Шаг 1",
          placement: "center",
          spotlight: false,
          wait_for_selector: false,
        },
      ],
    };

    const { result } = renderHook(() =>
      useTourRunner(reorderedTour, TEST_APP_KEY, false, onClose),
    );

    expect(result.current.hints.map((hint) => hint.id)).toEqual([
      "hint-1",
      "hint-2",
      "hint-3",
    ]);
    expect(result.current.step).toBe(1);
    expect(result.current.hint?.id).toBe("hint-2");
  });
});

describe("useTourRunner: якорь появляется позже первого кадра", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(sendOnboardingEvents).mockResolvedValue({
      accepted: 1,
      duplicates: 0,
      rejected: 0,
    });
  });

  afterEach(() => {
    cleanup();
    document.body.innerHTML = "";
    vi.restoreAllMocks();
  });

  // Хост-страница рисует якорь после своей загрузки данных, а тур стартует
  // сразу по on_load. Ждать элемент — работа рантайма, а не повод считать
  // селектор сломанным.
  it("дожидается элемента, который появился через 800 мс, и показывает подсказку", async () => {
    const onClose = vi.fn();
    const lateTour: Tour = {
      ...tour,
      hints: [{ ...tour.hints[0], selector: "#late-anchor" }],
    };

    setTimeout(() => {
      const anchor = document.createElement("div");
      anchor.id = "late-anchor";
      document.body.appendChild(anchor);
    }, 800);

    const { result } = renderHook(() =>
      useTourRunner(lateTour, TEST_APP_KEY, false, onClose),
    );

    await waitFor(() => expect(result.current.element).not.toBeNull(), {
      timeout: 4000,
    });

    const sentEvents = vi
      .mocked(sendOnboardingEvents)
      .mock.calls.flatMap((call) => call[1]);

    expect(sentEvents.map((event) => event.type)).not.toContain(
      "selector_missing",
    );
    expect(sentEvents.map((event) => event.type)).not.toContain(
      "tour_completed",
    );
    expect(onClose).not.toHaveBeenCalled();
  }, SELECTOR_MISSING_TIMEOUT + 5000);

  // Тур, у которого рантайм пропустил все шаги, пользователю не показали
  // вообще. Записывать это как tour_completed нельзя: бэкенд ставит прогрессу
  // статус completed, и при show_once тур больше никогда не выдастся.
  it("не отмечает тур пройденным, если ни одной подсказки не показали", async () => {
    const onClose = vi.fn();
    const unreachableTour: Tour = {
      ...tour,
      hints: [{ ...tour.hints[0], selector: "#never-appears" }],
    };

    renderHook(() =>
      useTourRunner(unreachableTour, TEST_APP_KEY, false, onClose),
    );

    await waitFor(() => expect(onClose).toHaveBeenCalled(), {
      timeout: SELECTOR_MISSING_TIMEOUT + 2000,
    });

    const sentTypes = vi
      .mocked(sendOnboardingEvents)
      .mock.calls.flatMap((call) => call[1])
      .map((event) => event.type);

    expect(sentTypes).toContain("selector_missing");
    expect(sentTypes).not.toContain("tour_completed");
  }, SELECTOR_MISSING_TIMEOUT + 5000);
});
