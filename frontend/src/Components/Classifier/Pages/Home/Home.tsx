import styles from "./Styles.module.scss";
import { memo, useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { useSelector } from "react-redux";
import { useQuery } from "@tanstack/react-query";
import { Card } from "../../Layouts/Card";
import { type RootState } from "../../../../Store/Store";
import { listingsAPI } from "../../../../Api/listings";

function HomeComponent() {
  const [searchParams] = useSearchParams();
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const searchQuery = useSelector((state: RootState) => state.search.search);

  const {
    data: listings = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ["listings"],
    queryFn: listingsAPI.getAll,
  });

  useEffect(() => {
    const category = searchParams.get("category") || "all";
    if (category !== selectedCategory) {
      setSelectedCategory(category);
    }
  }, [searchParams, selectedCategory]);

  const filteredListings = listings.filter((item) => {
    if (selectedCategory !== "all" && item.category !== selectedCategory) {
      return false;
    }
    if (
      searchQuery.trim() &&
      !item.title.toLowerCase().includes(searchQuery.toLowerCase().trim())
    ) {
      return false;
    }
    return true;
  });

  if (isLoading) {
    return <div className={styles.home__loading}>Загрузка объявлений...</div>;
  }

  if (error) {
    return <div className={styles.home__error}>Ошибка загрузки объявлений</div>;
  }

  return (
    <main className={styles.home}>
      <div className={styles.home__cards}>
        {filteredListings.map((listing) => (
          <div key={listing.id} className={styles.home__card}>
            <Card
              title={listing.title}
              price={listing.price}
              imageUrl={listing.imageUrl}
              city={listing.city || "Город не указан"}
              onClick={() => console.log(`Открыть объявление ${listing.id}`)}
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
