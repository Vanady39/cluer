import { memo, useRef, useState } from "react";
import { FaLocationArrow } from "react-icons/fa";
import useClickOutside from "../../../Hooks/useClickOutside";
import { cities, type CityValue } from "../../../types";
import styles from "./Styles.module.scss";

interface CitySelectProps {
  value?: string;
  onChange?: (value: string) => void;
  label?: string;
  error?: string;
  isHeader?: boolean;
}

function CitySelectComponent({
  value = "",
  onChange,
  label,
  error,
  isHeader = false,
}: CitySelectProps) {
  const [selectedCity, setSelectedCity] = useState<CityValue | undefined>(value as CityValue | undefined);
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const currentValue = isHeader ? (selectedCity || value) : value;
  const selectedCityLabel = cities.find(c => c.value === currentValue)?.label;

  useClickOutside(ref, () => setIsOpen(false), ["button"], isOpen);

  const handleSelect = (cityValue: CityValue) => {
    if (isHeader) {
      setSelectedCity(cityValue);
    } else {
      onChange?.(cityValue);
    }
    setIsOpen(false);
  };

  if (isHeader) {
    return (
      <div ref={ref} className={styles.city}>
        <div className={styles.city__wrapper} onClick={() => setIsOpen(!isOpen)}>
          <FaLocationArrow className={styles.city__icon} />
          <span className={styles.city__label}>
            {selectedCityLabel || "Город"}
          </span>
        </div>

        {isOpen && (
          <div className={styles.city__menu}>
            {cities.map((city) => (
              <button
                key={city.value}
                className={styles.city__item}
                onClick={() => handleSelect(city.value)}
              >
                {city.label}
              </button>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div ref={ref} className={styles.city__field}>
      {label && <label className={styles.city__label}>{label}</label>}
      <button
        type="button"
        className={styles.city__dropdownButton}
        onClick={() => setIsOpen(!isOpen)}
      >
        <span>{selectedCityLabel || "Выберите город"}</span>
        <span>▾</span>
      </button>

      {isOpen && (
        <div className={styles.city__dropdown}>
          {cities.map((city) => (
            <button
              type="button"
              key={city.value}
              className={styles.city__option}
              onClick={() => handleSelect(city.value)}
            >
              {city.label}
            </button>
          ))}
        </div>
      )}
      {error && <span className={styles.city__error}>{error}</span>}
    </div>
  );
}

export const CitySelect = memo(CitySelectComponent);