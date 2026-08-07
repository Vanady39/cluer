export {};

declare global {
  interface Window {
    startOnboarding: (id: string) => void;
  }
}