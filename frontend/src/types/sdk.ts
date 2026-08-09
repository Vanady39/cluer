export type TriggerType =
  | "on_load"
  | "delay"
  | "exit_intent"
  | "manual";

export type Placement =
  | "top"
  | "bottom"
  | "left"
  | "right"
  | "center";

export interface Audience {
  show_once: boolean;
  max_shows: number;
  only_new: boolean;
}

export interface TourHint {
  id: string;
  step: number;
  title: string;
  content: string;
  selector?: string;
  placement: Placement;
  page_path?: string;
  spotlight: boolean;
  wait_for_selector: boolean;
  media_url?: string;
}

export interface Tour {
  id: string;
  title: string;
  description: string;
  status?: "draft" | "published";
  target_path: string;
  priority: number;
  trigger_type: TriggerType;
  audience: Audience;

  hints: TourHint[];

  // Runtime поля от /resolve
  current_hint_id?: string;
  version_id?: string;

  updated_at?: string;
  created_at?: string;
}

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
  id: string;
  title: string;
  description?: string;
}

export interface TourVersion {
  id: string;
  tour_id: string;
  version: number;
  status: "draft" | "published" | "archived";

  trigger_type: TriggerType;
  target_path: string;
  audience: Audience;

  created_by?: string;
  created_at: string;
  published_at?: string | null;
  archived_at?: string | null;

  hints?: TourHint[];
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