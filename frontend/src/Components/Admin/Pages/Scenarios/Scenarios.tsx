import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button";
import logo from "/logo.svg";
import { onboardingAPI } from "../../../../Api/onboarding";
import type { Tour } from "../../../../types/sdk";

function ScenariosComponent() {
  const navigate = useNavigate();
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const queryClient = useQueryClient();
  const {
    data: scenarios = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ["tours"],
    queryFn: async () => {
      // Если API готово — раскомментируйте:
      // return await onboardingAPI.getAll();
      
      // Пока используем localStorage
      const data = JSON.parse(localStorage.getItem("tours") || "[]");
      return data as Tour[];
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      // Если API готово — раскомментируйте:
      // await onboardingAPI.deleteTour(id);
      
      // Пока через localStorage
      const updated = scenarios.filter((item) => item.id !== id);
      localStorage.setItem("tours", JSON.stringify(updated));
      return updated;
    },
    onSuccess: (updatedScenarios) => {
      queryClient.setQueryData(["tours"], updatedScenarios);
      setOpenMenu(null);
    },
    onError: (error) => {
      console.error("Ошибка удаления:", error);
      alert("Не удалось удалить сценарий");
    },
  });

  if (isLoading) {
    return (
        <div className={styles.loading}>Загрузка сценариев...</div>
    );
  }

  if (error) {
    return (
        <div className={styles.error}>
          Ошибка загрузки сценариев: {error.message}
        </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <img src={logo} className={styles.logo} />
        <div>
          <h1>Сценарии</h1>
        </div>
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

        {scenarios.map((item) => (
          <div className={styles.row} key={item.id}>
            <span className={styles.name}>{item.title}</span>

            <span>
              <span
                className={`
                  ${styles.status}
                  ${item.status === "published" ? styles.success : styles.warning}
                `}
              />
              {item.status === "published" ? "Опубликован" : "Черновик"}
            </span>

            <span>{item.hints.length}</span>

            <span>
              {item.updated_at
                ? new Date(item.updated_at).toLocaleDateString()
                : "—"}
            </span>

            <div className={styles.actions}>
              {item.status === "published" && (
                <Button
                  size="min"
                  onClick={() => {
                    window.open(`/?tour=${item.id}&preview=true`, "_blank");
                  }}
                >
                  Предпросмотр
                </Button>
              )}

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
                      className={styles.deleteButton}
                      onClick={() => deleteMutation.mutate(item.id)}
                    >
                      Удалить
                    </button>
                    <Button
                      size="min"
                      onClick={() =>
                        navigate(`/admin/scenarios/create?id=${item.id}`)
                      }
                    >
                      Редактировать
                    </Button>
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export const Scenarios = memo(ScenariosComponent);