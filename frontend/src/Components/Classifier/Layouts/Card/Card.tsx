import styles from './Styles.module.scss';
import { memo } from 'react';
import { FiMapPin } from 'react-icons/fi';

const priceFormatter = new Intl.NumberFormat('ru-RU');

interface CardProps {
  title: string;
  price: number;
  imageUrl: string;
  city: string;
  onClick?: () => void;
}

function CardComponent({ title, price, imageUrl, city, onClick }: CardProps) {
  return (
    <article className={styles.card} onClick={onClick}>
      <img
        className={styles.card__image}
        src={imageUrl}
        alt={title || "Карточка товара"}
        loading="lazy"
      />
      <div className={styles.card__body}>
        <div className={styles.card__information}>
          <h3 className={styles.card__title}>{title}</h3>
          <p className={styles.card__price}>
            {priceFormatter.format(price)} ₽
          </p>
          <p className={styles.card__location}>
            <FiMapPin className={styles.card__locationIcon} />
            <span>{city}</span>
          </p>
        </div>
      </div>
    </article>
  );
}

export const Card = memo(CardComponent);