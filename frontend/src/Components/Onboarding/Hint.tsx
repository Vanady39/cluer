import { useEffect, useState } from "react";
import styles from "./Styles.module.scss";
import type { TourHint } from "../../types/sdk";
import { Button } from "../UI/Button";

interface Props {
  hint: TourHint;
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
        <div className={styles.header}>
          <h3>{hint.title}</h3>
          <span className={styles.stepCounter}>
            {step + 1}/{total}
          </span>
        </div>

        <p>{hint.content}</p>

        <div className={styles.footer}>
          <button className={styles.skip} onClick={skip}>
            Пропустить
          </button>

          <Button onClick={next}>
            {step === total - 1 ? "Закрыть" : "Далее"}
          </Button>
        </div>
      </div>
    </>
  );
}