import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button";
import logo from "/logo.svg";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import { useDeleteTourMutation } from "../../../../Hooks/useDeleteTourMutation";

function ScenariosComponent() {
  const navigate = useNavigate();
  const [openMenu, setOpenMenu] = useState<string | null>(null);

  const { data: scenarios = [], isLoading, error } = useToursQuery();
  const deleteMutation = useDeleteTourMutation();

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
    setOpenMenu(null);
  };

  if (isLoading) {
    return <div className={styles.loading}>Загрузка сценариев...</div>;
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
              <Button
                size="min"
                onClick={() => {
                  window.open(`/?tour=${item.id}&preview=true`, "_blank");
                }}
              >
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
                      className={styles.deleteButton}
                      onClick={() => handleDelete(item.id)}
                    >
                      Удалить
                    </button>

                    {item.status !== "published" && (
                      <Button
                        size="min"
                        onClick={() =>
                          navigate(`/admin/scenarios/create?id=${item.id}`)
                        }
                      >
                        Редактировать
                      </Button>
                    )}
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