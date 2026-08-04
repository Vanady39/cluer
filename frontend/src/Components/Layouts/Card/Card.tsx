import styles from './Styles.module.scss';
import { memo } from 'react';
import { FiHeart, FiMapPin, FiMoreHorizontal, } from 'react-icons/fi';

interface CardProps {
  title: string;
  price: number;
  imageUrl: string;
  city: string;
  onClick?: () => void;
}

function CardComponent({
  title,
  price,
  imageUrl,
  city,
  onClick,
}: CardProps) {
  const formattedPrice = new Intl.NumberFormat('ru-RU').format(price);
  const location = city;

  return (
    <article
      className={styles.card}
      onClick={onClick}
    >
      <img
        className={styles.card__image}
        src={imageUrl}
        alt={title}
        loading="lazy"
      />

      <div className={styles.card__body}>
        <div className={styles.card__information}>
          <h3 className={styles.card__title}>
            {title}
          </h3>

          <p className={styles.card__price}>
            {formattedPrice} ₽
          </p>

          <p className={styles.card__location}>
            <FiMapPin className={styles.card__locationIcon} />
            <span>{location}</span>
          </p>
        </div>

        <div className={styles.card__actions}>
          <button
            type="button"
            className={styles.card__action}
            aria-label="Добавить в избранное"
            onClick={(event) => event.stopPropagation()}
          >
            <FiHeart />
          </button>

          <button
            type="button"
            className={styles.card__action}
            aria-label="Дополнительные действия"
            onClick={(event) => event.stopPropagation()}
          >
            <FiMoreHorizontal />
          </button>
        </div>
      </div>
    </article>
  );
}

export const Card = memo(CardComponent);