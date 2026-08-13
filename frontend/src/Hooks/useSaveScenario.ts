import { useMutation, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import { getCurrentApp } from "../Api/Helpers/Helpers";
import type { SaveScenarioData } from "../types/sdk";
import { isAxiosError } from "axios";

type BackendError = {
  error?: string;
  message?: string;
};

type TriggerField =
  | "trigger_config.delay_ms"
  | "trigger_config.scroll_depth"
  | "trigger_config.inactivity_secs";

type ValidationIssue = {
  field?: TriggerField;
  message: string;
};

function parseValidationIssues(errorText: string): ValidationIssue[] {
  return errorText
    .split(";")
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      if (item.includes("trigger_config.delay_ms")) {
        return {
          field: "trigger_config.delay_ms",
          message: "Укажите задержку от 100 до 600000 мс",
        };
      }

      if (item.includes("trigger_config.scroll_depth")) {
        return {
          field: "trigger_config.scroll_depth",
          message: "Укажите глубину прокрутки от 1 до 100%",
        };
      }

      if (item.includes("trigger_config.inactivity_secs")) {
        return {
          field: "trigger_config.inactivity_secs",
          message: "Укажите время бездействия не меньше 3 секунд",
        };
      }

      return {
        message: item
          .replace(/^draft is invalid:\s*/i, "")
          .replace(/^tour cannot be published:\s*/i, ""),
      };
    });
}

export function useSaveScenario(
  editId?: string | null,
  onValidationError?: (field: TriggerField, message: string) => void,
) {
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
          trigger_config: data.trigger_config,
          audience: data.audience,
        });

        for (const hint of data.deletedHints || []) {
          await onboardingAPI.deleteHint(editId, hint.id);
        }

        const orderedHintIds: string[] = [];

        for (const hint of data.hints || []) {
          if (hint.id.startsWith("new-")) {
            const createdHintId = await onboardingAPI.createHint(editId, {
              title: hint.title,
              content: hint.content,
              selector: hint.selector,
              placement: hint.placement,
              page_path: hint.page_path,
              spotlight: hint.spotlight,
              wait_for_selector: hint.wait_for_selector,
              media_url: hint.media_url,
            });

            orderedHintIds.push(createdHintId);
          } else {
            await onboardingAPI.updateHint(editId, hint.id, {
              title: hint.title,
              content: hint.content,
              selector: hint.selector,
              placement: hint.placement,
              page_path: hint.page_path,
              spotlight: hint.spotlight,
              wait_for_selector: hint.wait_for_selector,
              media_url: hint.media_url,
            });

            orderedHintIds.push(hint.id);
          }
        }

        if (orderedHintIds.length > 0)
          await onboardingAPI.reorderHints(editId, orderedHintIds);
        if (data.status === "published")
          await onboardingAPI.publishTour(editId);
        return editId;
      }

      const app = await getCurrentApp();
      const tourId = await onboardingAPI.createTour(app.id, {
        title: data.title,
        description: data.description,
        target_path: data.target_path || data.hints?.[0]?.page_path || "/",
        priority: 1,
        trigger_type: data.trigger_type,
        trigger_config: data.trigger_config,
        audience: data.audience,
      });

      if (data.hints?.length) {
        const orderedHintIds: string[] = [];
        for (const hint of data.hints) {
          const createdHintId = await onboardingAPI.createHint(tourId, {
            title: hint.title,
            content: hint.content,
            selector: hint.selector,
            placement: hint.placement,
            page_path: hint.page_path,
            spotlight: hint.spotlight,
            wait_for_selector: hint.wait_for_selector,
            media_url: hint.media_url,
          });
          orderedHintIds.push(createdHintId);
        }
        await onboardingAPI.reorderHints(tourId, orderedHintIds);
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
      console.error("[Scenario] SAVE ERROR", error);

      if (!isAxiosError<BackendError>(error)) {
        alert("Не удалось сохранить сценарий");
        return;
      }

      const status = error.response?.status;
      const data = error.response?.data;

      if ((status === 400 || status === 422) && data?.error) {
        const issues = parseValidationIssues(data.error);

        issues.forEach((issue) => {
          if (issue.field) onValidationError?.(issue.field, issue.message);
        });

        alert(
          [
            status === 422
              ? "Не удалось опубликовать сценарий:"
              : "Не удалось сохранить сценарий:",
            "",
            ...issues.map((issue) => `• ${issue.message}`),
          ].join("\n"),
        );
        return;
      }
      alert(data?.message || data?.error || "Не удалось сохранить сценарий");
    },
  });
}
