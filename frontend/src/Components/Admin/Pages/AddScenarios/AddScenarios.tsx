import { memo, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Input } from "../../../UI/Input/Input";
import { Button } from "../../../UI/Button/Button";
import { useTourLoader } from "../../../../Hooks/useTourLoader";
import { useHintManager } from "../../../../Hooks/useHintManager";
import { useSaveScenario } from "../../../../Hooks/useSaveScenario";

function CreateScenariosComponent() {
  const navigate = useNavigate();
  const editId = new URLSearchParams(window.location.search).get("id");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const { data: loadedTour } = useTourLoader(editId);
  const [hasLoaded, setHasLoaded] = useState(false);
  const { hints, addHint, updateHint, removeHint, setHints } = useHintManager([]);

  useEffect(() => {
  if (loadedTour && !hasLoaded) {
    setTitle(loadedTour.title);
    setDescription(loadedTour.description);
    if (loadedTour.hints?.length) {
      setHints(loadedTour.hints);
    }
    setHasLoaded(true);
  }
}, [loadedTour, hasLoaded]);

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (event.data?.type !== "SELECTOR_SELECTED") return;

      const selector = event.data.selector;
      if (!selector) return;

      if (hints.length === 0) {
        alert("Сначала добавьте хотя бы один шаг!");
        return;
      }
      const currentHint = hints[hints.length - 1];
      updateHint(currentHint.id, "selector", selector);
    };
    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [hints, updateHint]);

  const saveMutation = useSaveScenario(editId);
  const saveScenario = (scenarioStatus: "draft" | "published") => {
    saveMutation.mutate({
      title,
      description,
      hints,
      status: scenarioStatus,
    });
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
        </section>

        <section className={styles.card}>
          <div className={styles.stepHeader}>
            <h2>Шаги сценария</h2>
            <Button
              size="main"
              color="primary"
              className={styles.addButton}
              onClick={addHint}
            >
              + Добавить шаг
            </Button>
          </div>

          {hints.map((hint, index) => (
            <div className={styles.step} key={hint.id}>
              <div className={styles.stepTitle}>
                Шаг {index + 1}
                <Button
                  size="min"
                  color="transparent"
                  className={styles.removeButton}
                  onClick={() => removeHint(hint.id)}
                >
                  ×
                </Button>
              </div>

              <label className={styles.label}>Название подсказки</label>
              <Input
                value={hint.title}
                onChange={(value) =>
                  updateHint(hint.id, "title", String(value))
                }
                placeholder="Например: Создайте объявление"
              />

              <label className={styles.label}>Текст подсказки</label>
              <textarea
                className={styles.textarea}
                value={hint.content}
                placeholder="Нажмите сюда..."
                onChange={(e) => updateHint(hint.id, "content", e.target.value)}
              />

              <label className={styles.label}>Страница элемента</label>
              <select
                className={styles.select}
                value={hint.page_path}
                onChange={(e)=>updateHint(hint.id,"page_path",e.target.value)}
                >
                <option value="/">Главная</option>
                <option value="/addItem">Создание объявления</option>
                <option value="/profile">Профиль</option>
              </select>

              <label className={styles.label}>Позиция подсказки</label>
              <select
                className={styles.select}
                value={hint.placement}
                onChange={(e) => updateHint(hint.id, "placement", e.target.value)}
              >
                <option value="bottom">Снизу</option>
                <option value="top">Сверху</option>
                <option value="left">Слева</option>
                <option value="right">Справа</option>
                <option value="center">По центру</option>
              </select>

              <label className={styles.label}>Элемент сайта</label>
              <div className={styles.selectorRow}>
                <Input
                  className={styles.selectorInput}
                  value={hint.selector || ""}
                  placeholder="[data-onboarding='create-ad']"
                  onChange={(value) =>
                    updateHint(hint.id, "selector", String(value))
                  }
                />
                <Button
                  size="min"
                  color="primary"
                  className={styles.pickButton}
                  onClick={() => {
                    const page = hint.page_path || "/";
                    const builderUrl = `${window.location.origin}${page}?builder=true`;
                    window.open(builderUrl, "_blank");
                  }}
                >
                  Выбрать
                </Button>
              </div>
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
          onClick={() => saveScenario("draft")}
          disabled={saveMutation.isPending}
        >
          {saveMutation.isPending ? "Сохранение..." : "Сохранить черновик"}
        </Button>
        <Button
          size="main"
          color="primary"
          className={styles.save}
          onClick={() => saveScenario("published")}
          disabled={saveMutation.isPending}
        >
          {saveMutation.isPending ? "Сохранение..." : "Опубликовать"}
        </Button>
      </div>
    </div>
  );
}

export const CreateScenarios = memo(CreateScenariosComponent);