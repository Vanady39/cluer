import type { Tour, TourVersion, TourHint } from "../../types/sdk";

interface TourResponseDto {
  tour: {
    id: string;
    title: string;
    description: string;
    enabled: boolean;
    priority: number;
  };
  draft?: TourVersion | null;
  published?: TourVersion | null;
}

export interface NormalizedTour {
  id: string;
  title: string;
  description: string;
  enabled: boolean;
  priority: number;
  trigger_type: Tour["trigger_type"];
  target_path: string;
  audience: Tour["audience"];
  draft?: TourVersion | null;
  published?: TourVersion | null;
  hints: TourHint[]; 
}

export function normalizeTour(data: TourResponseDto, hints: TourHint[] = []): NormalizedTour {
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