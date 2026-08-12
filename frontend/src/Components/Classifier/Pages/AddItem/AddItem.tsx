import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Controller } from "react-hook-form";
import { Input } from "../../../UI/Input";
import { Button } from "../../../UI/Button";
import { CitySelect } from "../../../UI/CitySelect/CitySelect";
import { Categories } from "../../../UI/CategoriesSelect/CategoriesSelect";
import { Upload } from "../../../UI/Upload/Upload";
import { PreviewModal } from "../../Layouts/PreviewModal/PreviewModal";
import styles from "./Styles.module.scss";
import { useCreateAd } from "../../../../Hooks/useCreateAd";
import { useFileUpload } from "../../../../Hooks/useFileUpload";
import cn from "classnames";
import type { InferType } from "yup";
import type { createAdSchema } from "./schema";


function AddItemComponent() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { createForm, submitAd } = useCreateAd();
  const { fileList, setFileList, previewImage, openPreview, closePreview } =
    useFileUpload();

  const handleSubmit = async (data: InferType<typeof createAdSchema>) => {
    setLoading(true);
    const files = fileList.map((item) => item.file).filter(Boolean) as File[];
    await submitAd(data, files, () => navigate("/"), () => alert("Ошибка при создании"));
    setLoading(false);
  };

  return (
    <main className={styles.page}>
      <section className={styles.page__card}>
        <h1 className={styles.page__card__title}>Создать объявление</h1>

        <form
          className={styles.page__card__form}
          onSubmit={createForm.handleSubmit(handleSubmit)}
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
              render={({ field }) => {
                const error = createForm.formState.errors.description?.message;
                return (
                  <>
                    <textarea
                      {...field}
                      className={cn(styles.page__card__form__textarea,
                        error && styles.page__card__form__textarea__error)}
                      rows={6}
                    />
                    {error && (
                      <span className={styles.page__card__form__field__error}>{error}</span>
                    )}
                  </>
                );
              }}
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
