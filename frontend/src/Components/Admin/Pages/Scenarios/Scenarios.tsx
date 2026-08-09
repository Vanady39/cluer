import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";

import styles from "./Styles.module.scss";

import { Button } from "../../../UI/Button";

import { useToursQuery } from "../../../../Hooks/useToursQuery";
import { useDeleteTourMutation } from "../../../../Hooks/useDeleteTourMutation";

import { onboardingAPI } from "../../../../Api/onboarding";

import type { TourVersion } from "../../../../types/sdk";

import { ScenarioMenu } from "../ScenarioMenu/ScenarioMenu";
import logo from "/logo.svg";

function ScenariosComponent() {
  const navigate = useNavigate();

  const [openMenu, setOpenMenu] = useState<string | null>(null);

  const [updatingEnabledId, setUpdatingEnabledId] = useState<string | null>(
    null,
  );

  const [versionsTour, setVersionsTour] = useState<{
    id: string;
    title: string;
  } | null>(null);

  const [versions, setVersions] = useState<TourVersion[]>([]);

  const [versionsLoading, setVersionsLoading] = useState(false);

  const [rollbackVersionId, setRollbackVersionId] = useState<string | null>(
    null,
  );

  const { data: scenarios = [], isLoading, error, refetch } = useToursQuery();

  const deleteMutation = useDeleteTourMutation();

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
    setOpenMenu(null);
  };

  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    try {
      setUpdatingEnabledId(id);

      await onboardingAPI.updateTourMeta(id, {
        enabled,
      });

      await refetch();
    } catch (error) {
      console.error("TOGGLE ENABLED ERROR", error);

      alert("Не удалось изменить состояние сценария");
    } finally {
      setUpdatingEnabledId(null);
    }
  };

  const handleEdit = async (id: string) => {
    try {
      const tour = await onboardingAPI.getTour(id);

      // Published напрямую не редактируем.
      // Если draft отсутствует — создаём fork.
      if (tour.published && !tour.draft) {
        await onboardingAPI.createDraft(id);
      }

      navigate(`/admin/scenarios/create?id=${id}`);
    } catch (error) {
      console.error("EDIT TOUR ERROR", error);

      alert("Не удалось открыть сценарий");
    }
  };

  function getTourStatus(tour: any) {
    const hasDraft = Boolean(tour.draft);

    const hasPublished = Boolean(tour.published);

    if (hasPublished && hasDraft) {
      return {
        label: "Есть неопубликованные изменения",
        type: "changes",
      };
    }

    if (hasPublished) {
      return {
        label: "Опубликован",
        type: "published",
      };
    }

    if (hasDraft) {
      return {
        label: "Черновик",
        type: "draft",
      };
    }

    return {
      label: "Без версии",
      type: "unknown",
    };
  }

  const handlePreview = async (tourId: string) => {
    try {
      const card = await onboardingAPI.getTour(tourId);

      const version = card.draft ?? card.published;

      if (!version) {
        alert("У сценария пока нет версии для предпросмотра");

        return;
      }

      const path = version.target_path || "/";

      const previewUrl = new URL(path, window.location.origin);

      previewUrl.searchParams.set("preview", "true");

      previewUrl.searchParams.set("tourId", tourId);

      window.open(previewUrl.toString(), "_blank", "noopener,noreferrer");
    } catch (error) {
      console.error("[Admin] preview error", error);

      alert("Не удалось открыть предпросмотр");
    }
  };

  const handleOpenVersions = async (tourId: string, title: string) => {
    setOpenMenu(null);

    setVersionsTour({
      id: tourId,
      title,
    });

    setVersions([]);
    setVersionsLoading(true);

    try {
      const result = await onboardingAPI.getVersions(tourId);

      setVersions(result);
    } catch (error) {
      console.error("LOAD VERSIONS ERROR", error);

      alert("Не удалось загрузить историю версий");

      setVersionsTour(null);
    } finally {
      setVersionsLoading(false);
    }
  };

  const handleCloseVersions = () => {
    if (rollbackVersionId) {
      return;
    }

    setVersionsTour(null);
    setVersions([]);
  };

  const handleRollback = async (version: TourVersion) => {
    if (!versionsTour) {
      return;
    }

    const confirmed = window.confirm(
      `Откатить сценарий «${versionsTour.title}» к версии v${version.version}?\n\n` +
        "Текущая опубликованная версия будет архивирована, " +
        `а содержимое v${version.version} будет опубликовано как новая версия.`,
    );

    if (!confirmed) {
      return;
    }

    try {
      setRollbackVersionId(version.id);

      const published = await onboardingAPI.rollbackVersion(
        versionsTour.id,
        version.id,
      );

      const updatedVersions = await onboardingAPI.getVersions(versionsTour.id);

      setVersions(updatedVersions);

      await refetch();

      alert(`Откат выполнен. Опубликована версия v${published.version}.`);
    } catch (error) {
      console.error("ROLLBACK ERROR", error);

      const status = (
        error as {
          response?: {
            status?: number;
          };
        }
      ).response?.status;

      if (status === 409) {
        alert(
          "Откат невозможен: у сценария есть неопубликованный черновик. " +
            "Сначала опубликуйте или удалите изменения.",
        );
      } else {
        alert("Не удалось выполнить откат версии");
      }
    } finally {
      setRollbackVersionId(null);
    }
  };

  const formatVersionDate = (value?: string | null) => {
    if (!value) {
      return "—";
    }

    return new Date(value).toLocaleString("ru-RU");
  };

  const getVersionStatusLabel = (status: TourVersion["status"]) => {
    switch (status) {
      case "published":
        return "Опубликована";

      case "draft":
        return "Черновик";

      case "archived":
        return "Архивная";

      default:
        return status;
    }
  };

  if (isLoading) {
    return <div>Загрузка сценариев...</div>;
  }

  if (error) {
    return <div>Ошибка загрузки сценариев: {error.message}</div>;
  }

  return (
    <>
      <div className={styles.page}>
        <ScenarioMenu />

        <div className={styles.header}>
          <img src={logo} className={styles.logo}/>
          <h1>Сценарии</h1>
          <Button
            className={styles.createButton}
            size="main"
            color="primary"
            onClick={() => navigate("/admin/scenarios/create")}
          >
            + Создать
          </Button>
        </div>

        <div className={styles.table}>
          <div className={styles.rowHeader}>
            <span>Название</span>
            <span>Статус</span>
            <span>Шагов</span>
            <span>Обновлен</span>
            <span>Действия</span>
          </div>

          {scenarios.map((item) => {
            const status = getTourStatus(item);

            return (
              <div className={styles.row} key={item.id}>
                <span className={styles.name}>{item.title}</span>

                <span>
                  <span
                    className={`
                      ${styles.status}
                      ${
                        status.type === "published"
                          ? styles.success
                          : styles.warning
                      }
                    `}
                  />

                  {status.label}

                  <span>{item.enabled ? " · включен" : " · выключен"}</span>

                  <input
                    type="checkbox"
                    checked={Boolean(item.enabled)}
                    disabled={updatingEnabledId === item.id}
                    onChange={(event) =>
                      handleToggleEnabled(item.id, event.target.checked)
                    }
                  />
                </span>

                <span>{item.hints?.length || 0}</span>

                <span>
                  {item.updated_at
                    ? new Date(item.updated_at).toLocaleDateString()
                    : "—"}
                </span>

                <div className={styles.actions}>
                  <Button size="min" onClick={() => handlePreview(item.id)}>
                    Предпросмотр
                  </Button>

                  <div className={styles.menuWrapper}>
                    <Button
                      size="min"
                      className={styles.more}
                      onClick={() =>
                        setOpenMenu(openMenu === item.id ? null : item.id)
                      }
                    >
                      ⋮
                    </Button>

                    {openMenu === item.id && (
                      <div className={styles.dropdown}>
                        <button
                          type="button"
                          className={styles.menuButton}
                          onClick={() =>
                            handleOpenVersions(item.id, item.title)
                          }
                        >
                          История версий
                        </button>

                        <button
                          type="button"
                          className={styles.menuButton}
                          onClick={() => handleEdit(item.id)}
                        >
                          Редактировать
                        </button>

                        <button
                          type="button"
                          className={styles.deleteButton}
                          onClick={() => handleDelete(item.id)}
                        >
                          Удалить
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {versionsTour && (
        <div
          className={styles.modalOverlay}
          onMouseDown={(event) => {
            if (event.target === event.currentTarget) {
              handleCloseVersions();
            }
          }}
        >
          <div className={styles.versionsModal}>
            <div className={styles.modalHeader}>
              <div>
                <h2>История версий</h2>

                <p>{versionsTour.title}</p>
              </div>

              <button
                type="button"
                className={styles.closeButton}
                disabled={Boolean(rollbackVersionId)}
                onClick={handleCloseVersions}
              >
                ×
              </button>
            </div>

            {versionsLoading ? (
              <div className={styles.versionsLoading}>Загрузка истории...</div>
            ) : versions.length === 0 ? (
              <div className={styles.versionsEmpty}>Версий пока нет</div>
            ) : (
              <div className={styles.versionsList}>
                {versions.map((version) => {
                  const isPublished = version.status === "published";

                  const isDraft = version.status === "draft";

                  const isRollingBack = rollbackVersionId === version.id;

                  return (
                    <div key={version.id} className={styles.versionRow}>
                      <div className={styles.versionInfo}>
                        <div className={styles.versionTitle}>
                          <strong>v{version.version}</strong>

                          <span
                            className={`
                                ${styles.versionStatus}
                                ${
                                  isPublished
                                    ? styles.versionPublished
                                    : isDraft
                                      ? styles.versionDraft
                                      : styles.versionArchived
                                }
                              `}
                          >
                            {getVersionStatusLabel(version.status)}
                          </span>
                        </div>

                        <span className={styles.versionDate}>
                          {formatVersionDate(
                            version.published_at ?? version.created_at,
                          )}
                        </span>

                        <code className={styles.versionId}>{version.id}</code>
                      </div>

                      {version.status === "archived" && (
                        <button
                          type="button"
                          className={styles.rollbackButton}
                          disabled={Boolean(rollbackVersionId)}
                          onClick={() => handleRollback(version)}
                        >
                          {isRollingBack ? "Откат..." : "Откатить"}
                        </button>
                      )}

                      {isPublished && (
                        <span className={styles.currentVersion}>Текущая</span>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}

export const Scenarios = memo(ScenariosComponent);
