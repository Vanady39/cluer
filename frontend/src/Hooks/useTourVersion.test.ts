import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { onboardingAPI } from "../Api/onboarding";
import { useTourVersions } from "./useTourVersion";
import type { TourVersion } from "../types/tour";

vi.mock("../Api/onboarding", () => ({
  onboardingAPI: {
    getVersions: vi.fn(),
    rollbackVersion: vi.fn(),
  },
}));

const oldVersion = {
  id: "version-1",
  tour_id: "tour-1",
  version: 1,
  status: "archived",
} as TourVersion;

const newVersion = {
  id: "version-2",
  tour_id: "tour-1",
  version: 2,
  status: "published",
} as TourVersion;

describe("useTourVersions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(window, "alert").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("не делает rollback, если пользователь нажал Отмена", async () => {
    vi.mocked(onboardingAPI.getVersions).mockResolvedValue([oldVersion]);
    vi.spyOn(window, "confirm").mockReturnValue(false);

    const { result } = renderHook(() => useTourVersions());

    await act(async () => {
      await result.current.loadVersions("tour-1", "Тестовый сценарий");
    });

    await act(async () => {
      await result.current.rollback(oldVersion);
    });

    expect(onboardingAPI.rollbackVersion).not.toHaveBeenCalled();
  });

  it("делает rollback, если пользователь подтвердил действие", async () => {
    vi.mocked(onboardingAPI.getVersions)
      .mockResolvedValueOnce([oldVersion])
      .mockResolvedValueOnce([newVersion]);

    vi.mocked(onboardingAPI.rollbackVersion).mockResolvedValue(newVersion);
    vi.spyOn(window, "confirm").mockReturnValue(true);

    const { result } = renderHook(() => useTourVersions());

    await act(async () => {
      await result.current.loadVersions("tour-1", "Тестовый сценарий");
    });

    await act(async () => {
      await result.current.rollback(oldVersion);
    });

    expect(onboardingAPI.rollbackVersion).toHaveBeenCalledWith(
      "tour-1",
      "version-1",
    );

    // После rollback история версий загружается заново
    expect(onboardingAPI.getVersions).toHaveBeenCalledTimes(2);
    expect(result.current.versions).toEqual([newVersion]);
  });
});
