// ── Analyze ──
export interface AnalyzeRequest {
  ticker: string;
  use_fmi?: boolean;
  thread_id?: string;
}

export interface AnalyzeResponse {
  status?: "training" | "error";
  detail?: string;
  task_id?: string;
  ticker?: string;
  report?: string;
  final_report?: string;
  recommendation?: string;
  confidence?: string;
  predictions?: PredictionsRaw;
}

export type PredictionsRaw =
  | PredictionItem[]
  | PredictionsDict
  | string
  | undefined;

export interface PredictionItem {
  date?: string;
  dt?: string;
  close?: number;
  price?: number;
}

export interface PredictionsDict {
  full_forecast?: PredictionItem[];
  forecast?: PredictionItem[];
  predictions?: PredictionItem[];
  data?: PredictionItem[];
  history?: PredictionItem[];
  [key: string]: unknown;
}

export interface ForecastPoint {
  date: string;
  close: number;
}

// ── Train ──
export interface TrainChildRequest {
  ticker: string;
}

export interface TrainResponse {
  status: "started" | "completed" | "already running" | "running" | "started_parent";
  task_id?: string;
  [key: string]: unknown;
}

// ── Predict ──
export interface PredictChildRequest {
  ticker: string;
}

export interface PredictResponse {
  status?: "training";
  detail?: string;
  task_id?: string;
  result?: Record<string, unknown>;
}

// ── Status ──
export interface StatusResponse {
  status: "running" | "completed" | "failed";
  elapsed_seconds?: number;
  [key: string]: unknown;
}

// ── Monitor ──
export interface MonitorResponse {
  ticker: string;
  type?: string;
  is_parent?: boolean;
  drift?: Record<string, unknown>;
  agent_eval?: Record<string, unknown>;
  links?: Record<string, string>;
}

export interface DriftReport {
  health?: string;
  drift_score?: number;
  volatility_index?: number;
  feature_metrics?: Record<string, Record<string, number>>;
  [key: string]: unknown;
}

export interface EvalReport {
  metrics?: {
    overall_score?: number;
    status?: string;
    checks?: Record<string, boolean>;
  };
  output_preview_text?: string;
  [key: string]: unknown;
}

// ── System ──
export interface LogsResponse {
  logs: string;
  filename: string;
}

export interface CacheResponse {
  [key: string]: unknown;
}

// ── Outputs ──
export interface OutputsListResponse {
  path: string;
  total_items: number;
  contents: OutputItem[];
}

export interface OutputItem {
  name: string;
  type: string;
  [key: string]: unknown;
}

export interface OutputsTickerResponse {
  ticker: string;
  path: string;
  files_by_category: Record<string, string[]>;
  all_files: string[];
}
