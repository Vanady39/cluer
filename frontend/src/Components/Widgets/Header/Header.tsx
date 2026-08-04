import styles from "./Styles.module.scss";
import { memo, useRef, useState } from "react";
import { Avatar, Button, Dropdown, Select, type MenuProps } from "antd";
import { DownOutlined, UserOutlined } from "@ant-design/icons";
import { FcGoogle } from "react-icons/fc";
import { FaLocationArrow } from "react-icons/fa";
import { Link, useNavigate } from "react-router-dom";
import useClickOutside from "../../../Hooks/useClickOutside";
import logo from "/logo.svg";

const cities = [
  { value: "moscow", label: "Москва" },
  { value: "saint-petersburg", label: "Санкт-Петербург" },
  { value: "kazan", label: "Казань" },
  { value: "ekaterinburg", label: "Екатеринбург" },
];

const categories = [
  { key: "electronics", label: "Электроника" },
  { key: "things", label: "Личные вещи" },
  { key: "transport", label: "Транспорт" },
  { key: "real-estate", label: "Недвижимость" },
  { key: "services", label: "Услуги" },
  { key: "jobs", label: "Работа" },
  { key: "home", label: "Для дома и дачи" },
  { key: "hobbies", label: "Хобби и отдых" },
  { key: "animals", label: "Животные" },
  { key: "accessories", label: "Запчасти и аксессуары" }
];

function HeaderComponent() {
  const isAuthenticated = true;
  const [isDropMenu, setIsDropMenu] = useState(false);
  const [selectedCity, setSelectedCity] = useState<string>();
  const dropdownRef = useRef<HTMLDivElement>(null);
  const timeoutId = useRef<ReturnType<typeof setTimeout> | null>(null);
  const navigate = useNavigate();

  const handleMouseEnter = () => {
    if (timeoutId.current) {
      clearTimeout(timeoutId.current);
      timeoutId.current = null;
    }
    setIsDropMenu(true);
  };

  const handleMouseLeave = () => {
    timeoutId.current = setTimeout(() => {
      setIsDropMenu(false);
    }, 100);
  };

  const toggleMenu = () => {
    setIsDropMenu(!isDropMenu);
  };

  const handleLogout = () => {
    setIsDropMenu(false);
    navigate("/");
  };

  useClickOutside(
    dropdownRef,
    () => setIsDropMenu(false),
    ["button", "a[href]"],
    isDropMenu,
  );

   const handleCategoryClick = (category: string) => {
    navigate(`/?category=${category}`);
    console.log("Выбрана категория:", category);
  };

  const categoryMenuItems: MenuProps['items'] = categories.map(cat => ({
    key: cat.key,
    label: cat.label,
    onClick: () => handleCategoryClick(cat.key),
  }));

  return (
    <header className={styles.header}>
      <div className={styles.header__left}>
        <img src={logo} className={styles.header__logo} alt="Logo" />
        <Dropdown
          menu={{ items: categoryMenuItems }}
          trigger={['click']}
          placement="bottomLeft"
          className={styles.header__dropdownCategories}
        >
          <Button type="primary" className={styles.header__button} size="large">
            Все категории <DownOutlined />
          </Button>
        </Dropdown>
      </div>

      <div className={styles.header__right}>
        <button type="button" className={styles.header__button_add}>
          Разместить объявление
        </button>
        <div className={styles.header__city}>
          <FaLocationArrow className={styles.header__cityIcon} />
          <Select
            value={selectedCity}
            options={cities}
            placeholder="Город"
            onChange={setSelectedCity}
            variant="borderless"
            suffixIcon={null}
            className={styles.header__citySelect}
            popupMatchSelectWidth={false}
          />
        </div>

        {isAuthenticated ? (
          <div
            ref={dropdownRef}
            className={styles.header__dropMenu}
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
          >
            <Link to={"/profile"}>
              <button type="button" className={styles.header__profile} onClick={toggleMenu}>
                <Avatar size={40} icon={<UserOutlined />} />
              </button>
            </Link>
            {isDropMenu && (
              <div className={styles.header__menu}>
                <button className={styles.header__menu_item} onClick={handleLogout}>
                  Выйти
                </button>
              </div>
            )}
          </div>
        ) : (
          <Button type="text" className={styles.header__button_login} icon={<FcGoogle size={20} />}>
            Войти через Google
          </Button>
        )}
      </div>
    </header>
  );
}

export const Header = memo(HeaderComponent);