import client from "./client";
import type { LogsResponse } from "../types/api";

export async function getLogs(lines = 100): Promise<LogsResponse> {
  const { data } = await client.get<LogsResponse>("/system/logs", {
    params: { lines },
  });
  return data;
}

export async function getCache(ticker?: string) {
  const { data } = await client.get("/system/cache", {
    params: ticker ? { ticker } : undefined,
  });
  return data;
}
