import { memo, useRef, useState, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import styles from "./Styles.module.scss";
import { Button } from "../../../../../UI/Button";

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
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
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
            <button type="button" className={styles.userMenu__profile}>
              <div className={styles.userMenu__avatar}>
                <UserIcon size={20} />
              </div>
            </button>
          </Link>

          {isOpen && (
            <div className={styles.userMenu__dropdown}>
              <button
                className={styles.userMenu__item}
                onClick={() => {
                  window.dispatchEvent(new CustomEvent("start-onboarding"));
                }}
              >
                Помощь
              </button>
              <button
                className={styles.userMenu__item}
                onClick={() => {
                  setIsOpen(false);
                  navigate("/");
                }}
              >
                Выйти
              </button>
            </div>
          )}
        </>
      ) : (
        <Button className={styles.userMenu__login}>Войти через Google</Button>
      )}
    </div>
  );
}

function UserIcon({ size = 20 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
    >
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </svg>
  );
}

export const UserMenu = memo(UserMenuComponent);
