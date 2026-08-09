import { useMutation, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import type { CreateHintRequest, CreateTourRequest, Tour } from "../types/sdk";

// What the scenario form hands over: the tour fields it actually collects, plus
// its hints. target_path and priority are filled in below, not by the form, and
// status is carried by the form but not yet sent anywhere.
type SaveScenarioInput = Omit<CreateTourRequest, "target_path" | "priority"> & {
  hints: CreateHintRequest[];
  status: Tour["status"];
};

export function useSaveScenario(editId?: string | null) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: SaveScenarioInput) => {
      if (!editId) {
        console.log("SAVE DATA", data);
        const tourId = await onboardingAPI.createTour({
          title: data.title,
          description: data.description,
          target_path: "/",
          priority: 1,
          trigger_type: data.trigger_type,
          audience: data.audience,
        });

        for (const hint of data.hints) {
          await onboardingAPI.createHint(tourId, {
            title: hint.title,
            content: hint.content,
            selector: hint.selector,
            placement: hint.placement,
            page_path: hint.page_path,
            spotlight: hint.spotlight,
            wait_for_selector: hint.wait_for_selector,
            media_url: hint.media_url,
          });
        }
        return tourId;
      }
      throw new Error("Редактирование пока не поддерживается API");
    },

    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: ["tours"],
      });
      window.location.href = "/admin/scenarios";
    },
    onError(error) {
      console.error("SAVE ERROR", error);
    },
  });
}
