import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Input } from "../../../UI/Input/Input";
import { Button } from "../../../UI/Button/Button";

interface Step {
  id: number;
  target: string;
  text: string;
  position: string;
}

function CreateScenariosComponent() {
  const navigate = useNavigate();
  const [steps, setSteps] = useState<Step[]>([
    {
      id: 1,
      target: "",
      text: "",
      position: "bottom",
    },
  ]);

  const addStep = () => {
    setSteps((prev) => [
      ...prev,
      {
        id: Date.now(),
        target: "",
        text: "",
        position: "bottom",
      },
    ]);
  };

  const updateStep = (id: number, field: keyof Step, value: string) => {
    setSteps((prev) =>
      prev.map((step) =>
        step.id === id ? { ...step, [field]: value } : step
      )
    );
  };

  const removeStep = (id: number) => {
    setSteps((prev) => prev.filter((step) => step.id !== id));
  };

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1>Создание сценария</h1>
          <p>Настройте путь пользователя</p>
        </div>
      </div>

      <div className={styles.layout}>
        <section className={styles.card}>
          <h2>Основная информация</h2>

          <label className={styles.label}>Название сценария</label>
          <Input
            placeholder="Например: Первое объявление"
            className={styles.inputField}
          />

          <label className={styles.label}>Описание</label>
          <textarea
            className={styles.textarea}
            placeholder="Зачем нужен этот сценарий"
          />

          <label className={styles.label}>Статус</label>
          <select className={styles.select}>
            <option>Черновик</option>
            <option>Опубликован</option>
          </select>
        </section>

        {/* Шаги */}
        <section className={styles.card}>
          <div className={styles.stepHeader}>
            <h2>Шаги сценария</h2>
            <Button
              size="main"
              color="primary"
              className={styles.addButton}
              onClick={addStep}
            >
              + Добавить шаг
            </Button>
          </div>

          {steps.map((step, index) => (
            <div className={styles.step} key={step.id}>
              <div className={styles.stepTitle}>
                Шаг {index + 1}
                <Button
                  size="min"
                  color="transparent"
                  className={styles.removeButton}
                  onClick={() => removeStep(step.id)}
                >
                  ×
                </Button>
              </div>

              <label className={styles.label}>Элемент сайта</label>
              <Input
                value={step.target}
                placeholder="Например: create-ad-button"
                onChange={(value) =>
                  updateStep(step.id, "target", String(value))
                }
                className={styles.inputField}
              />

              <label className={styles.label}>Текст подсказки</label>
              <textarea
                className={styles.textarea}
                value={step.text}
                placeholder="Нажмите сюда, чтобы создать объявление"
                onChange={(e) => updateStep(step.id, "text", e.target.value)}
              />

              <label className={styles.label}>Позиция</label>
              <select
                className={styles.select}
                value={step.position}
                onChange={(e) =>
                  updateStep(step.id, "position", e.target.value)
                }
              >
                <option value="bottom">Снизу</option>
                <option value="top">Сверху</option>
                <option value="left">Слева</option>
                <option value="right">Справа</option>
              </select>
            </div>
          ))}
        </section>
      </div>

      <div className={styles.actions}>
        <Button size="main" onClick={() => navigate("/admin/scenarios")}>
          Отмена
        </Button>
        <Button size="main" color="primary" className={styles.save}>
          Опубликовать
        </Button>
      </div>
    </div>
  );
}

export const CreateScenarios = memo(CreateScenariosComponent);