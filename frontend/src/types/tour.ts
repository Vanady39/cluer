export type TriggerType = "on_load" | "delay" | "exit_intent" | "manual";

export type Placement = "top" | "bottom" | "left" | "right" | "center";

export interface TriggerConfig {
  delay_ms?: number;
}

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
  trigger_config?: TriggerConfig;
  audience: Audience;
  hints: TourHint[];
  current_hint_id?: string;
  version_id?: string;
  updated_at?: string;
  created_at?: string;
}

export interface TourVersion {
  id: string;
  tour_id: string;
  version: number;
  status: "draft" | "published" | "archived";
  trigger_type: TriggerType;
  trigger_config?: TriggerConfig;
  target_path: string;
  audience: Audience;
  created_by?: string;
  created_at: string;
  published_at?: string | null;
  archived_at?: string | null;
  hints?: TourHint[];
}