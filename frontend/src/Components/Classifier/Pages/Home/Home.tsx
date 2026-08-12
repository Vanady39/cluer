import styles from "./Styles.module.scss";
import { memo } from "react";
import { useSelector } from "react-redux";
import { useInfiniteQuery } from "@tanstack/react-query";
import { Card } from "../../Layouts/Card";
import { type RootState } from "../../../../Store/Store";
import { listingsAPI } from "../../../../Api/listings";
import { Button } from "../../../UI/Button";
import { PAGE_SIZE } from "../../../../Utils/constants";

function HomeComponent() {
  const searchQuery = useSelector(
    (state: RootState) => state.search.search,
  ).trim();

  const {
    data,
    isLoading,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["listings", searchQuery],
    initialPageParam: 0,

    queryFn: ({ pageParam }) =>
      listingsAPI.getAll({
        q: searchQuery || undefined,
        limit: PAGE_SIZE,
        offset: pageParam,
      }),

    getNextPageParam: (lastPage, allPages) =>
      lastPage.length < PAGE_SIZE ? undefined : allPages.length * PAGE_SIZE,
  });

  const listings = data?.pages.flat() ?? [];
  
  if (isLoading) return <div className={styles.home__loading}>Загрузка объявлений...</div>;
  if (error) return <div className={styles.home__error}>Ошибка загрузки объявлений</div>;

  return (
    <main className={styles.home}>
      <div className={styles.home__cards}>
        {listings.map((listing) => (
          <div key={listing.id} className={styles.home__card}>
            <Card
              title={listing.title}
              price={listing.price}
              imageUrl={listing.imageUrl}
              onClick={() => console.log(`Открыть объявление ${listing.id}`)}
            />
          </div>
        ))}
      </div>

      {listings.length === 0 && (
        <div className={styles.home__empty}>
          <p>Ничего не найдено</p>
        </div>
      )}

      {hasNextPage && (
        <div className={styles.home__more}>
          <Button
            type="button"
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? "Загрузка..." : "Показать ещё"}
          </Button>
        </div>
      )}
    </main>
  );
}

export const Home = memo(HomeComponent);
