import { memo, useRef } from "react";
import styles from "./Styles.module.scss";

interface FileItem {
  uid: string;
  name: string;
  url?: string;
  file?: File;
  size?: number;
}

interface PhotoUploadProps {
  fileList: FileItem[];
  onFileChange: (files: FileItem[]) => void;
  openPreview: (url: string) => void;
  [key: `data-${string}`]: unknown;
}

function UploadComponent({
  fileList,
  onFileChange,
  openPreview,
  ...props
}: PhotoUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;

    const newFiles: FileItem[] = [];
    for (const file of files) {
      if (fileList.length + newFiles.length >= 5) {
        alert("Можно загрузить не более 5 фото");
        break;
      }
      if (!file.type.startsWith("image/")) continue;
      if (file.size > 5 * 1024 * 1024) {
        alert(`Файл ${file.name} превышает 5MB`);
        continue;
      }
      newFiles.push({
        uid: `${Date.now()}-${file.name}`,
        name: file.name,
        url: URL.createObjectURL(file),
        file,
        size: file.size,
      });
    }
    onFileChange([...fileList, ...newFiles]);
    e.target.value = "";
  };

  const removeFile = (uid: string) => {
    onFileChange(fileList.filter((f) => f.uid !== uid));
  };

  return (
    <div className={styles.field} {...props}>
      <label className={styles.label}>Фотографии</label>
      <div className={styles.uploadArea}>
        <div className={styles.uploadArea__previewList}>
          {fileList.map((file) => (
            <div key={file.uid} className={styles.uploadArea__previewItem}>
              <img src={file.url} alt={file.name} onClick={() => openPreview(file.url!)} />
              <button type="button" className={styles.uploadArea__removeBtn} onClick={() => removeFile(file.uid)}>
                ✕
              </button>
            </div>
          ))}
          {fileList.length < 5 && (
            <button type="button" className={styles.uploadArea__uploadBox} onClick={() => fileInputRef.current?.click()}>
              <span>+</span>
              <span>Загрузить</span>
            </button>
          )}
        </div>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png"
          multiple
          onChange={handleFileChange}
          className={styles.uploadArea__fileInput}
        />
        <span className={styles.uploadArea__uploadHint}>Максимум 5 фото, JPG/PNG до 5MB</span>
      </div>
    </div>
  );
}

export const Upload = memo(UploadComponent);