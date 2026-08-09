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
    mutationFn: async (data: any) => {
      // =========================
      // РЕДАКТИРОВАНИЕ
      // =========================
      if (editId) {
        // Метаданные самого тура
        await onboardingAPI.updateTourMeta(editId, {
          title: data.title,
          description: data.description,
        });

        // Данные draft-версии
        await onboardingAPI.updateTour(editId, {
          target_path: data.target_path,
          trigger_type: data.trigger_type,
          audience: data.audience,
        });

        console.log("TOUR UPDATED");

        for (const hint of data.deletedHints || []) {
          console.log("DELETE HINT", hint.id);

          await onboardingAPI.deleteHint(editId, hint.id);
        }

        for (const hint of data.hints || []) {
          const isNew = hint.id.length !== 36;

          if (isNew) {
            console.log("CREATE NEW HINT", hint);

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
            console.log("UPDATE HINT", hint);

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

        if (data.status === "published") {
          await onboardingAPI.publishTour(editId);
        }

        return editId;
      }

      // =========================
      // СОЗДАНИЕ НОВОГО
      // =========================

      console.log("SAVE DATA", data);

      const tourId = await onboardingAPI.createTour({
        title: data.title,
        description: data.description,
        target_path: data.hints?.[0]?.page_path || "/",
        priority: 1,
        trigger_type: data.trigger_type,
        audience: data.audience,
      });

      console.log("HINTS BEFORE CREATE", data.hints);

      for (const hint of data.hints || []) {
        console.log("CREATE HINT", hint);

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

      if (data.status === "published") {
        await onboardingAPI.publishTour(tourId);
      }

      return tourId;
    },

    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: ["tours"],
      });

      window.location.href = "/admin/scenarios";
    },

    onError(error: any) {
      console.error("FULL ERROR", error);
      console.error("STATUS", error.response?.status);
      console.error("RESPONSE", error.response?.data);

      alert(
        JSON.stringify(
          {
            status: error.response?.status,
            data: error.response?.data,
          },
          null,
          2,
        ),
      );
    },
  });
}
