import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button";
import logo from "/logo.svg";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import { useDeleteTourMutation } from "../../../../Hooks/useDeleteTourMutation";
import { onboardingAPI } from "../../../../Api/onboarding";

function ScenariosComponent() {
  const navigate = useNavigate();

  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [updatingEnabledId, setUpdatingEnabledId] = useState<string | null>(
    null,
  );

  const {
    data: scenarios = [],
    isLoading,
    error,
    refetch,
  } = useToursQuery();

  const deleteMutation = useDeleteTourMutation();

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
    setOpenMenu(null);
  };

  const handleToggleEnabled = async (
    id: string,
    enabled: boolean,
  ) => {
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

  const handleEdit = async (id: string, status: string) => {
    try {
      const tour = await onboardingAPI.getTour(id);

      // Есть published, но draft ещё нет
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

  if (isLoading) {
    return <div>Загрузка сценариев...</div>;
  }

  if (error) {
    return (
      <div>
        Ошибка загрузки сценариев: {error.message}
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
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
              <span className={styles.name}>
                {item.title}
              </span>

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

                <span>
                  {item.enabled
                    ? " · включен"
                    : " · выключен"}
                </span>

                <input
                  type="checkbox"
                  checked={Boolean(item.enabled)}
                  disabled={updatingEnabledId === item.id}
                  onChange={(event) =>
                    handleToggleEnabled(
                      item.id,
                      event.target.checked,
                    )
                  }
                />
              </span>

              <span>{item.hints?.length || 0}</span>

              <span>
                {item.updated_at
                  ? new Date(
                      item.updated_at,
                    ).toLocaleDateString()
                  : "—"}
              </span>

              <div className={styles.actions}>
                <Button
                  size="min"
                  onClick={() => {
                    window.open(
                      `${
                        item.hints[0]?.page_path || "/"
                      }?tour=${item.id}&preview=true`,
                      "_blank",
                    );
                  }}
                >
                  Предпросмотр
                </Button>

                <div className={styles.menuWrapper}>
                  <Button
                    size="min"
                    className={styles.more}
                    onClick={() =>
                      setOpenMenu(
                        openMenu === item.id
                          ? null
                          : item.id,
                      )
                    }
                  >
                    ⋮
                  </Button>

                  {openMenu === item.id && (
                    <div className={styles.dropdown}>
                      <button
                        className={styles.deleteButton}
                        onClick={() =>
                          handleDelete(item.id)
                        }
                      >
                        Удалить
                      </button>

                      <Button
                        size="min"
                        onClick={() =>
                          handleEdit(
                            item.id,
                            item.status,
                          )
                        }
                      >
                        Редактировать
                      </Button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export const Scenarios = memo(ScenariosComponent);