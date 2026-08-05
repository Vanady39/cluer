import { memo } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button";
import logo from "/logo.svg";

const scenarios: {
  id: string;
  name: string;
  status: string;
  steps: string;
  updated: string;
}[] = [];

function ScenariosComponent() {
  const navigate = useNavigate();

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <img src={logo} className={styles.logo}/>
        <h1>Сценарии</h1>
        <Button
          className={styles.createButton} size="main"
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
            <span className={styles.name}>{item.name}</span>
            <span>
              <span
                className={`${styles.status} ${item.status === "published" ? styles.success : styles.warning}`}>
                </span>
              {item.status === "published" ? "Опубликован" : "Черновик"}
            </span>

            <span>{item.steps}</span>
            <span>{item.updated}</span>

            <div className={styles.actions}>
              <Button size="min">Открыть</Button>
              <Button size="min" className={styles.more}>⋮</Button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export const Scenarios = memo(ScenariosComponent);