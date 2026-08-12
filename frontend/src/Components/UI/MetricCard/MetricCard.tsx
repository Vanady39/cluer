import { memo } from "react";
import styles from "./Styles.module.scss";

interface MetricCardProps {
  title: string;
  value: string | number;
}

function MetricCardComponent({ title, value }: MetricCardProps) {
  return (
    <div className={styles.card}>
      <span>{title}</span>
      <strong>{value}</strong>
    </div>
  );
}

export const MetricCard = memo(MetricCardComponent);