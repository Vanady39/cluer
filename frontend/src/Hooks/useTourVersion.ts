import { useState } from "react";
import { onboardingAPI } from "../Api/onboarding";
import type { TourVersion } from "../types/tour";

export function useTourVersions() {
  const [versionsTour, setVersionsTour] = useState<{ id: string; title: string } | null>(null);
  const [versions, setVersions] = useState<TourVersion[]>([]);
  const [versionsLoading, setVersionsLoading] = useState(false);
  const [rollbackVersionId, setRollbackVersionId] = useState<string | null>(null);

  const loadVersions = async (tourId: string, title: string) => {
    setVersionsTour({ id: tourId, title });
    setVersions([]);
    setVersionsLoading(true);

    try {
      const data = await onboardingAPI.getVersions(tourId);
      setVersions(data);
    } catch (error) {
      console.error("Ошибка загрузки версий:", error);
      alert("Не удалось загрузить историю версий");
      setVersionsTour(null);
    } finally {
      setVersionsLoading(false);
    }
  };

  const closeVersions = () => {
    if (rollbackVersionId) return;
    setVersionsTour(null);
    setVersions([]);
  };

  const rollback = async (version: TourVersion): Promise<boolean> => {
    if (!versionsTour) return false;

    const confirmed = window.confirm(
      `Откатить сценарий «${versionsTour.title}» к версии v${version.version}?\n\n` +
        "Текущая опубликованная версия будет архивирована, " +
        `а содержимое v${version.version} будет опубликовано как новая версия.`,
    );
    if (!confirmed) return false;

    try {
      setRollbackVersionId(version.id);
      const published = await onboardingAPI.rollbackVersion(versionsTour.id, version.id);
      setVersions(await onboardingAPI.getVersions(versionsTour.id));
      alert(`Откат выполнен. Опубликована версия v${published.version}.`);
      return true;
    } catch (error) {
      const status = (error as { response?: { status?: number } }).response?.status;
      if (status === 409) {
        alert(
          "Откат невозможен: у сценария есть неопубликованный черновик. " +
            "Сначала опубликуйте или удалите изменения.",
        );
      } else {
        alert("Не удалось выполнить откат версии");
      }
      return false;
    } finally {
      setRollbackVersionId(null);
    }
  };

  return {
    versionsTour,
    versions,
    versionsLoading,
    rollbackVersionId,
    loadVersions,
    closeVersions,
    rollback,
  };
}