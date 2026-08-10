import axios from "axios";

export const APP_ID =
  "6adb48e4-a338-42b8-af6f-e46364e61aaa";

  export const demoApi = axios.create({
  baseURL: "http://localhost:8081/v1",
});