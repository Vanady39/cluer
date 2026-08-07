import { memo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import useClickOutside from "../../../Hooks/useClickOutside";
import { categories, type CategoryKey } from "../../../types";
import { Button } from "../../UI/Button/Button";
import styles from "./Styles.module.scss";

interface CategoriesProps {
  value?: string;
  onChange?: (value: string) => void;
  label?: string;
  error?: string;
  isHeader?: boolean;
}

function CategoriesComponent({
  value = "",
  onChange,
  label,
  error,
  isHeader = false,
}: CategoriesProps) {
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const selectedLabel = categories.find(c => c.key === value || "")?.label;

  useClickOutside(ref, () => setIsOpen(false), ["button"], isOpen);

  const handleSelect = (key: CategoryKey) => {
    if (onChange) {
      onChange(key);
    } else {
      navigate(`/?category=${key}`);
    }
    setIsOpen(false);
  };

  if (isHeader) {
    return (
      <div ref={ref} className={styles.categories}>
        <Button
          color='primary'
          size="main"
          className={styles.categories__btn}
          onClick={() => setIsOpen(!isOpen)}
        >
          Все категории <span className={styles.categories__arrow}>▾</span>
        </Button>

        {isOpen && (
          <div className={styles.categories__menu}>
            {categories.map((category) => (
              <button
                key={category.key}
                className={styles.categories__item}
                onClick={() => handleSelect(category.key)}
              >
                {category.label}
              </button>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div ref={ref} className={styles.categories__field}>
      {label && <label className={styles.label}>{label}</label>}
      <button
        type="button"
        className={styles.categories__dropdownButton}
        onClick={() => setIsOpen(!isOpen)}
      >
        <span>{selectedLabel || "Выберите категорию"}</span>
        <span>▾</span>
      </button>

      {isOpen && (
        <div className={styles.categories__dropdown}>
          {categories.map((item) => (
            <button
              type="button"
              key={item.key}
              className={styles.categories__option}
              onClick={() => {
                onChange?.(item.key);
                setIsOpen(false);
              }}
            >
              {item.label}
            </button>
          ))}
        </div>
      )}
      {error && <span className={styles.categories__error}>{error}</span>}
    </div>
  );
}

export const Categories = memo(CategoriesComponent);