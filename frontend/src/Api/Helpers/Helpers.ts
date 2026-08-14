import { onboardingAPI } from "../onboarding";
import { CLUER_APP_KEY } from "../../Config/env";
import type { App } from "../../types";

export async function getCurrentApp(): Promise<App> {
  const apps = await onboardingAPI.getApps();

  if (CLUER_APP_KEY) {
    const match = apps.find((a) => a.public_key === CLUER_APP_KEY);

    if (!match) {
      throw new Error(
          "VITE_CLUER_APP_KEY не совпадает ни с одним приложением — туры создадутся под чужим app_id"
      );
    }

    return match;
  }

  if (apps.length === 0) {
    const created = await onboardingAPI.createApp({
      name: "cluer-demo",
      allowed_origins: [window.location.origin],
    });

    console.info("[Onboarding] created app public_key:", created.public_key);

    return created;
  }

  throw new Error(
      "VITE_CLUER_APP_KEY не задан, а в системе уже есть приложения. Задайте VITE_CLUER_APP_KEY, чтобы админка работала с нужным приложением."
  );
}