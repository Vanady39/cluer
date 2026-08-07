import { api } from "./api";
import type { Tour, CreateTourRequest, CreateHintRequest,} from "../types/sdk";


function getIdFromLocation(location?: string) {
  if (!location) {
    throw new Error("Location header is missing");
  }
  return location.split("/").pop()!;
}

export const onboardingAPI = {

// Получение опубликованных сценариев для страницы
  getPublished: (path: string) =>
    api
      .get<Tour[]>("/tours/published", {
        params: {
          path,
        },
      })
      .then((res) => res.data),


  // Создание сценария
  // Сервер создаёт его как draft
  createTour: (data: CreateTourRequest) =>
    api
      .post("/tours", data)
      .then((res) =>
        getIdFromLocation(
          res.headers.location as string | undefined,
        ),
      ),


  // Добавление подсказки
  createHint: (
    tourId: string,
    data: CreateHintRequest,
  ) =>
    api
      .post(`/tours/${tourId}/hints`, data)
      .then((res) =>
        getIdFromLocation(
          res.headers.location as string | undefined,
        ),
      ),


  // Публикация сценария
  publishTour: (tourId: string) =>
    api.patch(`/tours/${tourId}`, {
      status: "published",
    }),


  // Редактирование сценария
  // Для published сервер должен вернуть 409
  updateTour: (
    tourId: string,
    data: Partial<CreateTourRequest>,
  ) =>
    api
      .put(`/tours/${tourId}`, data)
      .then((res) => res.data),


  // Редактирование подсказки
  updateHint: (
    tourId: string,
    hintId: string,
    data: Partial<CreateHintRequest>,
  ) =>
    api
      .put(
        `/tours/${tourId}/hints/${hintId}`,
        data,
      )
      .then((res) => res.data),


  // Удаление сценария
  deleteTour: (tourId: string) =>
    api
      .delete(`/tours/${tourId}`)
      .then((res) => res.data),

      
  // Удаление подсказки
  deleteHint: (
    tourId: string,
    hintId: string,
  ) =>
    api
      .delete(
        `/tours/${tourId}/hints/${hintId}`,
      )
      .then((res) => res.data),

};