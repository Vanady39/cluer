import { memo } from "react";
import { Controller, useWatch } from "react-hook-form";
import type { Control, FieldErrors } from "react-hook-form";
import type { InferType } from "yup";
import { Input } from "../../../../../UI/Input";
import { scenarioSchema } from "../../schema";
import styles from "./Styles.module.scss";

interface ScenarioMainFieldsProps {
  control: Control<InferType<typeof scenarioSchema>>;
  errors: FieldErrors<InferType<typeof scenarioSchema>>;
}

const getNumberInputValue = (value: number | undefined) =>
  value === undefined || Number.isNaN(value) ? "" : value;

const getNumberFormValue = (value: string | number) =>
  value === "" ? Number.NaN : Number(value);

function ScenarioMainFieldsComponent({
  control,
  errors,
}: ScenarioMainFieldsProps) {
  const triggerType = useWatch({
    control,
    name: "trigger_type",
  });

  return (
    <section className={styles.card}>
      <h2>Основная информация</h2>

      <Controller
        name="title"
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.card__label}>Название сценария</label>
            <Input
              {...field}
              placeholder="Например: Первое объявление"
              error={errors.title?.message}
            />
          </>
        )}
      />

      <Controller
        name="description"
        control={control}
        render={({ field }) => (
          <>
            <label className={styles.card__label}>Описание</label>
            <textarea
              {...field}
              className={styles.card__textarea}
              placeholder="Зачем нужен этот сценарий"
            />

            {errors.description?.message && (
              <span className={styles.card__error}>
                {errors.description.message}
              </span>
            )}
          </>
        )}
      />

      <div className={styles.card__settingsGroup}>
        <Controller
          name="trigger_type"
          control={control}
          render={({ field }) => (
            <>
              <label className={styles.card__settingsGroup__label}>
                Тип запуска
              </label>

              <select
                {...field}
                className={styles.card__settingsGroup__select}
              >
                <option value="on_load">При загрузке</option>
                <option value="delay">С задержкой</option>
                <option value="exit_intent">При выходе</option>
                <option value="manual">Вручную</option>
                <option value="scroll_depth">По глубине прокрутки</option>
                <option value="inactivity">При бездействии</option>
              </select>

              {errors.trigger_type?.message && (
                <span className={styles.card__settingsGroup__error}>
                  {errors.trigger_type.message}
                </span>
              )}
            </>
          )}
        />

        {triggerType === "delay" && (
          <Controller
            name="trigger_config.delay_ms"
            control={control}
            render={({ field }) => (
              <>
                <label className={styles.card__label}>
                  Задержка перед показом, мс
                </label>

                <Input
                  inputMode="numeric"
                  className={styles.card__numberInput}
                  value={getNumberInputValue(field.value)}
                  onChange={(value) =>
                    field.onChange(getNumberFormValue(value))
                  }
                  error={errors.trigger_config?.delay_ms?.message}
                />
              </>
            )}
          />
        )}

        {triggerType === "scroll_depth" && (
          <Controller
            name="trigger_config.scroll_depth"
            control={control}
            render={({ field }) => (
              <>
                <label className={styles.card__label}>
                  Глубина прокрутки, %
                </label>

                <Input
                  inputMode="numeric"
                  className={styles.card__numberInput}
                  value={getNumberInputValue(field.value)}
                  onChange={(value) =>
                    field.onChange(getNumberFormValue(value))
                  }
                  error={errors.trigger_config?.scroll_depth?.message}
                />
              </>
            )}
          />
        )}

        {triggerType === "inactivity" && (
          <Controller
            name="trigger_config.inactivity_secs"
            control={control}
            render={({ field }) => (
              <>
                <label className={styles.card__label}>
                  Время бездействия, сек.
                </label>

                <Input
                  inputMode="numeric"
                  className={styles.card__numberInput}
                  value={getNumberInputValue(field.value)}
                  onChange={(value) =>
                    field.onChange(getNumberFormValue(value))
                  }
                  error={errors.trigger_config?.inactivity_secs?.message}
                />
              </>
            )}
          />
        )}

        <Controller
          name="audience.show_once"
          control={control}
          render={({ field }) => (
            <label className={styles.card__checkboxRow}>
              <input
                type="checkbox"
                checked={field.value}
                onChange={(event) => field.onChange(event.target.checked)}
              />
              Показывать один раз
            </label>
          )}
        />

        <Controller
          name="audience.only_new"
          control={control}
          render={({ field }) => (
            <label className={styles.card__checkboxRow}>
              <input
                type="checkbox"
                checked={field.value}
                onChange={(event) => field.onChange(event.target.checked)}
              />
              Только новым пользователям
            </label>
          )}
        />

        <Controller
          name="audience.max_shows"
          control={control}
          render={({ field }) => (
            <>
              <label className={styles.card__label}>
                Максимальное количество показов
              </label>

              <Input
                inputMode="numeric"
                className={styles.card__numberInput}
                value={getNumberInputValue(field.value)}
                onChange={(value) =>
                  field.onChange(getNumberFormValue(value))
                }
                error={errors.audience?.max_shows?.message}
              />
            </>
          )}
        />
      </div>
    </section>
  );
}

export const ScenarioMainFields = memo(ScenarioMainFieldsComponent);