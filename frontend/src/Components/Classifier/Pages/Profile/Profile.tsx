import { memo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { usersAPI } from "../../../../Api/user";
import { Icon } from "../../../UI/Icon/Icon";
import styles from "./Styles.module.scss";

function ProfileComponent() {
  const [avatarError, setAvatarError] = useState(false);
  const {
    data: user,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["current-user"],
    queryFn: usersAPI.getMe,
  });

  if (isLoading) {
    return (
      <main className={styles.profile}>
        <div className={styles.profile__status}>Загрузка профиля...</div>
      </main>
    );
  }

  if (isError || !user) {
    return (
      <main className={styles.profile}>
        <div className={styles.profile__status}>
          Не удалось загрузить профиль
        </div>
      </main>
    );
  }

  return (
    <main className={styles.profile}>
      <div className={styles.profile__card}>
        <div className={styles.profile__avatar}>
          {user.avatarUrl && !avatarError ? (
            <img src={user.avatarUrl} onError={() => setAvatarError(true)} />
          ) : (<Icon size={40} />)}
        </div>

        <div className={styles.profile__info}>
          <h1 className={styles.profile__name}>
            {user.name || user.username || user.email || "Пользователь"}
          </h1>
          <div className={styles.profile__field}>
            <span>ID пользователя</span>
            <strong>{user.subject}</strong>
          </div>
        </div>
      </div>
    </main>
  );
}

export const Profile = memo(ProfileComponent);
