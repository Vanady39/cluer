import axios from "axios";

export const demoApi = axios.create({
  baseURL: "http://localhost:8081/v1",
});