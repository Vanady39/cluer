import { useMemo, useRef, useState, memo } from "react";
import html2pdf from "html2pdf.js";
import {
  useAllToursAnalytics,
  type TourAnalyticsItem,
} from "../../../../Hooks/useAllToursAnalytics";
import { percent, formatDate } from "../../../../Utils/format";
import styles from "./Styles.module.scss";
import { ScenarioMenu } from "./Components/ScenarioMenu/ScenarioMenu";
import { Button } from "../../../UI/Button";

export function Analytics() {
  const reportRef = useRef<HTMLDivElement | null>(null);
  const [isGeneratingPdf, setIsGeneratingPdf] = useState(false);
  const { data = [], isLoading, isError } = useAllToursAnalytics();
  const [selectedTourId, setSelectedTourId] = useState<string | null>(null);

  const successfulTours = useMemo(
    () => data.filter((item) => item.analytics),
    [data],
  );

  const totals = useMemo(() => {
    return successfulTours.reduce(
      (acc, item) => {
        const analytics = item.analytics;
        if (!analytics) return acc;
        acc.started += analytics.totals.started;
        acc.completed += analytics.totals.completed;
        acc.dismissed += analytics.totals.dismissed;
        acc.goalReached += analytics.totals.goal_reached;

        return acc;
      },
      {
        started: 0,
        completed: 0,
        dismissed: 0,
        goalReached: 0,
      },
    );
  }, [successfulTours]);

  const totalCompletionRate =
    totals.started > 0 ? totals.completed / totals.started : 0;
  const totalGoalRate =
    totals.started > 0 ? totals.goalReached / totals.started : 0;

  const selectedItem = useMemo(
    () => data.find((item) => item.tour.id === selectedTourId),
    [data, selectedTourId],
  );

  const downloadPdf = async () => {
    if (!reportRef.current || isGeneratingPdf) {
      return;
    }
    try {
      setIsGeneratingPdf(true);
      await new Promise<void>((resolve) => setTimeout(resolve, 0));
      await html2pdf()
        .set({
          margin: [10, 10, 10, 10],
          filename: `onboarding-analytics-${new Date().toISOString().slice(0, 10)}.pdf`,
          image: { type: "jpeg", quality: 0.98 },
          html2canvas: { scale: 2, backgroundColor: "#ffffff" },
          jsPDF: { unit: "mm", format: "a4", orientation: "landscape" },
        })
        .from(reportRef.current)
        .save();
    } catch (error) {
      console.error("[Analytics] PDF generation failed", error);
      alert("Не удалось сформировать PDF");
    } finally {
      setIsGeneratingPdf(false);
    }
  };

  if (isLoading) {
    return (
      <div className={styles.page}>
        <h1>Аналитика</h1>
        <div className={styles.page__loading}>
          Загружаем аналитику сценариев...
        </div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.page}>
        <h1>Аналитика</h1>
        <div className={styles.page__error}>
          Не удалось загрузить аналитику.
        </div>
      </div>
    );
  }

  return (
    <div className={styles.page}>
      <ScenarioMenu />
      <div ref={reportRef}>
        <header className={styles.page__header}>
          <div>
            <h1>Аналитика</h1>
            <p>Эффективность сценариев</p>
          </div>

          {!isGeneratingPdf && (
            <Button className={styles.page__pdfButton} onClick={downloadPdf}>
              Скачать PDF
            </Button>
          )}
        </header>

        <section className={styles.page__stats}>
          <MetricCard title="Сценариев" value={data.length} />
          <MetricCard title="Запусков" value={totals.started} />
          <MetricCard title="Завершений" value={totals.completed} />
          <MetricCard
            title="Completion rate"
            value={percent(totalCompletionRate)}
          />
          <MetricCard title="Достигли цели" value={totals.goalReached} />
          <MetricCard title="Goal rate" value={percent(totalGoalRate)} />
        </section>

        <section className={styles.page__section}>
          <div className={styles.page__sectionHeader}>
            <div>
              <h2>Эффективность сценариев</h2>
              <p>Нажмите на сценарий, чтобы посмотреть подробности</p>
            </div>
          </div>

          {data.length === 0 ? (
            <div className={styles.page__empty}>Сценариев пока нет</div>
          ) : (
            <div className={styles.page__tableWrapper}>
              <table className={styles.page__table}>
                <thead>
                  <tr>
                    <th>Сценарий</th>
                    <th>Версия</th>
                    <th>Запуски</th>
                    <th>Завершили</th>
                    <th>Completion</th>
                    <th>Цель</th>
                    <th>Goal rate</th>
                    <th>Закрыли</th>
                  </tr>
                </thead>
                <tbody>
                  {data.map((item) => {
                    const analytics = item.analytics;
                    const isSelected = selectedTourId === item.tour.id;
                    return (
                      <tr
                        key={item.tour.id}
                        onClick={() =>
                          setSelectedTourId(isSelected ? null : item.tour.id)
                        }
                        className={
                          isSelected ? styles.page__selectedRow : undefined
                        }
                      >
                        <td>
                          <div className={styles.page__tourName}>
                            <strong>{item.tour.title}</strong>
                            {item.tour.description && (
                              <span>{item.tour.description}</span>
                            )}
                          </div>
                        </td>

                        {!analytics ? (
                          <td colSpan={7} className={styles.page__noAnalytics}>
                            {item.unpublished
                              ? "Сценарий ещё не опубликован"
                              : "Аналитика недоступна"}
                          </td>
                        ) : (
                          <>
                            <td>v{analytics.version}</td>
                            <td>{analytics.totals.started}</td>
                            <td>{analytics.totals.completed}</td>
                            <td>
                              <strong>
                                {percent(analytics.totals.completion_rate)}
                              </strong>
                            </td>
                            <td>{analytics.totals.goal_reached}</td>
                            <td>{percent(analytics.totals.goal_rate)}</td>
                            <td>{analytics.totals.dismissed}</td>
                          </>
                        )}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </section>
        {selectedItem?.analytics && <TourDetails item={selectedItem} />}
      </div>
    </div>
  );
}

function MetricCard({
  title,
  value,
}: {
  title: string;
  value: string | number;
}) {
  return (
    <div className={styles.page__card}>
      <span>{title}</span>
      <strong>{value}</strong>
    </div>
  );
}

const TourDetails = memo(({ item }: { item: TourAnalyticsItem }) => {
  const analytics = item.analytics;
  if (!analytics) return null;

  return (
    <>
      <section className={styles.page__section}>
        <div className={styles.page__detailsHeader}>
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
          <div className={styles.page__versionId}>
            <span>Version ID</span>
            <code>{analytics.tour_version_id}</code>
          </div>
        </div>

        <h3>Аналитика</h3>
        {analytics.funnel.length === 0 ? (
          <div className={styles.page__empty}>
            По этому сценарию пока нет данных
          </div>
        ) : (
          <div className={styles.page__funnel}>
            {analytics.funnel.map((step) => (
              <div key={step.hint_id} className={styles.page__funnelRow}>
                <div className={styles.page__stepInfo}>
                  <div className={styles.page__stepNumber}>{step.step}</div>
                  <div>
                    <strong>{step.title}</strong>
                    <span>Drop-off: {percent(step.dropoff)}</span>
                  </div>
                </div>
                <div className={styles.page__stepStats}>
                  <StepMetric title="Показано" value={step.shown} />
                  <StepMetric title="Завершено" value={step.completed} />
                  <StepMetric title="Пропущено" value={step.skipped} />
                  <StepMetric
                    title="Selector missing"
                    value={step.selector_missing}
                  />
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      <section className={styles.page__section}>
        <h2>Проблемные селекторы</h2>
        {analytics.broken_selectors.length === 0 ? (
          <div className={styles.page__success}>
            Проблемных селекторов не обнаружено
          </div>
        ) : (
          <div className={styles.page__brokenSelectors}>
            <p>
              Найдено проблемных подсказок:{" "}
              <strong>{analytics.broken_selectors.length}</strong>
            </p>
            {analytics.broken_selectors.map((id) => {
              const step = analytics.funnel.find((item) => item.hint_id === id);

              return (
                <div key={id}>
                  {step ? (
                    <strong>
                      Шаг {step.step} — {step.title}
                    </strong>
                  ) : (
                    <code>{id}</code>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </section>
    </>
  );
});

function StepMetric({ title, value }: { title: string; value: number }) {
  return (
    <div>
      <span>{title}</span>
      <strong>{value}</strong>
    </div>
  );
}
