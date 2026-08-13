import axios from "axios";
import { API_DEMO_URL } from "../Config/env";

export const demoApi = axios.create({
  baseURL: API_DEMO_URL,
});
