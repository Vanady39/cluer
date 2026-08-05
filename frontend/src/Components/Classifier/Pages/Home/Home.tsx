import styles from './Styles.module.scss';
import { memo, useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useSelector } from 'react-redux'; 
import { Card } from '../../Layouts/Card';
import { type RootState } from '../../../../Store/Store';

const listings = [
  {
    id: 1,
    title: 'Британский короткошёрстный котёнок',
    price: 15000,
    imageUrl: 'https://30.img.avito.st/image/1/1.dzt_6ra429JJXVnfHa81GWhK2dTBS1nESUbZ0M9D09jJ.gHMiadOthtEO5UrcRemha1NtkibX6DgvW99UgG_cUss?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
    category: 'animals',
  },
  {
    id: 2,
    title: 'HAVAL Jolion 1.5 AMT, 2022, 92 000 км',
    price: 1172000,
    imageUrl: 'https://30.img.avito.st/image/1/1.5ZXVHLa4SXzjq8txkxfhzcK8S3prvctq47BLfmW1QXZj.GLRcWhh7xrLntc3Awv3HJ5wTSDJUNPXw7v7XQpOAPPA?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
    category: 'transport',
  },
];

function HomeComponent() {
  const [searchParams] = useSearchParams();
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  
  // 👇 Берем поиск из Redux
  const searchQuery = useSelector((state: RootState) => state.search.search);

  useEffect(() => {
    const category = searchParams.get('category') || 'all';
    setSelectedCategory(category);
  }, [searchParams]);

  // 👇 Фильтруем по категории И по поисковому запросу
  const filteredListings = listings.filter((item) => {
    // Фильтр по категории
    if (selectedCategory !== 'all' && item.category !== selectedCategory) {
      return false;
    }
    // Фильтр по поиску (поиск по названию)
    if (searchQuery.trim() && !item.title.toLowerCase().includes(searchQuery.toLowerCase().trim())) {
      return false;
    }
    return true;
  });

  return (
    <main className={styles.home}>
      <div className={styles.home__cards}>
        {filteredListings.map((listing) => (
          <div key={listing.id} className={styles.home__card}>
            <Card
              title={listing.title}
              price={listing.price}
              imageUrl={listing.imageUrl}
              city={listing.city}
              onClick={() => {
                console.log(`Открыть объявление ${listing.id}`);
              }}
            />
          </div>
        ))}
      </div>

      {filteredListings.length === 0 && (
        <div className={styles.home__empty}>
          <p>Ничего не найдено</p>
        </div>
      )}
    </main>
  );
}

export const Home = memo(HomeComponent);