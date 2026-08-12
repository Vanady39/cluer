import * as yup from "yup";

export const scenarioSchema = yup.object({
  title: yup.string().trim().required("Введите название сценария").min(3, "Название должно содержать минимум 3 символа"),

  description: yup.string().trim(),

  trigger_type: yup.string().required("Выберите тип запуска"),

  audience: yup
    .object({
      show_once: yup.boolean().required(),
      max_shows: yup
        .number()
        .typeError("Введите количество показов")
        .integer("Количество показов должно быть целым числом")
        .min(1, "Минимум 1 показ")
        .required("Введите количество показов"),

      only_new: yup.boolean().required(),
    }).required(),

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