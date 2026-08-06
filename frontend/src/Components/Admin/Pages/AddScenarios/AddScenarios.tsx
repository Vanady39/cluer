import { memo, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import styles from "./Styles.module.scss";

import { Input } from "../../../UI/Input/Input";
import { Button } from "../../../UI/Button/Button";

interface Step {
  id: number;
  title: string;
  content: string;
  selector: string;
  placement: string;
  spotlight: boolean;
  targetPage: string;
}

interface Tour {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published";
  target_path: string;
  priority: number;
  hints: Step[];
}

function CreateScenariosComponent() {
  const navigate = useNavigate();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [targetPath, setTargetPath] = useState("/");
  const [status, setStatus] = useState<"draft" | "published">("published");
  const [steps, setSteps] = useState<Step[]>([
    {
      id: 1,
      title: "Первый шаг",
      content: "",
      selector: "",
      placement: "bottom",
      spotlight: true,
      targetPage: "/",
    },
  ]);

  const addStep = () => {
    setSteps((prev) => [
      ...prev,
      {
        id: Date.now(),
        title: `Шаг ${prev.length + 1}`,
        content: "",
        selector: "",
        placement: "bottom",
        spotlight: true,
        targetPage: "/",
      },
    ]);
  };

  const updateStep = (
    id: number,
    field: keyof Step,
    value: string | boolean,
  ) => {
    setSteps((prev) =>
      prev.map((step) =>
        step.id === id
          ? {
              ...step,
              [field]: value,
            }
          : step,
      ),
    );
  };

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type === "SELECTOR_SELECTED") {
        const selector = event.data.selector;
        if (!selector) return;

        if (steps.length === 0) {
          alert("Сначала добавьте хотя бы один шаг!");
          return;
        }

        const currentStep = steps[steps.length - 1];
        updateStep(currentStep.id, "selector", selector);
      }
    };

    window.addEventListener("message", handleMessage);

    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, [steps]);

  const removeStep = (id: number) => {
    setSteps((prev) => prev.filter((step) => step.id !== id));
  };

  const saveScenario = () => {
    const scenario: Tour = {
      id: String(Date.now()),
      title,
      description,
      status,
      target_path: targetPath,
      priority: 1,
      hints: steps.map((step, index) => ({
        id: step.id,
        title: step.title || `Шаг ${index + 1}`,
        content: step.content,
        selector: step.selector,
        placement: step.placement,
        spotlight: step.spotlight,
        targetPage: step.targetPage, 
      })),
    };

    const oldTours = JSON.parse(localStorage.getItem("tours") || "[]");
    localStorage.setItem("tours", JSON.stringify([...oldTours, scenario]));
    navigate("/admin/scenarios");
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
            value={title}
            onChange={(value) => setTitle(String(value))}
            placeholder="Например: Первое объявление"
          />

          <label className={styles.label}>Описание</label>
          <textarea
            className={styles.textarea}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Зачем нужен этот сценарий"
          />
          <label className={styles.label}>Статус</label>
          <select
            className={styles.select}
            value={status}
            onChange={(e) => setStatus(e.target.value as "draft" | "published")}
          >
            <option value="draft">Черновик</option>
            <option value="published">Опубликован</option>
          </select>
        </section>

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

              <label className={styles.label}>Название подсказки</label>
              <Input
                value={step.title}
                onChange={(value) =>
                  updateStep(step.id, "title", String(value))
                }
                placeholder="Например: Создайте объявление"
              />

              <label className={styles.label}>Текст подсказки</label>
              <textarea
                className={styles.textarea}
                value={step.content}
                placeholder="Нажмите сюда..."
                onChange={(e) => updateStep(step.id, "content", e.target.value)}
              />

              <label className={styles.label}>Страница элемента</label>
              <select
                className={styles.select}
                value={step.targetPage || "/"}
                onChange={(e) =>
                  updateStep(step.id, "targetPage", e.target.value)
                }
              >
                <option value="/">Главная</option>
                <option value="/addItem">Создание объявления</option>
                <option value="/profile">Профиль</option>
              </select>
              <label className={styles.label}>Элемент сайта</label>
              <div className={styles.selectorRow}>
                <Input
                  className={styles.selectorInput}
                  value={step.selector}
                  placeholder="[data-onboarding='create-ad']"
                  onChange={(value) =>
                    updateStep(step.id, "selector", String(value))
                  }
                />
                <Button
                  size="min"
                  color="primary"
                  className={styles.pickButton}
                  onClick={() => {
                    const page = step.targetPage || "/";
                    const builderUrl = `${window.location.origin}${page}?builder=true`;
                    window.open(builderUrl, "_blank");
                  }}
                >
                  Выбрать
                </Button>
              </div>

              <label className={styles.label}>Позиция подсказки</label>
              <select
                className={styles.select}
                value={step.placement}
                onChange={(e) =>
                  updateStep(step.id, "placement", e.target.value)
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
        <Button
          size="main"
          color="primary"
          className={styles.save}
          onClick={saveScenario}
        >
          Опубликовать
        </Button>
      </div>
    </div>
  );
}

export const CreateScenarios = memo(CreateScenariosComponent);
