export interface Hint {
  id: string;
  title: string;
  content: string;
  selector: string;
  placement: "top" | "bottom" | "left" | "right";
  spotlight: boolean;
}

export interface Tour {
  id: string;
  title: string;
  description: string;
  status: "draft" | "published";
  target_path: string;
  hints: Hint[];
}
