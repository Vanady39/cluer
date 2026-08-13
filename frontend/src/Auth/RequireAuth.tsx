import { useEffect, useState, type ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { userManager } from "./oidc";

type RequireAuthProps = {
  children: ReactNode;
};

export function RequireAuth({ children }: RequireAuthProps) {
  const location = useLocation();
  const [authenticated, setAuthenticated] = useState<boolean | null>(null);

  useEffect(() => {
    let active = true;

    userManager
      .getUser()
      .then(async (user) => {
        if (user && !user.expired) {
          if (active) setAuthenticated(true);
          return;
        }

        await userManager.signinRedirect({
          state: {
            returnTo: location.pathname + location.search + location.hash,
          },
        });
      })
      .catch((error) => {
        console.error("OIDC authentication failed", error);
        if (active) setAuthenticated(false);
      });

    return () => {
      active = false;
    };
  }, [location.pathname, location.search, location.hash]);

  if (authenticated === null) return <div>Проверка авторизации...</div>;
  if (!authenticated) return <div>Не удалось выполнить вход</div>;

  return <>{children}</>;
}
