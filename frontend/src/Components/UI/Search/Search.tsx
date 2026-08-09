import { memo, useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { FaSearch } from "react-icons/fa";
import { setSearch, cleanSearch } from "../../../Reducers/searchReduce";
import { type RootState } from "../../../Store/Store";
import styles from "./Styles.module.scss";

function SearchBarComponent() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const searchQuery = useSelector((state: RootState) => state.search.search);
  const [localSearch, setLocalSearch] = useState(searchQuery || "");

  const handleFind = () => {
    if (localSearch.trim()) {
      dispatch(setSearch(localSearch.trim()));
      navigate(`/search?q=${encodeURIComponent(localSearch.trim())}`);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && localSearch.trim()) handleFind();
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
          onChange={(e) => {
            setLocalSearch(e.target.value);
            dispatch(setSearch(e.target.value));
          }}
          onKeyDown={handleKeyDown}
          className={styles.search__input}
        />
        {localSearch && (
          <button
            type="button"
            className={styles.search__clear}
            onClick={handleClear}
          >
            ✕
          </button>
        )}
      </div>
    </div>
  );
}

export const SearchBar = memo(SearchBarComponent);