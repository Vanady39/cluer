import { memo, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import styles from "./Styles.module.scss";

import { Button } from "../../../UI/Button";
import logo from "/logo.svg";

interface Scenario {
  id: string;
  title: string;
  status: "draft" | "published";
  hints: unknown[];
  updated?: string;
}

function ScenariosComponent() {
  const navigate = useNavigate();
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [scenarios, setScenarios] = useState<Scenario[]>([]);

  useEffect(() => {
    const data = JSON.parse(localStorage.getItem("tours") || "[]");

    setScenarios(data);
  }, []);

  const deleteScenario = (id: string) => {
    const updated = scenarios.filter((item) => item.id !== id);

    localStorage.setItem("tours", JSON.stringify(updated));

    setScenarios(updated);
  };

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
                    ${
                      item.status === "published"
                        ? styles.success
                        : styles.warning
                    }
                  `}
              />

              {item.status === "published" ? "Опубликован" : "Черновик"}
            </span>

            <span>{item.hints.length}</span>

            <span>{new Date().toLocaleDateString()}</span>

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
                      onClick={() => {
                        deleteScenario(item.id);
                        setOpenMenu(null);
                      }}
                    >
                      Удалить
                    </button>
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
