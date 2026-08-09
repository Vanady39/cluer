import { NavLink } from "react-router-dom";
import styles from './Styles.module.scss';
import cn from "classnames";

export function ScenarioMenu() {
  return (
    <nav className={styles.menu}>
      <NavLink
        to="/admin/scenarios"
        className={({ isActive }) => cn(styles.menu__link, isActive && styles.menu__link__active)}
      >
        Сценарии
      </NavLink>
      <NavLink
        to="/admin/analytics"
        className={({ isActive }) => cn(styles.menu__link, isActive && styles.menu__link__active)}
      >
        Аналитика
      </NavLink>
    </nav>
  );
}
