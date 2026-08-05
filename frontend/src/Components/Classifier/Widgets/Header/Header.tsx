import styles from "./Styles.module.scss";
import { memo } from "react";
import { Link, useLocation } from "react-router-dom";
import logo from "/logo.svg";
import { Categories } from "../../../UI/CategoriesSelect/CategoriesSelect";
import { SearchBar } from "../../../UI/Search/Search";
import { CitySelect } from "../../../UI/CitySelect/CitySelect";
import { UserMenu } from "./Components/UserMenu/UserMenu";
import { Button } from "../../../UI/Button";

function HeaderComponent() {
  const location = useLocation();
  const isCreateAdPage = location.pathname === '/addItem';

  return (
    <header className={styles.header}>
      <div className={styles.header__left}>
        <Link to={"/"}>
          <img src={logo} className={styles.header__logo} alt="Logo" />
        </Link>
        {!isCreateAdPage && <Categories isHeader />}
      </div>

      {!isCreateAdPage && <SearchBar />}

      <div className={styles.header__right}>
        {!isCreateAdPage && (
          <>
            <Link to={"/addItem"}>
              <Button className={styles.header__button_add} color="transparent" size='nav'>
                Разместить объявление
              </Button>
            </Link>
            <CitySelect isHeader />
          </>
        )}
        <UserMenu />
      </div>
    </header>
  );
}

export const Header = memo(HeaderComponent);