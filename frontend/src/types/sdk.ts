export interface TourHint {
  id: string;
  step: number;
  title: string;
  content: string;
  selector?: string;
  placement: "top" | "bottom" | "left" | "right" | "center" | "left-top" | "right-top";
  page_path: string;
  spotlight: boolean;
  wait_for_selector: boolean;
  media_url?: string;
}

export interface Tour {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published";
  target_path: string;
  priority: number;
  trigger_type: "on_load" | "delay" | "exit_intent" | "manual";
  audience: {
    show_once: boolean;
    max_shows: number;
    only_new: boolean;
  };
  hints: TourHint[];
  updated_at: string;
}

export interface CreateTourRequest {
  title: string;
  target_path: string;
  description: string;
  priority: number;
  trigger_type: "on_load" | "delay" | "exit_intent" | "manual";
  audience: {
    show_once: boolean;
    max_shows: number;
    only_new: boolean;
  };
}

export interface CreateHintRequest {
  title: string;
  content: string;
  placement: "top" | "bottom" | "center" | "left" | "right" | "left-top" | "right-top";
  page_path: string;
  selector?: string;
  spotlight: boolean;
  wait_for_selector: boolean;
  media_url?: string;
}