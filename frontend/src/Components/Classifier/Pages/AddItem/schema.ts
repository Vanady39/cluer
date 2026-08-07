import * as yup from "yup";

export const createAdSchema = yup.object().shape({
  category: yup.string().required("Выберите категорию"),

  title: yup
    .string()
    .required("Введите заголовок")
    .min(3, "Заголовок должен содержать минимум 3 символа")
    .max(100, "Заголовок не должен превышать 100 символов"),

  description: yup
    .string()
    .required("Введите описание")
    .min(10, "Описание должно содержать минимум 10 символов")
    .max(3000, "Описание не должно превышать 3000 символов"),

  price: yup
    .number()
    .required("Введите цену")
    .positive("Цена должна быть больше 0")
    .integer("Цена должна быть целым числом")
    .max(999999999, "Цена слишком большая"),

  city: yup.string().required("Выберите город"),
});
