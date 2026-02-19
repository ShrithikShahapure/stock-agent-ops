import client from "./client";
import type { PredictChildRequest, PredictResponse } from "../types/api";

export async function postPredictChild(req: PredictChildRequest): Promise<PredictResponse> {
  const { data } = await client.post<PredictResponse>("/predict-child", req);
  return data;
}

export async function postPredictParent(): Promise<PredictResponse> {
  const { data } = await client.post<PredictResponse>("/predict-parent");
  return data;
}
