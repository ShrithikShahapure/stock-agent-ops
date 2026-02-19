import axios from "axios";
import { API_URL } from "../config";

const client = axios.create({
  baseURL: API_URL,
  timeout: 30_000,
  headers: { "Content-Type": "application/json" },
});

export default client;
