import { memo, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import styles from "./Styles.module.scss";
import { Input } from "../../../UI/Input/Input";
import { Button } from "../../../UI/Button/Button";
import type { TourHint, Tour } from "../../../../types/sdk";

// API импорт закомментирован
// import { onboardingAPI } from "../../../../Api/onboarding";
// import type { CreateTourRequest, CreateHintRequest } from "../../../../types/sdk";

function CreateScenariosComponent() {
  const navigate = useNavigate();
  const editId = new URLSearchParams(window.location.search).get("id");
  const queryClient = useQueryClient();
  const USE_API = false; 

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [hints, setHints] = useState<TourHint[]>([
    {
      id: String(Date.now()),
      tour_id: "",
      step: 1,
      title: "Первый шаг",
      content: "",
      selector: "",
      placement: "bottom",
      target_path: "/",
      spotlight: true,
      required: false,
      wait_for_selector: false,
    },
  ]);

  const { data: loadedTour } = useQuery({
    queryKey: ["tour", editId],
    queryFn: async () => {
      if (!editId) return null;
      
      //API-ветка закомментирована
      // if (USE_API) {
      //   return await onboardingAPI.getTourById(editId);
      // }
      
      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      return tours.find((item: Tour) => item.id === editId) || null;
    },
    enabled: !!editId, 
  });

  useEffect(() => {
    if (loadedTour) {
      setTitle(loadedTour.title);
      setDescription(loadedTour.description);
      setHints(loadedTour.hints);
    }
  }, [loadedTour]);

  const addHint = () => {
    setHints((prev) => [
      ...prev,
      {
        id: String(Date.now()),
        tour_id: "",
        step: prev.length + 1,
        title: `Шаг ${prev.length + 1}`,
        content: "",
        selector: "",
        placement: "bottom",
        target_path: "/",
        spotlight: true,
        required: false,
        wait_for_selector: false,
      },
    ]);
  };

  const updateHint = (
    id: string,
    field: keyof TourHint,
    value: string | boolean,
  ) => {
    setHints((prev) =>
      prev.map((hint) =>
        hint.id === id
          ? {
              ...hint,
              [field]: value,
            }
          : hint,
      ),
    );
  };

  const removeHint = (id: string) => {
    setHints((prev) => prev.filter((hint) => hint.id !== id));
  };

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
  }, [hints]);

  const saveMutation = useMutation({
    mutationFn: async (scenarioStatus: "draft" | "published") => {
      // API-логика закомментирована
      // const tourData: CreateTourRequest = {
      //   title,
      //   target_path: "/",
      //   description,
      //   priority: 1,
      //   trigger_type: "on_load",
      //   audience: {
      //     show_once: true,
      //     max_shows: 1,
      //     only_new: false,
      //   },
      // };
      // let tourId = editId;
      // if (!tourId) {
      //   tourId = await onboardingAPI.createTour(tourData);
      // } else {
      //   await onboardingAPI.updateTour(tourId, tourData);
      // }
      // for (const hint of hints) {
      //   const hintData: CreateHintRequest = {
      //     title: hint.title,
      //     content: hint.content,
      //     placement: hint.placement,
      //     selector: hint.selector || undefined,
      //     spotlight: hint.spotlight,
      //     required: hint.required,
      //     wait_for_selector: hint.wait_for_selector,
      //   };
      //   if (hint.id.startsWith("new-") || !hint.id) {
      //     await onboardingAPI.createHint(tourId, hintData);
      //   } else {
      //     await onboardingAPI.updateHint(tourId, hint.id, hintData);
      //   }
      // }
      // if (scenarioStatus === "published") {
      //   await onboardingAPI.publishTour(tourId);
      // }
      
      const scenario = {
        id: editId || String(Date.now()),
        title,
        description,
        status: scenarioStatus,
        target_path: "/",
        priority: 1,
        trigger_type: "on_load",
        audience: {
          show_once: true,
          max_shows: 1,
          only_new: false,
        },
        hints,
        updated_at: new Date().toISOString()
      };

      const tours = JSON.parse(localStorage.getItem("tours") || "[]");
      const updatedTours = editId
        ? tours.map((tour: Tour) => tour.id === editId ? scenario : tour)
        : [...tours, scenario];

      localStorage.setItem("tours", JSON.stringify(updatedTours));
      return scenario.id;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tours"] });
      navigate("/admin/scenarios");
    },
    onError: (error) => {
      console.error("Ошибка сохранения:", error);
      alert("Не удалось сохранить сценарий");
    },
  });

  const saveScenario = (scenarioStatus: "draft" | "published") => {
    saveMutation.mutate(scenarioStatus);
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
                value={hint.target_path}
                onChange={(e) => updateHint(hint.id, "target_path", e.target.value)}
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
                    const page = hint.target_path || "/";
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