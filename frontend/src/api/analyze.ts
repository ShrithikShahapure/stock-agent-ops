import client from "./client";
import type { AnalyzeRequest, AnalyzeResponse } from "../types/api";

export async function postAnalyze(req: AnalyzeRequest): Promise<AnalyzeResponse> {
  const { data } = await client.post<AnalyzeResponse>("/analyze", req, {
    timeout: 300_000,
  });
  return data;
}
