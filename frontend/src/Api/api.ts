import axios from "axios";
import { userManager } from "../Auth/oidc";
import { API_URL } from "../Config/env";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use(async (config) => {
  const user = await userManager.getUser();
  if (user?.id_token) config.headers.Authorization = `Bearer ${user.id_token}`;
  return config;
});

const retriedRequests = new WeakSet<object>();

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (error.response?.status === 403) {
      if (window.location.pathname !== "/admin/access-denied") window.location.replace("/admin/access-denied");
      return Promise.reject(error);
    }

    if (error.response?.status !== 401 || !originalRequest)
      return Promise.reject(error);

    if (retriedRequests.has(originalRequest)) return Promise.reject(error);
    retriedRequests.add(originalRequest);

    try {
      const user = await userManager.signinSilent();
      if (!user?.id_token) throw new Error("OIDC renewal returned no id_token");

      originalRequest.headers.Authorization = `Bearer ${user.id_token}`;
      return api.request(originalRequest);
    } catch (renewError) {
      await userManager.signinRedirect({
        state: {
          returnTo:
            window.location.pathname +
            window.location.search +
            window.location.hash,
        },
      });
      return Promise.reject(renewError);
    }
  },
);
