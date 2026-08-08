import { demoApi } from "./api_demo";
import type { Listing } from "../types/listings";

export const listingsAPI = {
  // Получить список объявлений
  getAll: () =>
    demoApi
      .get<{ data: Listing[] }>("/listings")
      .then((res) => res.data.data),


  // Получить объявление по id
  getById: (id: number) =>
    demoApi
      .get<{ data: Listing }>(`/listings/${id}`)
      .then((res) => res.data.data),


  // Создать объявление
  create: (data: Omit<Listing, "id">) =>
    demoApi
      .post<{ data: Listing }>("/listings", data)
      .then((res) => res.data.data),


  // Обновить объявление
  update: (
    id: number,
    data: Partial<Omit<Listing, "id">>,
  ) =>
    demoApi
      .put<{ data: Listing }>(
        `/listings/${id}`,
        data,
      )
      .then((res) => res.data.data),


  // Удалить объявление
  delete: (id: number) =>
    demoApi
      .delete<{ data: null }>(`/listings/${id}`)
      .then((res) => res.data.data),
};