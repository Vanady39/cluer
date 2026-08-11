import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm, Controller } from "react-hook-form";
import { yupResolver } from "@hookform/resolvers/yup";
import type { InferType } from "yup";
import { createAdSchema } from "./schema";
import { Input } from "../../../UI/Input";
import { Button } from "../../../UI/Button";
import { CitySelect } from "../../../UI/CitySelect/CitySelect";
import { Categories } from "../../../UI/CategoriesSelect/CategoriesSelect";
import { Upload } from "../../../UI/Upload/Upload";
import { PreviewModal } from "../../Layouts/PreviewModal/PreviewModal";
import styles from "./Styles.module.scss";
import { trackOnboardingGoal } from "../../../Onboarding/goal";

interface FileItem {
  uid: string;
  name: string;
  url?: string;
  file?: File;
  size?: number;
}

function AddItemComponent() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [fileList, setFileList] = useState<FileItem[]>([]);
  const [previewImage, setPreviewImage] = useState<string | null>(null);

  const createForm = useForm({
    resolver: yupResolver(createAdSchema),
    defaultValues: {
      category: "",
      title: "",
      description: "",
      city: "",
    },
  });

  const onSubmit = async (data: InferType<typeof createAdSchema>) => {
    setLoading(true);

    try {
      const formData = new FormData();

      Object.entries(data).forEach(([key, value]) => {
        formData.append(key, String(value));
      });

      fileList.forEach((item) => {
        if (item.file) {
          formData.append("images", item.file);
        }
      });

      console.log("Отправка:", Object.fromEntries(formData));

      trackOnboardingGoal("listing_created", {
        title: data.title,
        category: data.category,
      });

      alert("Объявление создано!");

      navigate("/");
    } catch {
      alert("Ошибка при создании");
    } finally {
      setLoading(false);
    }
  };

  const openPreview = (url: string) => setPreviewImage(url);
  const closePreview = () => setPreviewImage(null);

  return (
    <main className={styles.page}>
      <section className={styles.page__card}>
        <h1 className={styles.page__card__title}>Создать объявление</h1>

        <form
          className={styles.page__card__form}
          onSubmit={createForm.handleSubmit(onSubmit)}
        >
          <Controller
            name="category"
            control={createForm.control}
            render={({ field }) => (
              <Categories
                value={field.value}
                onChange={field.onChange}
                label="Категория"
                error={createForm.formState.errors.category?.message}
              />
            )}
          />

          <Controller
            name="title"
            control={createForm.control}
            render={({ field }) => (
              <Input
                {...field}
                label="Заголовок"
                placeholder="Например: iPhone 15 Pro Max"
                error={createForm.formState.errors.title?.message}
              />
            )}
          />

          <div className={styles.page__card__form__field}>
            <label className={styles.page__card__form__field__label}>Описание</label>
            <Controller
              name="description"
              control={createForm.control}
              render={({ field }) => (
                <textarea {...field} className={styles.page__card__form__textarea} rows={6} />
              )}
            />
          </div>

          <div className={styles.page__card__form__row}>
            <Controller
              name="price"
              control={createForm.control}
              render={({ field }) => (
                <Input
                  {...field}
                  label="Цена"
                  type="number"
                  placeholder="₽"
                  value={field.value || ""}
                  onChange={field.onChange}
                  error={createForm.formState.errors.price?.message}
                />
              )}
            />

            <Controller
              name="city"
              control={createForm.control}
              render={({ field }) => (
                <CitySelect
                  value={field.value || ""}
                  onChange={field.onChange}
                  label="Город"
                  error={createForm.formState.errors.city?.message}
                />
              )}
            />
          </div>
          <div data-onboarding="photo-upload">
            <Upload
              fileList={fileList}
              onFileChange={setFileList}
              openPreview={openPreview}
            />
          </div>

          <div className={styles.page__card__form__actions}>
            <Button
              type="button"
              size="main"
              onClick={() => navigate("/")}
              className={styles.page__card__form__actions__button}
            >
              Отмена
            </Button>
            <Button
              type="submit"
              disabled={loading}
              color="primary"
              size="main"
              className={styles.page__card__form__actions__button_pub}
            >
              {loading ? "Публикация..." : "Опубликовать"}
            </Button>
          </div>
        </form>
      </section>
      <PreviewModal image={previewImage} onClose={closePreview} />
    </main>
  );
}

export const AddItem = memo(AddItemComponent);