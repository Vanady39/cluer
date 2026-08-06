import { useEffect, useState } from "react";
import styles from "./Styles.module.scss";
import type { Hint as HintType } from "../../types/sdk";

interface Props {
  hint: HintType;
  element: HTMLElement;
  step: number;
  total: number;
  next: () => void;
  skip: () => void;
}

export function Hint({ hint, element, step, total, next, skip }: Props) {
  const [position, setPosition] = useState({ top: 0, left: 0 });

  useEffect(() => {
    const updatePosition = () => {
      const rect = element.getBoundingClientRect();
      setPosition({ top: rect.bottom + 12, left: rect.left });
    };
    updatePosition();
    window.addEventListener("scroll", updatePosition);
    window.addEventListener("resize", updatePosition);
    return () => {
      window.removeEventListener("scroll", updatePosition);
      window.removeEventListener("resize", updatePosition);
    };
  }, [element]);

  const rect = element.getBoundingClientRect();

  return (
    <>
      {hint.spotlight && (
        <div
          className={styles.spotlight}
          style={{
            top: rect.top - 8,
            left: rect.left - 8,
            width: rect.width + 16,
            height: rect.height + 16,
          }}
        />
      )}
      <div
        className={styles.tooltip}
        style={{ top: position.top, left: position.left }}
      >
        {/* Оборачиваем заголовок и счетчик в header */}
        <div className={styles.header}>
          <h3>{hint.title}</h3>
          <span className={styles.stepCounter}>
            {step + 1}/{total}
          </span>
        </div>

        <p>{hint.content}</p>

        <div className={styles.footer}>
          {step === total - 1 ? (
            <button className={styles.next} onClick={next}>
              Завершить
            </button>
          ) : (
            <button className={styles.skip} onClick={skip}>
              Пропустить
            </button>
          )}
        </div>
      </div>
    </>
  );
}
