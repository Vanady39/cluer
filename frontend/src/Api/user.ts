import type { User } from "../types";
import { demoApi } from "./constants";
import { userManager } from "../Auth/oidc";

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