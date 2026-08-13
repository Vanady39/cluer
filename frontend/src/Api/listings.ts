import { API_DEMO_URL } from "../Config/env";
import type { Listing, ListingsQuery } from "../types";
import axios from "axios";

export const demoApi = axios.create({
  baseURL: API_DEMO_URL,
});

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
