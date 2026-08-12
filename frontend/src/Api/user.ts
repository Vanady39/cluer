import type { User } from "../types";
import { demoApi } from "./constants";

export const usersAPI = {
  getMe: async (): Promise<User> => {
    return (await demoApi.get<{ data: User }>("/users/me")).data.data;
  },
};