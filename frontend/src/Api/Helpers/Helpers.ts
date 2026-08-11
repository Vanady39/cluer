import { onboardingAPI } from "../onboarding";

export async function getCurrentApp() {
  const apps = await onboardingAPI.getApps();
  const app = apps[0];

  if (!app) throw new Error("No apps available");
  return app;
}