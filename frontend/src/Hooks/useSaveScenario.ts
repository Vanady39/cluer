import {
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";

import { onboardingAPI } from "../Api/onboarding";

import type {
  CreateHintRequest,
  TourHint,
} from "../types/sdk";

function toHintRequest(
  hint: TourHint,
): CreateHintRequest {
  return {
    title: hint.title,
    content: hint.content,

    selector: hint.selector ?? "",
    placement: hint.placement,

    media_url: hint.media_url ?? "",
    spotlight: Boolean(
      hint.spotlight,
    ),

    required: Boolean(
      hint.required,
    ),

    wait_for_selector: Boolean(
      hint.wait_for_selector,
    ),

    input_placeholder:
      hint.input_placeholder ?? "",

    expected_input:
      hint.expected_input ?? "",
  };
}

export function useSaveScenario(
  editId?: string | null,
) {
  const queryClient =
    useQueryClient();

  return useMutation({
    mutationFn: async (
      data: any,
    ) => {
      // =========================
      // РЕДАКТИРОВАНИЕ
      // =========================

      if (editId) {
        await onboardingAPI.updateTourMeta(
          editId,
          {
            title: data.title,
            description:
              data.description,
          },
        );

        await onboardingAPI.updateTour(
          editId,
          {
            target_path:
              data.target_path,

            trigger_type:
              data.trigger_type,

            audience:
              data.audience,
          },
        );

        for (
          const hint of
          data.deletedHints || []
        ) {
          await onboardingAPI.deleteHint(
            editId,
            hint.id,
          );
        }

        for (
          const hint of
          data.hints || []
        ) {
          const request =
            toHintRequest(hint);

          const isNew =
            hint.id.length !== 36;

          if (isNew) {
            await onboardingAPI.createHint(
              editId,
              request,
            );
          } else {
            await onboardingAPI.updateHint(
              editId,
              hint.id,
              request,
            );
          }
        }

        if (
          data.status ===
          "published"
        ) {
          await onboardingAPI.publishTour(
            editId,
          );
        }

        return editId;
      }

      // =========================
      // НОВЫЙ СЦЕНАРИЙ
      // =========================

      const tourId =
        await onboardingAPI.createTour(
          {
            title: data.title,

            description:
              data.description,

            target_path:
              data.target_path,

            priority: 1,

            trigger_type:
              data.trigger_type,

            audience:
              data.audience,
          },
        );

      for (
        const hint of
        data.hints || []
      ) {
        await onboardingAPI.createHint(
          tourId,
          toHintRequest(hint),
        );
      }

      if (
        data.status ===
        "published"
      ) {
        await onboardingAPI.publishTour(
          tourId,
        );
      }

      return tourId;
    },

    onSuccess() {
      queryClient.invalidateQueries({
        queryKey: ["tours"],
      });

      window.location.href =
        "/admin/scenarios";
    },

    onError(error: any) {
      console.error(
        "FULL ERROR",
        error,
      );

      console.error(
        "STATUS",
        error.response?.status,
      );

      console.error(
        "RESPONSE",
        error.response?.data,
      );

      alert(
        JSON.stringify(
          {
            status:
              error.response?.status,

            data:
              error.response?.data,
          },
          null,
          2,
        ),
      );
    },
  });
}