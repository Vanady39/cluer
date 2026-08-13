import { memo } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button/Button";
import { useCreateScenario } from "../../../../Hooks/useCreateScenario";
import { ScenarioMainFields } from "./Components/ScenarioMainFields/ScenarioMainFields";
import { ScenarioStep } from "./Components/ScenarioStep/ScenarioStep";
import { DragDropProvider } from "@dnd-kit/react";
import { isSortable } from "@dnd-kit/react/sortable";

function AddScenariosComponent() {
  const navigate = useNavigate();
  const editId = new URLSearchParams(window.location.search).get("id");
  const {
    createForm,
    saveDraft,
    onSubmit,
    addHint,
    removeHint,
    reorderHint,
    selectElement,
    isPending,
    isTourLoading,
    isTourLoadError,
    tourLoadError,
    errors,
  } = useCreateScenario(editId);

  const hints = createForm.watch("hints") ?? [];

  if (editId && isTourLoading) return <div>Подготовка сценария к редактированию...</div>;

  if (editId && isTourLoadError) {
    return (
      <div>
        Не удалось подготовить сценарий к редактированию:{" "}
        {tourLoadError instanceof Error ? tourLoadError.message : "Неизвестная ошибка"}
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <div className={styles.page__header}>
        <div>
          <h1>Создание сценария</h1>
          <p>Настройте путь пользователя</p>
        </div>
      </div>
      <form onSubmit={createForm.handleSubmit(onSubmit)}>
        <div className={styles.page__layout}>
          <ScenarioMainFields
            control={createForm.control}
            errors={createForm.formState.errors}
          />
          <section className={styles.page__card}>
            <div className={styles.page__stepHeader}>
              <h2>Шаги сценария</h2>
            </div>
            {errors.hints?.message && (
              <span className={styles.page__error}>{errors.hints.message}</span>
            )}

            <DragDropProvider
              onDragEnd={(event) => {
                if (event.canceled) return;

                const { source } = event.operation;
                if (!isSortable(source)) return;

                const { initialIndex, index } = source;
                if (initialIndex !== index) reorderHint(initialIndex, index);
              }}
            >
              {hints.map((hint, index) => (
                <ScenarioStep
                  key={hint.id}
                  id={hint.id}
                  index={index}
                  control={createForm.control}
                  errors={createForm.formState.errors}
                  onRemove={removeHint}
                  onSelectElement={selectElement}
                />
              ))}
            </DragDropProvider>

            <div>
              <Button size="main" color="primary" onClick={addHint}>
                + Добавить шаг
              </Button>
            </div>
          </section>
        </div>
        <div className={styles.page__actions}>
          <Button size="main" onClick={() => navigate("/admin/scenarios")}>
            Отмена
          </Button>
          <Button size="main" onClick={saveDraft} disabled={isPending}>
            {isPending ? "Сохранение..." : "Сохранить черновик"}
          </Button>
          <Button
            type="submit"
            size="main"
            color="primary"
            className={styles.page__save}
            disabled={isPending}
          >
            {isPending ? "Сохранение..." : "Опубликовать"}
          </Button>
        </div>
      </form>
    </div>
  );
}

export const AddScenarios = memo(AddScenariosComponent);
