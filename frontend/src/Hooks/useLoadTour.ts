import { useCallback, useEffect, useState } from "react";
import { resolveTour } from "../Components/Onboarding/client";
import { API_URL } from "../Config/env";
import type { ResolvePayload, Tour } from "../types";
import { onboardingAPI } from "../Api/onboarding";

export function useLoadTour(
  appKey: string,
  isPreview: boolean,
  previewTourId: string | null,
  isBuilder: boolean,
  setIsOpen: (open: boolean) => void,
) {
  const [tour, setTour] = useState<Tour | null>(null);

  const loadRuntimeTour = useCallback(async (): Promise<Tour | null> => {
    if (!appKey) {
      throw new Error("[Onboarding] appKey is not configured");
    }

    const response = await resolveTour({
      apiUrl: API_URL,
      appKey,
      props: { isNewUser: true },
    });

    if (!response) return null;

    const raw = response as ResolvePayload;
    const source = raw.tour ?? raw;
    const tourId = source.id ?? raw.tour_id;

    if (!tourId) {
      throw new Error("Resolve response does not contain tour id");
    }

    return {
      id: tourId,
      title: source.title ?? "",
      description: source.description ?? "",
      target_path: source.target_path ?? "/",
      priority: source.priority ?? 0,
      trigger_type: source.trigger_type ?? "on_load",
      trigger_config: source.trigger_config,
      audience: source.audience ?? {
        show_once: false,
        max_shows: 0,
        only_new: false,
      },
      hints: source.hints ?? raw.hints ?? [],
      current_hint_id: raw.current_hint_id ?? source.current_hint_id,
      version_id: raw.tour_version_id ?? source.version_id,
    };
  }, [appKey]);

  const loadPreviewTour = useCallback(
    async (tourId: string | null): Promise<Tour | null> => {
      if (!tourId) return null;

      const card = await onboardingAPI.getTour(tourId);
      const source = card.draft ?? card.published;

      if (!source) {
        console.warn(
          "[Onboarding] PREVIEW: tour has no draft or published version",
        );
        return null;
      }

      const hints = [...(card.hints ?? [])].sort((a, b) => a.step - b.step);
      return {
        id: card.id,
        title: card.title ?? "",
        description: card.description ?? "",
        target_path: source.target_path ?? card.target_path ?? "/",
        priority: card.priority ?? 0,
        trigger_type: source.trigger_type ?? card.trigger_type ?? "on_load",
        trigger_config: source.trigger_config ?? card.trigger_config,
        audience: source.audience ??
          card.audience ?? {
            show_once: false,
            max_shows: 0,
            only_new: false,
          },
        hints,
        current_hint_id: hints[0]?.id,
        version_id: source.id,
      };
    },
    [],
  );

  const loadTour = useCallback(
    async (isPreview: boolean, previewTourId: string | null) => {
      if (isPreview) return await loadPreviewTour(previewTourId);
      return await loadRuntimeTour();
    },
    [loadRuntimeTour, loadPreviewTour],
  );

  useEffect(() => {
    let cancelled = false;
    let cleanupTrigger: (() => void) | undefined;
    async function start() {
      try {
        if (isBuilder) return;
        const resolvedTour = await loadTour(isPreview, previewTourId);

        if (cancelled) return;
        if (!resolvedTour) {
          setTour(null);
          setIsOpen(false);
          return;
        }
        setTour(resolvedTour);
        setIsOpen(false);

        const startTour = () => setIsOpen(true);
        switch (resolvedTour.trigger_type) {
          case "on_load":
            startTour();
            break;

          case "delay": {
            const delay = resolvedTour.trigger_config?.delay_ms ?? 3000;
            const timer = setTimeout(() => {
              startTour();
            }, delay);
            cleanupTrigger = () => {
              window.clearTimeout(timer);
            };
            break;
          }

          case "exit_intent": {
            const handleEvent = (event: MouseEvent) => {
              if (event.clientY <= 0) {
                startTour();
                document.removeEventListener("mouseleave", handleEvent);
              }
            };
            document.addEventListener("mouseleave", handleEvent);
            cleanupTrigger = () => {
              document.removeEventListener("mouseleave", handleEvent);
            };
            break;
          }

          case "manual":
            break;

          case "scroll_depth": {
            const requiredDepth = resolvedTour.trigger_config?.scroll_depth;
            if (requiredDepth === undefined) break;

            const handleScroll = () => {
              const currentDepth =
                ((window.scrollY + window.innerHeight) /
                  document.documentElement.scrollHeight) *
                100;

              if (currentDepth >= requiredDepth) {
                startTour();
                window.removeEventListener("scroll", handleScroll);
              }
            };
            window.addEventListener("scroll", handleScroll, { passive: true });
            handleScroll();

            cleanupTrigger = () => {
              window.removeEventListener("scroll", handleScroll);
            };
            break;
          }

          case "inactivity": {
            const inactivitySecs = resolvedTour.trigger_config?.inactivity_secs;
            if (inactivitySecs === undefined) break;

            let timer: number;
            const resetTimer = () => {
              window.clearTimeout(timer);
              timer = window.setTimeout(startTour, inactivitySecs * 1000);
            };

            const events = ["mousemove", "keydown", "scroll"] as const;
            events.forEach((eventName) => {
              window.addEventListener(eventName, resetTimer, {
                passive: true,
              });
            });

            resetTimer();
            cleanupTrigger = () => {
              window.clearTimeout(timer);
              events.forEach((eventName) => {
                window.removeEventListener(eventName, resetTimer);
              });
            };
            break;
          }

          case "element_visible": {
            const selector = resolvedTour.trigger_config?.element_selector;
            if (!selector) break;

            const element = document.querySelector(selector);
            if (!element) break;

            const observer = new IntersectionObserver((entries) => {
              if (entries.some((entry) => entry.isIntersecting)) {
                startTour();
                observer.disconnect();
              }
            });
            observer.observe(element);

            cleanupTrigger = () => {
              observer.disconnect();
            };
            break;
          }
        }
      } catch (error) {
        console.error("[Onboarding] LOAD ERROR", error);
        if (!cancelled) {
          setTour(null);
          setIsOpen(false);
        }
      }
    }
    void start();
    return () => {
      cancelled = true;
      cleanupTrigger?.();
    };
  }, [isBuilder, isPreview, previewTourId, loadTour, setIsOpen]);

  return { tour, loadTour };
}
