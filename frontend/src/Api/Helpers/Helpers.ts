import { onboardingAPI } from "../onboarding";
import type { App } from "../../types";

let currentAppPromise: Promise<App> | null = null;

export async function getCurrentApp(): Promise<App> {
  if (!currentAppPromise) {
    currentAppPromise = (async () => {
      const apps = await onboardingAPI.getApps();
      if (apps[0]) return apps[0];

      return onboardingAPI.createApp({
        name: "Cluer",
        allowed_origins: [window.location.origin],
      });
    })().catch((error) => {
      currentAppPromise = null;
      throw error;
    });
  }

  return currentAppPromise;
}