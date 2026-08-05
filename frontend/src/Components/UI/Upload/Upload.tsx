import { memo, useRef, useState } from "react";
import styles from './Styles.module.scss';
import cn from "classnames";

interface FileItem {
  uid: string;
  name: string;
  url?: string;
  file?: File;
}

interface UploadProps {
  value?: FileItem[];
  onChange?: (files: FileItem[]) => void;
  maxCount?: number;
  accept?: string;
  label?: string;
  error?: string;
  className?: string;
  disabled?: boolean;
}

function UploadComponent({
  value = [],
  onChange,
  maxCount = 5,
  accept = "image/jpeg,image/png",
  label,
  error,
  className = "",
  disabled = false,
}: UploadProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [errorMessage, setErrorMessage] = useState<string>("");

  const handleClick = () => {
    if (!disabled) {
      inputRef.current?.click();
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;

    const newFiles: FileItem[] = [];
    let hasError = false;

    Array.from(files).forEach((file) => {
      if (!accept.split(",").includes(file.type)) {
        setErrorMessage(`Файл "${file.name}" имеет недопустимый формат`);
        hasError = true;
        return;
      }

      if (value.length + newFiles.length >= maxCount) {
        setErrorMessage(`Можно загрузить не более ${maxCount} файлов`);
        hasError = true;
        return;
      }

      newFiles.push({
        uid: `${Date.now()}-${file.name}`,
        name: file.name,
        url: URL.createObjectURL(file),
        file,
      });
    });

    if (hasError) {
      setTimeout(() => setErrorMessage(""), 3000);
      return;
    }

    onChange?.([...value, ...newFiles]);
    e.target.value = "";
    setErrorMessage("");
  };

  const handleRemove = (uid: string) => {
    onChange?.(value.filter((f) => f.uid !== uid));
  };

  return (
    <div className={cn(styles.upload, className)}>
      {label && <label className={styles.upload__label}>{label}</label>}
      <div
        className={cn(styles.upload__dropzone, {
          [styles['upload__dropzone--error']]: error || errorMessage,
          [styles['upload__dropzone--disabled']]: disabled,
        })}
      >
        {value.length > 0 && (
          <div className={styles.upload__previewList}>
            {value.map((file) => (
              <div key={file.uid} className={styles.upload__previewItem}>
                <img
                  src={file.url}
                  alt={file.name}
                  className={styles.upload__previewImage}
                />
                <span className={styles.upload__previewName}>{file.name}</span>
                {!disabled && (
                  <button
                    type="button"
                    className={styles.upload__previewRemove}
                    onClick={() => handleRemove(file.uid)}
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
        {value.length < maxCount && !disabled && (
          <div className={styles.upload__area} onClick={handleClick}>
            <div className={styles.upload__icon}>📸</div>
            <div className={styles.upload__text}>Загрузить фото</div>
            <div className={styles.upload__hint}>
              {accept.split(",").map((type) => type.split("/")[1]).join(", ")}
            </div>
          </div>
        )}
        {value.length > 0 && (
          <div className={styles.upload__counter}>
            {value.length} / {maxCount}
          </div>
        )}
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple
          onChange={handleChange}
          className={styles.upload__input}
          disabled={disabled}
        />
      </div>
      {(error || errorMessage) && (
        <span className={styles.upload__error}>{error || errorMessage}</span>
      )}
    </div>
  );
}

export const Upload = memo(UploadComponent);