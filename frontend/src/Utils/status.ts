import type { TourVersion } from "../types/tour";

export type VersionStatus = TourVersion["status"];

export type TourStatusType = "changes" | "published" | "draft" | "unknown";

export interface TourStatusInfo {
  label: string;
  type: TourStatusType;
}

const VERSION_STATUS_LABELS: Record<VersionStatus, string> = {
  draft: "Черновик",
  published: "Опубликована",
  archived: "Архивная",
};

const TOUR_STATUS_META: Record<TourStatusType, TourStatusInfo> = {
  changes: {
    label: "Есть неопубликованные изменения",
    type: "changes",
  },
  published: {
    label: "Опубликован",
    type: "published",
  },
  draft: {
    label: "Черновик",
    type: "draft",
  },
  unknown: {
    label: "Без версии",
    type: "unknown",
  },
};

export function getVersionStatusLabel(
  status: VersionStatus,
): string {
  return VERSION_STATUS_LABELS[status];
}

export function getTourStatus(tour: {
  draft?: TourVersion | null;
  published?: TourVersion | null;
}): TourStatusInfo {
  if (tour.published && tour.draft) {
    return TOUR_STATUS_META.changes;
  }

  if (tour.published) {
    return TOUR_STATUS_META.published;
  }

  if (tour.draft) {
    return TOUR_STATUS_META.draft;
  }

  return TOUR_STATUS_META.unknown;
}