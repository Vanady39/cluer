import { useEffect, useState } from "react";
import styles from "./Styles.module.scss";
import type { TourHint } from "../../types/sdk";
import { Button } from "../UI/Button";

interface Props {
  hint: TourHint;
  element: HTMLElement | null;
  step: number;
  total: number;
  next: () => void;
  skip: () => void;
}

export function Hint({ hint, element, step, total, next, skip }: Props) {
  const [position, setPosition] = useState({
    top: 0,
    left: 0,
  });

  const [spotlightRect, setSpotlightRect] = useState<{
    top: number;
    left: number;
    width: number;
    height: number;
  } | null>(null);

  useEffect(() => {
    const updatePosition = () => {
      const width = 320;
      const height = 120;

      if (!element) return;
      const rect = element.getBoundingClientRect();
      let top: number;
      let left: number;

      switch (hint.placement) {
        case "top":
          top = rect.top - height - 12;
          left = rect.left;
          break;
        case "left":
          top = rect.top;
          left = rect.left - width - 12;
          break;
        case "right":
          top = rect.top;
          left = rect.right + 12;
          break;
        case "bottom":
        default:
          top = rect.bottom + 12;
          left = rect.left;
          break;
      }

      if (left + width > window.innerWidth - 20)
        left = window.innerWidth - width - 20;
      if (left < 20) left = 20;
      if (top + height > window.innerHeight - 20) top = rect.top - height - 12;
      if (top < 20) top = 20;

      setPosition({ top, left });
      if (hint.spotlight) {
        setSpotlightRect({
          top: rect.top - 8,
          left: rect.left - 8,
          width: rect.width + 16,
          height: rect.height + 16,
        });
      } else {
        setSpotlightRect(null);
      }
    };
    updatePosition();

    window.addEventListener("scroll", updatePosition);
    window.addEventListener("resize", updatePosition);
    return () => {
      window.removeEventListener("scroll", updatePosition);
      window.removeEventListener("resize", updatePosition);
    };
  }, [element, hint.placement, hint.spotlight]);

  return (
    <>
      {hint.spotlight && spotlightRect && (
        <div
          className={styles.spotlight}
          style={{
            top: spotlightRect.top,
            left: spotlightRect.left,
            width: spotlightRect.width,
            height: spotlightRect.height,
          }}
        />
      )}
      <div
        className={styles.spotlight__tooltip}
        style={{
          top: position.top,
          left: position.left,
        }}
      >
        <h3>{hint.title}</h3>
        <span>
          {step + 1}/{total}
        </span>
        <p>{hint.content}</p>
        {hint.media_url && (
          <img
            src={hint.media_url}
            alt=""
            className={styles.spotlight__media}
          />
        )}

        <div className={styles.spotlight__footer}>
          {step !== total - 1 && (
            <button className={styles.spotlight__skip} onClick={skip}>
              Пропустить
            </button>
          )}
          <Button onClick={next}>
            {step === total - 1 ? "Завершить" : "Далее"}
          </Button>
        </div>
      </div>
    </>
  );
}