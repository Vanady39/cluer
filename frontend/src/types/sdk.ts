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

  // Пока оставляем для редактора.
  // Backend многостраничность пока не поддерживает.
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

  // Пока можно оставить в UI-типе,
  // но в запрос backend сейчас не отправляем.
  page_path?: string;
}