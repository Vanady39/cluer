import type { TourVersion, Audience, TriggerType, Placement } from "./tour";

export interface CreateTourRequest {
  title: string;
  target_path: string;
  description: string;
  priority: number;
  trigger_type: TriggerType;
  audience: Audience;
}

export interface CreateHintRequest {
  title: string;
  content: string;
  placement: Placement;
  selector?: string;
  spotlight: boolean;
  wait_for_selector: boolean;
  media_url?: string;
  page_path?: string;
}

export interface AnalyticsTotals {
  started: number;
  completed: number;
  dismissed: number;
  goal_reached: number;
  completion_rate: number;
  goal_rate: number;
}

export interface AnalyticsFunnelStep {
  step: number;
  hint_id: string;
  title: string;
  shown: number;
  completed: number;
  skipped: number;
  selector_missing: number;
  dropoff: number;
}

export interface TourAnalytics {
  tour_id: string;
  tour_version_id: string;
  version: number;
  from: string;
  to: string;
  totals: AnalyticsTotals;
  funnel: AnalyticsFunnelStep[];
  broken_selectors: string[];
}

export interface TourListItem {
  updated_at: string,
  id: string;
  title: string;
  description?: string;
}

export interface UpdateTourRequest {
  trigger_type?: TriggerType;
  target_path?: string;
  audience?: Audience;
}

export interface UpdateTourMetaRequest {
  title?: string;
  description?: string;
  enabled?: boolean;
  priority?: number;
}

export interface AnalyticsQuery {
  versionId?: string;
  from?: string;
  to?: string;
}

export interface TourResponseDto {
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

export interface Listing {
  id: number;
  title: string;
  description: string;
  price: number;
  imageUrl: string;
}

export interface User {
  id: number;
  name: string;
  avatarUrl?: string;
}

export interface ListingsQuery {
  q?: string;
  limit?: number;
  offset?: number;
}