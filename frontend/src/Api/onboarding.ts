import { api } from "./api";
import type { App } from "../types";
import type { CreateHintRequest, CreateTourRequest, TourAnalytics, TourListItem, UpdateTourRequest, UpdateTourMetaRequest, 
              AnalyticsQuery } from "../types";
import type { TourVersion  } from "../types";
import { normalizeTour } from "./Mappers/tourMapper";


function getIdFromLocation(location?: string): string {
  if (!location) throw new Error("Location header is missing");

  const id = location.split("/").pop();
  
  if (!id) throw new Error("Invalid Location header format");
  return id;
}

export const onboardingAPI = {
  getApps: () =>
    api
      .get<App[]>("/apps")
      .then((res) => res.data),
      
  createApp: (data: { name: string; allowed_origins: string[] }) =>
  api
    .post<App>("/apps", data)
    .then((res) => res.data),
      
  getTours: (appId: string) =>
    api
      .get<TourListItem[]>("/tours", {
        params: { appId },
      })
      .then((res) => res.data),

  getTour: async (tourId: string) => {
    const data = (await api.get(`/tours/${tourId}`)).data;
    const hints = data.draft?.hints || data.published?.hints || [];
    return normalizeTour(data, hints);
  },

  getVersions: (tourId: string) =>
    api.get<TourVersion[]>(`/tours/${tourId}/versions`).then((res) => res.data),

  getVersion: (versionId: string) =>
    api.get<TourVersion>(`/versions/${versionId}`).then((res) => res.data),

  rollbackVersion: (tourId: string, versionId: string) =>
    api
      .post<TourVersion>(`/tours/${tourId}/rollback`, {
        to_version_id: versionId,
      })
      .then((res) => res.data),

  getPublished: (appId: string, path: string) =>
    api
      .get("/tours/published", {
        params: {
          appId,
          path,
        },
      })
      .then((res) => res.data),

  createTour: (appId: string, data: CreateTourRequest) =>
    api
      .post("/tours", data, {
        params: {
          appId,
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

  updateTour: (tourId: string, data: UpdateTourRequest) =>
    api.patch(`/tours/${tourId}/draft`, data).then((res) => res.data),

  updateTourMeta: (tourId: string, data: UpdateTourMetaRequest) =>
    api.patch(`/tours/${tourId}`, data).then((res) => res.data),

  createDraft: async (tourId: string) => {
    return (await api.post(`/tours/${tourId}/draft`)).data;
  },

  updateHint: (tourId: string, hintId: string, data: CreateHintRequest) =>
    api.patch(`/tours/${tourId}/hints/${hintId}`, data).then((res) => res.data),

  deleteTour: (tourId: string) => api.delete(`/tours/${tourId}`),

  deleteHint: (tourId: string, hintId: string) =>
    api.delete(`/tours/${tourId}/hints/${hintId}`),

  getAnalytics: (tourId: string, params?: AnalyticsQuery) =>
    api
      .get<TourAnalytics>(`/tours/${tourId}/analytics`, {
        params,
      })
      .then((res) => res.data),
};