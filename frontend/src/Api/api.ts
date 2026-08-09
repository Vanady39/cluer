import axios from 'axios';

// Relative on purpose: the API is reached through whatever serves this page —
// nginx in the image, the Vite dev server locally — both of which forward /v1
// to the onboarding service. Same origin, so no CORS and no API address baked
// into the bundle. Override with VITE_API_URL at build time to point a build at
// an API elsewhere.
// `||`, not `??`: an unset build arg reaches the bundle as an empty string,
// which is not nullish but is just as unusable as undefined.
const API_URL = import.meta.env.VITE_API_URL || '/v1';

export const api = axios.create({
  baseURL: API_URL,
  headers: { 'Content-Type': 'application/json' },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});