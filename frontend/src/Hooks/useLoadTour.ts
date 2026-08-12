import { useCallback, useEffect, useState } from "react";
import { resolveTour } from "../Components/Onboarding/client";
import { getCurrentApp } from "../Api/Helpers/Helpers";
import { API_URL } from "../Config/env";
import type { PreviewTourCard, ResolvePayload, Tour } from "../types";

export function useLoadTour(
  isPreview: boolean,
  previewTourId: string | null,
  isBuilder: boolean,
  setIsOpen: (open: boolean) => void
) {
  const [tour, setTour] = useState<Tour | null>(null);
  const [appKey, setAppKey] = useState<string | null>(null);

  const loadRuntimeTour = useCallback(async (): Promise<Tour | null> => {
    const app = await getCurrentApp();
    setAppKey(app.public_key);
    const response = await resolveTour({
      apiUrl: API_URL,
      appKey: app.public_key,
      props: { isNewUser: true },
    });

    if (!response) return null;

    const raw = response as ResolvePayload;
    const source = raw.tour ?? raw;
    const tourId = source.id ?? raw.tour_id;

    if (!tourId) throw new Error("Resolve response does not contain tour id");
    return {
      id: tourId,
      title: source.title ?? "",
      description: source.description ?? "",
      target_path: source.target_path ?? "/",
      priority: source.priority ?? 0,
      trigger_type: source.trigger_type ?? "on_load",
      trigger_config: source.trigger_config,
      audience: source.audience ?? { show_once: false, max_shows: 0, only_new: false },
      hints: source.hints ?? raw.hints ?? [],
      current_hint_id: raw.current_hint_id ?? source.current_hint_id,
      version_id: raw.tour_version_id ?? source.version_id,
    };
  }, []);

  const loadPreviewTour = useCallback(
    async (tourId: string | null): Promise<Tour | null> => {
      if (!tourId) return null;

      const response = await fetch(`${API_URL}/tours/${tourId}`);
      if (!response.ok) throw new Error(`Preview request failed: ${response.status}`);

      const card = (await response.json()) as PreviewTourCard;
      const source = card.draft ?? card.published;

      if (!source) {
        console.warn("[Onboarding] PREVIEW: tour has no draft or published version");
        return null;
      }

      const hints = [...(source.hints ?? [])].sort((a, b) => a.step - b.step);
      return {
        id: card.tour.id,
        title: card.tour.title ?? "",
        description: card.tour.description ?? "",
        target_path: source.target_path ?? "/",
        priority: card.tour.priority ?? 0,
        trigger_type: source.trigger_type ?? "on_load",
        trigger_config: source.trigger_config,
        audience: source.audience ?? { show_once: false, max_shows: 0, only_new: false },
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

  return { tour, appKey, loadTour };
}