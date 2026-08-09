import { memo, useState } from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import styles from "./Styles.module.scss";
import { Button } from "../../../UI/Button";
import { ScenarioMenu } from "../Analytics/Components/ScenarioMenu/ScenarioMenu";
import { useToursQuery } from "../../../../Hooks/useToursQuery";
import { useDeleteTourMutation } from "../../../../Hooks/useDeleteTourMutation";
import { onboardingAPI } from "../../../../Api/onboarding";
import type { TourVersion } from "../../../../types/sdk";
import { TourVersionsModal } from "./Components/TourVersionModal/TourVersionModal";
import logo from "/logo.svg";
import { ToggleSwitch } from "../../../UI/ToggleSwitch/ToggleSwitch";

function ScenariosComponent() {
  const navigate = useNavigate();
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [updatingEnabledId, setUpdatingEnabledId] = useState<string | null>(null);
  const [showVersions, setShowVersions] = useState<{ id: string; title: string } | null>(null);
  const { data: scenarios = [], isLoading, error, refetch } = useToursQuery();
  const deleteMutation = useDeleteTourMutation();

  const handleDelete = (id: string) => {
    deleteMutation.mutate(id);
    setOpenMenu(null);
  };

  const handleToggleEnabled = async (id: string, enabled: boolean) => {
    try {
      setUpdatingEnabledId(id);
      await onboardingAPI.updateTourMeta(id, { enabled });
      await refetch();
    } catch (error) {
      console.error("TOGGLE ENABLED ERROR", error);
      alert("Не удалось изменить состояние сценария");
    } finally {
      setUpdatingEnabledId(null);
    }
  };

  const handleEdit = async (id: string) => {
    try {
      const tour = await onboardingAPI.getTour(id);
      if (tour.published && !tour.draft) {
        await onboardingAPI.createDraft(id);
      }
      navigate(`/admin/scenarios/create?id=${id}`);
    } catch (error) {
      console.error("EDIT TOUR ERROR", error);
      alert("Не удалось открыть сценарий");
    }
  };

  const getTourStatus = (tour: { draft?: TourVersion; published?: TourVersion }) => {
    const hasDraft = Boolean(tour.draft);
    const hasPublished = Boolean(tour.published);

    if (hasPublished && hasDraft) return { label: "Есть неопубликованные изменения", type: "changes" };
    if (hasPublished) return { label: "Опубликован", type: "published" };
    if (hasDraft) return { label: "Черновик", type: "draft" };
    return { label: "Без версии", type: "unknown" };
  };

  const handlePreview = async (tourId: string) => {
    try {
      const card = await onboardingAPI.getTour(tourId);
      const version = card.draft ?? card.published;
      if (!version) {
        alert("У сценария пока нет версии для предпросмотра");
        return;
      }
      const previewUrl = new URL(version.target_path || "/", window.location.origin);
      previewUrl.searchParams.set("preview", "true");
      previewUrl.searchParams.set("tourId", tourId);
      window.open(previewUrl.toString(), "_blank", "noopener,noreferrer");
    } catch (error) {
      console.error("[Admin] preview error", error);
      alert("Не удалось открыть предпросмотр");
    }
  };

  if (isLoading) return <div>Загрузка сценариев...</div>;
  if (error) return <div>Ошибка загрузки сценариев: {error.message}</div>;

  return (
    <div className={styles.page}>
      <ScenarioMenu />
      <div className={styles.page__header}>
        <img src={logo} className={styles.page__header__logo} />
        <h1>Сценарии</h1>
        <Button
          className={styles.page__header__createButton}
          size="main"
          color="primary"
          onClick={() => navigate("/admin/scenarios/create")}
        >
          + Создать
        </Button>
      </div>

      <div className={styles.page__table}>
        <div className={styles.page__table__rowHeader}>
          <span>Название</span>
          <span>Статус</span>
          <span>Обновлен</span>
          <span>Шагов</span>
          <span>Показ сценария пользователям</span>
          <span>Действия</span>
        </div>
        {scenarios.map((item) => {
          const status = getTourStatus(item);

          return (
            <div className={styles.page__table__row} key={item.id}>
              <span className={styles.page__table__row__name}>{item.title}</span>

              <span>
                <span
                  className={cn(
                    styles.page__table__row__status,
                    status.type === "published"
                      ? styles.page__table__row__status__success
                      : styles.page__table__row__status__warning
                  )}
                />
                {status.label}
              </span>

              <span>{item.updated_at ? new Date(item.updated_at).toLocaleDateString() : "—"}</span>

              <span>{item.hints?.length || 0}</span>

              <span className={styles.page__table__row__toggleCell}>
                <ToggleSwitch
                  checked={Boolean(item.enabled)}
                  disabled={updatingEnabledId === item.id}
                  onChange={(checked) => handleToggleEnabled(item.id, checked)}
                />
              </span>

              <div className={styles.page__table__row__actions}>
                <Button size="min" onClick={() => handlePreview(item.id)}>
                  Предпросмотр
                </Button>
                <div className={styles.page__table__row__menuWrapper}>
                  <Button
                    size="min"
                    className={styles.page__table__row__menuWrapper__more}
                    onClick={() => setOpenMenu(openMenu === item.id ? null : item.id)}
                  >
                    ⋮
                  </Button>
                  {openMenu === item.id && (
                    <div className={styles.page__table__row__menuWrapper__dropdown}>
                      <button
                        type="button"
                        className={styles.page__table__row__menuWrapper__dropdown__menuButton}
                        onClick={() => setShowVersions({ id: item.id, title: item.title })}
                      >
                        История версий
                      </button>
                      <button
                        type="button"
                        className={styles.page__table__row__menuWrapper__dropdown__menuButton}
                        onClick={() => handleEdit(item.id)}
                      >
                        Редактировать
                      </button>
                      <button
                        type="button"
                        className={styles.page__table__row__menuWrapper__dropdown__deleteButton}
                        onClick={() => handleDelete(item.id)}
                      >
                        Удалить
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {showVersions && (
        <TourVersionsModal
          tourId={showVersions.id}
          title={showVersions.title}
          onClose={() => setShowVersions(null)}
        />
      )}
    </div>
  );
}

export const Scenarios = memo(ScenariosComponent);