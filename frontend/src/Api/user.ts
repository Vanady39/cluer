import type { User } from "../types";
import { API_DEMO_URL } from "../Config/env";
import { userManager } from "../Auth/oidc";
import axios from "axios";

export const demoApi = axios.create({
  baseURL: API_DEMO_URL,
});

export const usersAPI = {
  getMe: async (): Promise<User> => {
    const oidcUser = await userManager.getUser();

    return (
      await demoApi.get<{ data: User }>("/users/me", {
        headers: oidcUser?.id_token
          ? {
              Authorization: `Bearer ${oidcUser.id_token}`,
            }
          : undefined,
      })
    ).data.data;
  },
};