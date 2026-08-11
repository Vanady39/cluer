import { demoApi } from "./constants";
import type { Listing } from "../types/listings";

export const listingsAPI = {
  getAll: () =>
    demoApi
      .get<{ data: Listing[] }>("/listings")
      .then((res) => res.data.data),

  getById: (id: number) =>
    demoApi
      .get<{ data: Listing }>(`/listings/${id}`)
      .then((res) => res.data.data),

  create: (data: Omit<Listing, "id">) =>
    demoApi
      .post<{ data: Listing }>("/listings", data)
      .then((res) => res.data.data),

  update: (id: number,data: Partial<Omit<Listing, "id">>) =>
    demoApi
      .put<{ data: Listing }>(
        `/listings/${id}`,
        data,
      )
      .then((res) => res.data.data),

  delete: (id: number) =>
    demoApi
      .delete<{ data: null }>(`/listings/${id}`)
      .then((res) => res.data.data),
};