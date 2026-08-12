import { describe, expect, it } from "vitest";
import type { TourVersion } from "../types/tour";
import { getTourStatus, getVersionStatusLabel } from "./status";

const createVersion = (status: TourVersion["status"]): TourVersion => ({
  id: "test-id",
  tour_id: "tour-123",
  version: 1,
  status,
  trigger_type: "on_load",
  target_path: "/",
  audience: {
    show_once: false,
    max_shows: 0,
    only_new: false,
  },
  created_at: "2026-08-12T00:00:00Z",
});

describe("getVersionStatusLabel", () => {
  it("возвращает Черновик для draft", () => {
    expect(getVersionStatusLabel("draft")).toBe("Черновик");
  });

  it("возвращает Опубликована для published", () => {
    expect(getVersionStatusLabel("published")).toBe("Опубликована");
  });

  it("возвращает Архивная для archived", () => {
    expect(getVersionStatusLabel("archived")).toBe("Архивная");
  });
});

describe("getTourStatus", () => {
  it("возвращает неопубликованные изменения, если есть draft и published", () => {
    const tour = {
      draft: createVersion("draft"),
      published: createVersion("published"),
    };

    expect(getTourStatus(tour)).toEqual({
      type: "changes",
      label: "Есть неопубликованные изменения",
    });
  });

  it("возвращает Опубликован, если есть только published", () => {
    const tour = {
      published: createVersion("published"),
    };

    expect(getTourStatus(tour)).toEqual({
      type: "published",
      label: "Опубликован",
    });
  });

  it("возвращает Черновик, если есть только draft", () => {
    const tour = {
      draft: createVersion("draft"),
    };

    expect(getTourStatus(tour)).toEqual({
      type: "draft",
      label: "Черновик",
    });
  });

  it("возвращает Без версии, если версий нет", () => {
    const tour = {};

    expect(getTourStatus(tour)).toEqual({
      type: "unknown",
      label: "Без версии",
    });
  });
});
