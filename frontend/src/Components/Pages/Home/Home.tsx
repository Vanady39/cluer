import styles from './Styles.module.scss';
import { memo } from 'react';
import { Card } from '../../Layouts/Card/Card';

const listings = [
  {
    id: 1,
    title: 'Британский короткошёрстный котёнок',
    price: 15000,
    imageUrl: 'https://30.img.avito.st/image/1/1.dzt_6ra429JJXVnfHa81GWhK2dTBS1nESUbZ0M9D09jJ.gHMiadOthtEO5UrcRemha1NtkibX6DgvW99UgG_cUss?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
  {
    id: 2,
    title: 'HAVAL Jolion 1.5 AMT, 2022, 92 000 км',
    price: 1172000,
    imageUrl: 'https://30.img.avito.st/image/1/1.5ZXVHLa4SXzjq8txkxfhzcK8S3prvctq47BLfmW1QXZj.GLRcWhh7xrLntc3Awv3HJ5wTSDJUNPXw7v7XQpOAPPA?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
  {
    id: 3,
    title: 'Британский короткошёрстный котёнок',
    price: 15000,
    imageUrl: 'https://30.img.avito.st/image/1/1.dzt_6ra429JJXVnfHa81GWhK2dTBS1nESUbZ0M9D09jJ.gHMiadOthtEO5UrcRemha1NtkibX6DgvW99UgG_cUss?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
  {
    id: 4,
    title: 'HAVAL Jolion 1.5 AMT, 2022, 92 000 км',
    price: 1172000,
    imageUrl: 'https://30.img.avito.st/image/1/1.5ZXVHLa4SXzjq8txkxfhzcK8S3prvctq47BLfmW1QXZj.GLRcWhh7xrLntc3Awv3HJ5wTSDJUNPXw7v7XQpOAPPA?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
  {
    id: 5,
    title: 'Британский короткошёрстный котёнок',
    price: 15000,
    imageUrl: 'https://30.img.avito.st/image/1/1.dzt_6ra429JJXVnfHa81GWhK2dTBS1nESUbZ0M9D09jJ.gHMiadOthtEO5UrcRemha1NtkibX6DgvW99UgG_cUss?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
  {
    id: 6,
    title: 'HAVAL Jolion 1.5 AMT, 2022, 92 000 км',
    price: 1172000,
    imageUrl: 'https://30.img.avito.st/image/1/1.5ZXVHLa4SXzjq8txkxfhzcK8S3prvctq47BLfmW1QXZj.GLRcWhh7xrLntc3Awv3HJ5wTSDJUNPXw7v7XQpOAPPA?cqp=2.TSzMy-m0u9ojo94xoNTr4TIkcUBjMu1L_y5Z6Lr-VHA-3xfoRpsvf1jN4IBF36LEO13sxOCjYv9KRoXzpmAFvKTQ',
    city: 'Екатеринбург',
  },
];

function HomeComponent() {
  return (
    <main className={styles.home}>
      <div className={styles.home__cards}>
        {listings.map((listing) => (
          <div key={listing.id}>
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
    </main>
  );
}

export const Home = memo(HomeComponent);