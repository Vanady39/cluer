import { useMutation, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import type { Audience, TourHint, TriggerType } from "../types/sdk";

interface SaveScenarioData {
  title: string;
  description: string;
  target_path: string;
  trigger_type: TriggerType;
  audience: Audience;
  hints: TourHint[];
  deletedHints: TourHint[];
  status: "draft" | "published";
}

export function useSaveScenario(editId?: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: SaveScenarioData) => {
      if (editId) {
        await onboardingAPI.updateTourMeta(editId, {
          title: data.title,
          description: data.description,
        });

        await onboardingAPI.updateTour(editId, {
          target_path: data.target_path,
          trigger_type: data.trigger_type,
          audience: data.audience,
        });

        for (const hint of data.deletedHints || []) {
          await onboardingAPI.deleteHint(editId, hint.id);
        }

        for (const hint of data.hints || []) {
          if (hint.id.length !== 36) {
            await onboardingAPI.createHint(editId, {
              title: hint.title,
              content: hint.content,
              selector: hint.selector,
              placement: hint.placement,
              spotlight: hint.spotlight,
              wait_for_selector: hint.wait_for_selector,
              media_url: hint.media_url,
            });
          } else {
            await onboardingAPI.updateHint(editId, hint.id, {
              title: hint.title,
              content: hint.content,
              selector: hint.selector,
              placement: hint.placement,
              spotlight: hint.spotlight,
              wait_for_selector: hint.wait_for_selector,
              media_url: hint.media_url,
            });
          }
        }

        if (data.status === "published") await onboardingAPI.publishTour(editId);
        return editId;
      }

      const tourId = await onboardingAPI.createTour({
        title: data.title,
        description: data.description,
        target_path: data.target_path || data.hints?.[0]?.page_path || "/",
        priority: 1,
        trigger_type: data.trigger_type,
        audience: data.audience,
      });

      if (data.hints?.length) {
        for (const hint of data.hints) {
          await onboardingAPI.createHint(tourId, {
            title: hint.title,
            content: hint.content,
            selector: hint.selector,
            placement: hint.placement,
            spotlight: hint.spotlight,
            wait_for_selector: hint.wait_for_selector,
            media_url: hint.media_url,
          });
        }
      }

      if (data.status === "published") await onboardingAPI.publishTour(tourId);
      return tourId;
    },

    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: ["tours"],
      });
      window.location.href = "/admin/scenarios";
    },

    onError(error: unknown) {
      console.error("FULL ERROR", error);
      const err = error as { response?: { status?: number; data?: unknown } };
      console.error("STATUS", err.response?.status);
      console.error("RESPONSE", err.response?.data);

      alert(
        JSON.stringify(
          {
            status: err.response?.status,
            data: err.response?.data,
          },
          null,
          2,
        ),
      );
    },
  });
}