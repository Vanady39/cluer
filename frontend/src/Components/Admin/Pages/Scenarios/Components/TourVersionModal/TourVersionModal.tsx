import { memo, useEffect, useRef } from "react";
import cn from "classnames";
import styles from './Styles.module.scss';
import type { TourVersion } from "../../../../../../types/sdk";
import { useTourVersions } from "../../../../../../Hooks/useTourVersion";

interface TourVersionsModalProps {
  tourId: string;
  title: string;
  onClose: () => void;
}

function TourVersionsModalComponent({ tourId, title, onClose }: TourVersionsModalProps) {
  const { versions, versionsLoading, rollbackVersionId, loadVersions, rollback } = useTourVersions();
  const hasLoaded = useRef(false);

  useEffect(() => {
    if (!hasLoaded.current) {
      loadVersions(tourId, title);
      hasLoaded.current = true;
    }
  }, [tourId, title]);

  const formatDate = (value?: string | null) => {
    if (!value) return "—";
    return new Date(value).toLocaleString("ru-RU");
  };

  const getStatusLabel = (status: TourVersion["status"]) => {
    switch (status) {
      case "published": return "Опубликована";
      case "draft": return "Черновик";
      case "archived": return "Архивная";
      default: return status;
    }
  };

  const handleClose = () => {
    if (rollbackVersionId) return;
    onClose();
  };

  return (
    <div
      className={styles.modalOverlay}
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          handleClose();
        }
      }}
    >
      <div className={styles.modalOverlay__versionsModal}>
        <div className={styles.modalOverlay__modalHeader}>
          <div>
            <h2>История версий</h2>
            <p>{title}</p>
          </div>
          <button
            type="button"
            className={styles.modalOverlay__closeButton}
            disabled={Boolean(rollbackVersionId)}
            onClick={handleClose}
          >
            ×
          </button>
        </div>

        {versionsLoading ? (
          <div className={styles.modalOverlay__versionsLoading}>Загрузка истории...</div>
        ) : versions.length === 0 ? (
          <div className={styles.modalOverlay__versionsEmpty}>Версий пока нет</div>
        ) : (
          <div className={styles.modalOverlay__versionsList}>
            {versions.map((version) => {
              const isPublished = version.status === "published";

              return (
                <div key={version.id} className={styles.modalOverlay__versionRow}>
                  <div className={styles.modalOverlay__versionInfo}>
                    <div className={styles.modalOverlay__versionTitle}>
                      <strong>v{version.version}</strong>
                      <span
                        className={cn(
                          styles.modalOverlay__versionStatus,
                          isPublished
                            ? styles.modalOverlay__versionPublished
                            : version.status === "draft"
                              ? styles.modalOverlay__versionDraft
                              : styles.modalOverlay__versionArchived
                        )}
                      >
                        {getStatusLabel(version.status)}
                      </span>
                    </div>
                    <span className={styles.modalOverlay__versionDate}>
                      {formatDate(version.published_at ?? version.created_at)}
                    </span>
                    <code className={styles.modalOverlay__versionId}>{version.id}</code>
                  </div>

                  {version.status === "archived" && (
                    <button
                      type="button"
                      className={styles.modalOverlay__rollbackButton}
                      disabled={Boolean(rollbackVersionId)}
                      onClick={() => rollback(version)}
                    >
                      {rollbackVersionId === version.id ? "Откат..." : "Откатить"}
                    </button>
                  )}
                  {isPublished && (
                    <span className={styles.modalOverlay__currentVersion}>Текущая</span>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}

export const TourVersionsModal = memo(TourVersionsModalComponent);