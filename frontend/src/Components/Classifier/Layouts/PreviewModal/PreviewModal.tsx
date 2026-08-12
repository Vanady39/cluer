import { memo } from "react";
import styles from './Styles.module.scss';
import { Button } from "../../../UI/Button";

interface PreviewModalProps {
  image: string | null;
  alt?: string;
  onClose: () => void;
}

function PreviewModalComponent({ image, alt, onClose }: PreviewModalProps) {
  if (!image) return null;

  return (
    <div className={styles.modal} onClick={onClose} role="dialog" aria-modal="true">
      <div className={styles.modal__content} onClick={(e) => e.stopPropagation()}>
        <Button size="min" className={styles.modal__close} onClick={onClose} aria-label="Закрыть">
          ✕
        </Button>
        <img src={image} alt={alt || "Превью"} />
      </div>
    </div>
  );
}

export const PreviewModal = memo(PreviewModalComponent);