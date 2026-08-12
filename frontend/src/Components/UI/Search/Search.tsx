import { memo, useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaSearch } from "react-icons/fa";
import { setSearch, cleanSearch } from "../../../Reducers/searchReduce";
import { type RootState } from "../../../Store/Store";
import styles from "./Styles.module.scss";
import { SEARCH_DEBOUNCE_MS } from "../../../Utils/constants";

function SearchBarComponent() {
  const dispatch = useDispatch();
  const searchQuery = useSelector((state: RootState) => state.search.search);
  const [localSearch, setLocalSearch] = useState(searchQuery || "");

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      dispatch(setSearch(localSearch));
    }, SEARCH_DEBOUNCE_MS);

    return () => window.clearTimeout(timeoutId);
  }, [dispatch, localSearch]);

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") dispatch(setSearch(localSearch));
  };

  const handleClear = () => {
    setLocalSearch("");
    dispatch(cleanSearch());
  };

  return (
    <div className={styles.search}>
      <div className={styles.search__wrapper}>
        <FaSearch className={styles.search__icon} />

        <input
          type="text"
          placeholder="Поиск..."
          value={localSearch}
          onChange={(event) => setLocalSearch(event.target.value)}
          onKeyDown={handleKeyDown}
          className={styles.search__input}
        />

        {localSearch && (
          <button
            type="button"
            className={styles.search__clear}
            onClick={handleClear}
            aria-label="Очистить поиск"
          >
            ✕
          </button>
        )}
      </div>
    </div>
  );
}

export const SearchBar = memo(SearchBarComponent);