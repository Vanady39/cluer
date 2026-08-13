import { demoApi } from "./constants";
import type { Listing, ListingsQuery } from "../types";

export const listingsAPI = {
  getAll: ({ q, limit = 20, offset = 0 }: ListingsQuery = {}) =>
    demoApi
      .get<{ data: Listing[] }>("/listings", {
        params: {
          ...(q?.trim()
            ? { q: Array.from(q.trim()).slice(0, 200).join("") }
            : {}),
          limit,
          offset,
        },
      })
      .then((res) => res.data.data)
};
