export const API_URL = import.meta.env.VITE_API_URL || "/v1";
// Относительный путь по той же причине, что и API_URL: абсолютный адрес,
// зашитый в бандл, указывает на машину, где открыт браузер, а не на сервер.
// Оба API живут за одним nginx фронтенда, он же разводит префиксы по сервисам.
export const API_DEMO_URL = import.meta.env.VITE_API_DEMO_URL || "/demo/v1";
export const CLUER_APP_KEY = import.meta.env.VITE_CLUER_APP_KEY?.trim() || "";
