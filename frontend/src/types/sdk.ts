import type { Tour, TourHint, TriggerType, Audience, TriggerConfig } from "./tour";
import type { TourAnalytics } from "./api";
import type { TourVersion } from "./tour";

export interface NormalizedTour {
  id: string;
  title: string;
  description: string;
  enabled: boolean;
  priority: number;
  trigger_type: Tour["trigger_type"];
  trigger_config: TriggerConfig;
  target_path: string;
  audience: Tour["audience"];
  draft?: TourVersion | null;
  published?: TourVersion | null;
  hints: TourHint[]; 
}

export interface App {
  id: string;
  name: string;
  public_key: string;
  allowed_origins: string[];
}

export interface ResolveConfig {
  apiUrl: string;
  appKey: string;
  subjectId?: string;
  props?: Record<string, unknown>;
}

export type ResolvePayload = Partial<Tour> & {
  tour_id?: string;
  tour_version_id?: string;
  current_hint_id?: string;
  hints?: Tour["hints"];
  tour?: Partial<Tour>;
};

export interface PreviewVersion {
  id: string;
  trigger_type?: Tour["trigger_type"];
  trigger_config?: Tour["trigger_config"];
  target_path?: string;
  audience?: Tour["audience"];
  hints?: Tour["hints"];
}

export interface PreviewTourCard {
  tour: {
    id: string;
    title?: string;
    description?: string;
    priority?: number;
  };
  draft?: PreviewVersion | null;
  published?: PreviewVersion | null;
}

export interface PreviewState {
  onboardingPreviewTourId?: string;
  onboardingPreviewHintId?: string;
}

export interface TourAnalyticsItem {
  tour: {
    id: string;
    title: string;
    description?: string;
  };
  analytics: TourAnalytics | null;
  unpublished?: boolean;
  error?: string;
}

export interface SaveScenarioData {
  title: string;
  description: string;
  target_path: string;
  trigger_type: TriggerType;
  trigger_config: TriggerConfig;
  audience: Audience;
  hints: TourHint[];
  deletedHints: TourHint[];
  status: "draft" | "published";
}