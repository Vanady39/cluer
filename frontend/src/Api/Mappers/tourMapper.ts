import type { NormalizedTour } from "../../types";
import type { TourResponseDto } from "../../types";
import type { TourHint } from "../../types";

export function normalizeTour(
  data: TourResponseDto,
  hints: TourHint[] = [],
): NormalizedTour {
  return {
    id: data.tour.id,
    title: data.tour.title,
    description: data.tour.description,
    enabled: data.tour.enabled,
    priority: data.tour.priority,

    trigger_type:
      data.draft?.trigger_type ??
      data.published?.trigger_type ??
      "on_load",

    trigger_config:
      data.draft?.trigger_config ??
      data.published?.trigger_config ??
      {},

    target_path:
      data.draft?.target_path ??
      data.published?.target_path ??
      "/",

    audience:
      data.draft?.audience ??
      data.published?.audience ??
      {
        show_once: true,
        max_shows: 1,
        only_new: false,
      },

    draft: data.draft,
    published: data.published,
    hints,
  };
}