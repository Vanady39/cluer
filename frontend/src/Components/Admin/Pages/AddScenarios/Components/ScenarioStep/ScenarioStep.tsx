import { memo } from "react";
import { Controller } from "react-hook-form";
import type { Control, FieldErrors } from "react-hook-form";
import type { InferType } from "yup";
import { Input } from "../../../../../UI/Input/Input";
import { Button } from "../../../../../UI/Button/Button";
import { scenarioSchema } from "../../schema";
import styles from "./Styles.module.scss";
import cn from "classnames";
import { useSortable } from "@dnd-kit/react/sortable";

interface ScenarioStepProps {
  id: string;
  index: number;
  control: Control<InferType<typeof scenarioSchema>>;
  errors: FieldErrors<InferType<typeof scenarioSchema>>;
  onRemove: (index: number) => void;
  onSelectElement: (index: number) => void;
}

function ScenarioStepComponent({ id, index, control, errors, onRemove, onSelectElement }: ScenarioStepProps) {
  const { ref, handleRef, isDragging } = useSortable({ id, index });
  return (
    <div ref={ref} className={cn(styles.step, isDragging && styles.step_dragging)}>
      <div className={styles.step__stepTitle}>
        <div className={styles.step__titleLeft}>
          <button
            type="button"
            ref={handleRef}
            className={styles.step__dragHandle}
            aria-label={`Перетащить шаг ${index + 1}`}
            title="Перетащите, чтобы изменить порядок"
          >
            <span />
            <span />
            <span />
            <span />
            <span />
            <span />
          </button>
          <span>Шаг {index + 1}</span>
        </div>
        <Button
          size="min"
          color="transparent"
          className={styles.step__removeButton}
          onClick={() => onRemove(index)}
        >
          ×
        </Button>
      </div>

      <Controller
        name={`hints.${index}.title`}
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.step__label}>Название подсказки</label>
            <Input
              {...field}
              placeholder="Например: Создайте объявление"
              error={errors.hints?.[index]?.title?.message}
            />
          </>
        )}
      />

      <Controller
        name={`hints.${index}.content`}
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.step__label}>Текст подсказки</label>
            <textarea
              {...field}
              className={cn(
                styles.step__textarea,
                errors.hints?.[index]?.content?.message &&
                  styles.step__textarea__error,
              )}
              placeholder="Нажмите сюда..."
            />
            {errors.hints?.[index]?.content?.message && (
              <span className={styles.step__error}>
                {errors.hints[index]?.content?.message}
              </span>
            )}
          </>
        )}
      />

      <Controller
        name={`hints.${index}.page_path`}
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.step__label}>Страница элемента</label>
            <select {...field} className={styles.step__select}>
              <option value="/">Главная</option>
              <option value="/addItem">Создание объявления</option>
              <option value="/profile">Профиль</option>
            </select>
            {errors.hints?.[index]?.page_path?.message && (
              <span className={styles.step__error}>
                {errors.hints[index]?.page_path?.message}
              </span>
            )}
          </>
        )}
      />

      <Controller
        name={`hints.${index}.placement`}
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.step__label}>Позиция подсказки</label>
            <select {...field} className={styles.step__select}>
              <option value="bottom">Снизу</option>
              <option value="top">Сверху</option>
              <option value="left">Слева</option>
              <option value="right">Справа</option>
            </select>
            {errors.hints?.[index]?.placement?.message && (
              <span className={styles.step__error}>
                {errors.hints[index]?.placement?.message}
              </span>
            )}
          </>
        )}
      />

      <Controller
        name={`hints.${index}.selector`}
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.step__label}>Элемент сайта</label>
            <div className={styles.step__selectorRow}>
              <Input
                {...field}
                className={styles.step__selectorInput}
                value={field.value || ""}
                placeholder="[data-onboarding='create-ad']"
                error={errors.hints?.[index]?.selector?.message}
              />
              <Button
                size="min"
                color="primary"
                className={styles.step__pickButton}
                onClick={() => onSelectElement(index)}
              >
                Выбрать
              </Button>
            </div>
          </>
        )}
      />
    </div>
  );
}

export const ScenarioStep = memo(ScenarioStepComponent);
