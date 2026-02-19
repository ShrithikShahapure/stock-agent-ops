import client from "./client";
import type { MonitorResponse, DriftReport, EvalReport } from "../types/api";

export async function postMonitorParent(): Promise<MonitorResponse> {
  const { data } = await client.post<MonitorResponse>("/monitor/parent");
  return data;
}

export async function postMonitorTicker(ticker: string): Promise<MonitorResponse> {
  const { data } = await client.post<MonitorResponse>(`/monitor/${ticker}`);
  return data;
}

export async function getDrift(ticker: string): Promise<DriftReport> {
  const { data } = await client.get<DriftReport>(`/monitor/${ticker}/drift`);
  return data;
}

export async function getEval(ticker: string): Promise<EvalReport> {
  const { data } = await client.get<EvalReport>(`/monitor/${ticker}/eval`);
  return data;
}
