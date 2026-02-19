import client from "./client";
import type { TrainChildRequest, TrainResponse } from "../types/api";

export async function postTrainParent(): Promise<TrainResponse> {
  const { data } = await client.post<TrainResponse>("/train-parent");
  return data;
}

export async function postTrainChild(req: TrainChildRequest): Promise<TrainResponse> {
  const { data } = await client.post<TrainResponse>("/train-child", req);
  return data;
}
