import { memo } from "react";
import styles from './Styles.module.scss';

interface PreviewModalProps {
  image: string | null;
  onClose: () => void;
}

function PreviewModalComponent({ image, onClose }: PreviewModalProps) {
  if (!image) return null;

  return (
    <div className={styles.modal} onClick={onClose}>
      <div className={styles.modalContent} onClick={(e) => e.stopPropagation()}>
        <button className={styles.modalClose} onClick={onClose}>✕</button>
        <img src={image} alt="Preview" />
      </div>
    </div>
  );
}

export const PreviewModal = memo(PreviewModalComponent);