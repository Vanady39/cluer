import { memo } from "react";
import { useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button/Button";
import { useCreateScenario } from "../../../../Hooks/useCreateScenario";
import { ScenarioMainFields } from "./Components/ScenarioMainFields/ScenarioMainFields";
import { ScenarioStep } from "./Components/ScenarioStep/ScenarioStep";

function CreateScenariosComponent() {
  const navigate = useNavigate();
  const editId = new URLSearchParams(window.location.search).get("id");
  const { createForm, saveDraft, onSubmit, addHint, removeHint, selectElement, isPending, errors, } = useCreateScenario(editId);

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

            {createForm.watch("hints").map((hint, index) => (
              <ScenarioStep
                key={hint.id}
                index={index}
                control={createForm.control}
                errors={createForm.formState.errors}
                onRemove={removeHint}
                onSelectElement={selectElement}
              />
            ))}

            <div>
              <Button
                type="button"
                size="main"
                color="primary"
                onClick={addHint}
              >
                + Добавить шаг
              </Button>
            </div>
          </section>
        </div>
        <div className={styles.page__actions}>
          <Button
            type="button"
            size="main"
            onClick={() => navigate("/admin/scenarios")}
          >
            Отмена
          </Button>
          <Button
            type="button"
            size="main"
            onClick={saveDraft}
            disabled={isPending}
          >
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

export const CreateScenarios = memo(CreateScenariosComponent);