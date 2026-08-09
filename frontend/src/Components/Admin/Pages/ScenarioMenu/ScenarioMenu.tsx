import { NavLink } from "react-router-dom";
import styles from './Styles.module.scss';

export function ScenarioMenu() {
  return (
    <nav className={styles.menu}>
      <NavLink
        to="/admin/scenarios"
        className={({ isActive }) =>
          `${styles.link} ${isActive ? styles.active : ""}`
        }
      >
        Сценарии
      </NavLink>

      <NavLink
        to="/admin/analytics"
        className={({ isActive }) =>
          `${styles.link} ${isActive ? styles.active : ""}`
        }
      >
        Аналитика
      </NavLink>
    </nav>
  );
}
