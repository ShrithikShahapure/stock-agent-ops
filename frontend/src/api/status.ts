import client from "./client";
import type { StatusResponse } from "../types/api";

export async function getStatus(taskId: string): Promise<StatusResponse> {
  const { data } = await client.get<StatusResponse>(`/status/${taskId}`);
  return data;
}
