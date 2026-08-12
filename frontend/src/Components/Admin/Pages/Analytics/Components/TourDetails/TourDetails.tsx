import { memo } from "react";
import { percent, formatDate } from "../../../../../../Utils/format";
import type { TourAnalyticsItem } from "../../../../../../types";
import styles from "./Styles.module.scss";

interface TourDetailsProps {
  item: TourAnalyticsItem;
}

function TourDetailsComponent({ item }: TourDetailsProps) {
  const analytics = item.analytics;
  if (!analytics) return null;

  return (
    <>
      <section className={styles.section}>
        <div className={styles.section__detailsHeader}>
          <div>
            <h2>{item.tour.title}</h2>
            <p>
              Версия {analytics.version}
              {" · "}
              {formatDate(analytics.from)}
              {" — "}
              {formatDate(analytics.to)}
            </p>
          </div>
          <div className={styles.section__versionId}>
            <span>Version ID</span>
            <code>{analytics.tour_version_id}</code>
          </div>
        </div>

        <h3>Аналитика</h3>
        {analytics.funnel.length === 0 ? (
          <div className={styles.section__empty}>По этому сценарию пока нет данных</div>
        ) : (
          <div className={styles.section__funnel}>
            {analytics.funnel.map((step) => (
              <div key={step.hint_id} className={styles.section__funnelRow}>
                <div className={styles.section__stepInfo}>
                  <div className={styles.section__stepNumber}>{step.step}</div>
                  <div>
                    <strong>{step.title}</strong>
                    <span>Drop-off: {percent(step.dropoff)}</span>
                  </div>
                </div>
                <div className={styles.section__stepStats}>
                  <StepMetric title="Показано" value={step.shown} />
                  <StepMetric title="Завершено" value={step.completed} />
                  <StepMetric title="Пропущено" value={step.skipped} />
                  <StepMetric title="Selector missing" value={step.selector_missing} />
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section className={styles.section}>
        <h2>Проблемные селекторы</h2>
        {analytics.broken_selectors.length === 0 ? (
          <div className={styles.section__success}>Проблемных селекторов не обнаружено</div>
        ) : (
          <div className={styles.section__brokenSelectors}>
            <p>
              Найдено проблемных подсказок:{" "}
              <strong>{analytics.broken_selectors.length}</strong>
            </p>
            {analytics.broken_selectors.map((id) => {
              const step = analytics.funnel.find((item) => item.hint_id === id);
              return (
                <div key={id}>
                  {step ? (<strong>Шаг {step.step} — {step.title}</strong>) : (<code>{id}</code>)}
                </div>
              );
            })}
          </div>
        )}
      </section>
    </>
  );
}

function StepMetric({ title, value }: { title: string; value: number }) {
  return (
    <div>
      <span>{title}</span>
      <strong>{value}</strong>
    </div>
  );
}

export const TourDetails = memo(TourDetailsComponent);