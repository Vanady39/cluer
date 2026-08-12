import { describe, expect, it } from "vitest";
import { normalizeTour } from "./tourMapper";
import type { TourHint, TourResponseDto } from "../../types";

describe("normalizeTour", () => {
  it("берет настройки из draft, если он существует", () => {
    const data = {
      tour: {
        id: "tour-1",
        title: "Тестовый сценарий",
        description: "Описание",
        enabled: true,
        priority: 10,
      },

      draft: {
        id: "draft-1",
        trigger_type: "delay",
        target_path: "/draft-page",
        audience: {
          show_once: true,
          max_shows: 2,
          only_new: true,
        },
      },

      published: {
        id: "published-1",
        trigger_type: "on_load",
        target_path: "/published-page",
        audience: {
          show_once: false,
          max_shows: 5,
          only_new: false,
        },
      },
    } as TourResponseDto;

    const result = normalizeTour(data);

    expect(result.trigger_type).toBe("delay");
    expect(result.target_path).toBe("/draft-page");
    expect(result.audience).toEqual({
      show_once: true,
      max_shows: 2,
      only_new: true,
    });
  });

  it("берет настройки из published, если draft отсутствует", () => {
    const data = {
      tour: {
        id: "tour-1",
        title: "Тестовый сценарий",
        description: "Описание",
        enabled: true,
        priority: 10,
      },

      draft: null,
      published: {
        id: "published-1",
        trigger_type: "exit_intent",
        target_path: "/profile",
        audience: {
          show_once: false,
          max_shows: 3,
          only_new: false,
        },
      },
    } as TourResponseDto;

    const result = normalizeTour(data);

    expect(result.trigger_type).toBe("exit_intent");
    expect(result.target_path).toBe("/profile");
    expect(result.audience).toEqual({
      show_once: false,
      max_shows: 3,
      only_new: false,
    });
  });

  it("использует значения по умолчанию, если версий нет", () => {
    const data = {
      tour: {
        id: "tour-1",
        title: "Тестовый сценарий",
        description: "Описание",
        enabled: true,
        priority: 10,
      },

      draft: null,
      published: null,
    } as TourResponseDto;

    const result = normalizeTour(data);

    expect(result.trigger_type).toBe("on_load");
    expect(result.target_path).toBe("/");
    expect(result.audience).toEqual({
      show_once: true,
      max_shows: 1,
      only_new: false,
    });
  });

  it("передает hints в результат", () => {
    const data = {
      tour: {
        id: "tour-1",
        title: "Тестовый сценарий",
        description: "Описание",
        enabled: true,
        priority: 10,
      },

      draft: null,
      published: null,
    } as TourResponseDto;

    const hints = [
      {
        id: "hint-1",
        step: 1,
        title: "Первый шаг",
      },
    ] as TourHint[];

    const result = normalizeTour(data, hints);
    expect(result.hints).toEqual(hints);
  });
});
