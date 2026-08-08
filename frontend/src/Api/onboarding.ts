import { api } from "./api";
import type { CreateTourRequest, CreateHintRequest } from "../types/sdk";

const APP_ID = "6adb48e4-a338-42b8-af6f-e46364e61aaa";

function getIdFromLocation(location?: string) {
  if (!location) {
    throw new Error("Location header is missing");
  }
  return location.split("/").pop()!;
}

export const onboardingAPI = {
  getTours: () =>
    api
      .get("/tours", {
        params: {
          appId: APP_ID,
        },
      })
      .then((res) => res.data),

  getTour: async (tourId: string) => {
    const tourResponse = await api.get(`/tours/${tourId}`);
    const data = tourResponse.data;

    console.log("TOUR RESPONSE", data);

    let hints: any[] = [];

    if (data.draft) {
      const hintsResponse = await api.get(`/tours/${tourId}/hints`);
      hints = hintsResponse.data ?? [];
    } else if (data.published?.hints) {
      hints = data.published.hints;
    }

    return {
      id: data.tour.id,
      title: data.tour.title,
      description: data.tour.description,
      enabled: data.tour.enabled,
      priority: data.tour.priority,

      trigger_type:
        data.draft?.trigger_type ?? data.published?.trigger_type ?? "on_load",

      target_path:
        data.draft?.target_path ?? data.published?.target_path ?? "/",

      audience: data.draft?.audience ??
        data.published?.audience ?? {
          show_once: true,
          max_shows: 1,
          only_new: false,
        },

      hints,
      draft: data.draft,
      published: data.published,
    };
  },

  getPublished: (path: string) =>
    api
      .get("/tours/published", {
        params: {
          appId: APP_ID,
          path,
        },
      })
      .then((res) => res.data),

  createTour: (data: CreateTourRequest) =>
    api
      .post("/tours", data, {
        params: {
          appId: APP_ID,
        },
      })
      .then((res) =>
        getIdFromLocation(res.headers.location as string | undefined),
      ),

  createHint: (tourId: string, data: CreateHintRequest) =>
    api
      .post(`/tours/${tourId}/hints`, data)
      .then((res) =>
        getIdFromLocation(res.headers.location as string | undefined),
      ),

  publishTour: (tourId: string) => api.post(`/tours/${tourId}/publish`),

  updateTour: (tourId: string, data: any) => {
    console.log("UPDATE TOUR REQUEST", {
      tourId,
      data,
    });

    return api.patch(`/tours/${tourId}/draft`, data).then((res) => {
      console.log("UPDATE TOUR RESPONSE", res.data);
      return res.data;
    });
  },

  updateTourMeta: (
    tourId: string,
    data: {
      title?: string;
      description?: string;
      enabled?: boolean;
      priority?: number;
    },
  ) => api.patch(`/tours/${tourId}`, data).then((res) => res.data),

  createDraft: async (tourId: string) => {
    const response = await api.post(`/tours/${tourId}/draft`);
    return response.data;
  },

  updateHint: (
    tourId: string,
    hintId: string,
    data: Partial<CreateHintRequest>,
  ) =>
    api.patch(`/tours/${tourId}/hints/${hintId}`, data).then((res) => res.data),

  deleteTour: (tourId: string) => api.delete(`/tours/${tourId}`),

  deleteHint: (tourId: string, hintId: string) =>
    api.delete(`/tours/${tourId}/hints/${hintId}`),
};
