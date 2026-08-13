import { useEffect, useState } from "react";
import { userManager } from "./oidc";

type LoginState = {
  returnTo?: string;
};

export function OidcCallback() {
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    userManager
      .signinRedirectCallback()
      .then((user) => {
        window.location.replace((user.state as LoginState | undefined)?.returnTo || "/admin/scenarios");
      })
      .catch((err: unknown) => {
        console.error("OIDC callback failed", err);
        setError(err instanceof Error ? err.message : "Не удалось завершить авторизацию");
      });
  }, []);

  if (error) return <div>Ошибка авторизации: {error}</div>;
  return <div>Выполняется вход...</div>;
}
