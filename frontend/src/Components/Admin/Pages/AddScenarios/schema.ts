import * as yup from "yup";
import type { TriggerType } from "../../../../types/tour";

const optionalNumber = () =>
  yup.number().transform((value, originalValue) => {
    if (
      originalValue === "" ||
      originalValue === null ||
      originalValue === undefined ||
      Number.isNaN(originalValue)
    ) {
      return undefined;
    }

    return value;
  });

export const scenarioSchema = yup.object({
  title: yup
    .string()
    .trim()
    .required("Введите название сценария")
    .min(3, "Название должно содержать минимум 3 символа"),

  description: yup.string().trim(),

  trigger_type: yup
    .mixed<TriggerType>()
    .oneOf([
      "on_load",
      "delay",
      "exit_intent",
      "manual",
      "scroll_depth",
      "inactivity",
    ])
    .required("Выберите тип запуска"),

  trigger_config: yup
    .object({
      delay_ms: optionalNumber().optional(),
      scroll_depth: optionalNumber().optional(),
      inactivity_secs: optionalNumber().optional(),
    })
    .when("trigger_type", {
      is: "delay",
      then: (schema) =>
        schema.shape({
          delay_ms: optionalNumber()
            .typeError("Введите задержку")
            .integer("Задержка должна быть целым числом")
            .min(100, "Минимум 100 мс")
            .max(600000, "Максимум 600000 мс")
            .required("Введите задержку"),
        }),
    })
    .when("trigger_type", {
      is: "scroll_depth",
      then: (schema) =>
        schema.shape({
          scroll_depth: optionalNumber()
            .typeError("Введите процент прокрутки")
            .integer("Процент должен быть целым числом")
            .min(1, "Минимум 1%")
            .max(100, "Максимум 100%")
            .required("Введите глубину прокрутки"),
        }),
    })
    .when("trigger_type", {
      is: "inactivity",
      then: (schema) =>
        schema.shape({
          inactivity_secs: optionalNumber()
            .typeError("Введите время бездействия")
            .integer("Время должно быть целым числом")
            .min(3, "Минимум 3 секунды")
            .required("Введите время бездействия"),
        }),
    })
    .default({}),

  audience: yup
    .object({
      show_once: yup.boolean().required(),

      max_shows: optionalNumber()
        .typeError("Введите количество показов")
        .integer("Количество показов должно быть целым числом")
        .min(1, "Минимум 1 показ")
        .required("Введите количество показов"),

      only_new: yup.boolean().required(),
    })
    .required(),

  hints: yup
    .array()
    .of(
      yup.object({
        id: yup.string().required(),
        step: yup.number().required(),
        title: yup.string().trim().required("Введите название подсказки"),
        content: yup.string().trim().required("Введите текст подсказки"),
        selector: yup.string().trim().required("Выберите элемент сайта"),
        placement: yup.string().required("Выберите позицию подсказки"),
        page_path: yup.string().required("Выберите страницу"),
        spotlight: yup.boolean().required(),
        wait_for_selector: yup.boolean().required(),
        media_url: yup.string().default(""),
      }),
    )
    .min(1, "Добавьте хотя бы один шаг")
    .required("Добавьте хотя бы один шаг"),
});