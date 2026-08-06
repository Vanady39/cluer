export interface TourHint {
  id: string;
  tour_id?: string;
  step?: number;
  title: string;
  content: string;
  selector?: string;
  placement: "top" | "bottom" | "center" | "left" | "right";
  target_path: string;
  spotlight: boolean;
  required: boolean;
  wait_for_selector: boolean;
  media_url?: string;
  input_placeholder?: string;
  expected_input?: string;
}

export interface Tour {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published" | "archived";
  target_path: string;
  priority: number;
  trigger_type: "on_load" | "delay" | "exit_intent" | "manual";
  audience: {
    show_once: boolean;
    max_shows: number;
    only_new: boolean;
  };
  created_at: string; 
  updated_at: string
  hints: TourHint[];
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
  selector?: string;
  spotlight: boolean;
  required: boolean;
  wait_for_selector: boolean;
  media_url?: string;
  input_placeholder?: string;
  expected_input?: string;
}