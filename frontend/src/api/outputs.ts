import client from "./client";
import type { OutputsListResponse, OutputsTickerResponse } from "../types/api";

export async function getOutputs(): Promise<OutputsListResponse> {
  const { data } = await client.get<OutputsListResponse>("/outputs");
  return data;
}

export async function getOutputsTicker(ticker: string): Promise<OutputsTickerResponse> {
  const { data } = await client.get<OutputsTickerResponse>(`/outputs/${ticker}`);
  return data;
}
