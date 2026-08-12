import { memo, useRef, useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../../../UI/Button";
import { Icon } from "../../../../../UI/Icon/Icon";

function UserMenuComponent() {
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const isAuthenticated = true;
  const hoverTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleMouseEnter = () => {
    if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current);
    setIsOpen(true);
  };

  const handleMouseLeave = () => {
    if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current);
    hoverTimerRef.current = setTimeout(() => {
      setIsOpen(false);
    }, 200);
  };

  useEffect(() => {
    return () => {
      if (hoverTimerRef.current) clearTimeout(hoverTimerRef.current);
    };
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) setIsOpen(false);
    };
    document.addEventListener("click", handleClickOutside);
    return () => document.removeEventListener("click", handleClickOutside);
  }, []);

  return (
    <div
      ref={ref}
      className={styles.userMenu}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {isAuthenticated ? (
        <>
          <Link to="/profile">
            <Button className={styles.userMenu__profile}>
              <div className={styles.userMenu__avatar}>
                <Icon size={20} />
              </div>
            </Button>
          </Link>

          {isOpen && (
            <div className={styles.userMenu__dropdown}>
              <Button
                className={styles.userMenu__item}
                onClick={() => { window.dispatchEvent(new CustomEvent("start-onboarding")); }}>
                Помощь
              </Button>
              <Button
                className={styles.userMenu__item}
                onClick={() => { setIsOpen(false); navigate("/"); }}>
                Выйти
              </Button>
            </div>
          )}
        </>
      ) : ( <Button className={styles.userMenu__login}>Войти через Google</Button> )}
    </div>
  );
}


export const UserMenu = memo(UserMenuComponent);
