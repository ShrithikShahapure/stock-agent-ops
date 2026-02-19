import type { ForecastPoint, PredictionsRaw, PredictionItem, PredictionsDict } from "../types/api";

function extractPoint(item: PredictionItem): ForecastPoint | null {
  const d = item.date ?? item.dt;
  const c = item.close ?? item.price;
  if (d != null && c != null) {
    return { date: String(d), close: Number(c) };
  }
  return null;
}

export function parsePredictions(
  raw: PredictionsRaw,
): { forecast: ForecastPoint[]; history: ForecastPoint[] } {
  const forecast: ForecastPoint[] = [];
  const history: ForecastPoint[] = [];

  if (!raw) return { forecast, history };

  // Array of items
  if (Array.isArray(raw)) {
    for (const item of raw) {
      if (typeof item === "object" && item !== null) {
        const pt = extractPoint(item as PredictionItem);
        if (pt) forecast.push(pt);
      }
    }
    return { forecast, history };
  }

  // Dict format
  if (typeof raw === "object") {
    const dict = raw as PredictionsDict;
    let candidates: PredictionItem[] = [];

    for (const k of ["full_forecast", "forecast", "predictions", "data"]) {
      const val = dict[k];
      if (Array.isArray(val)) {
        candidates = val as PredictionItem[];
        break;
      }
    }

    if (candidates.length === 0) {
      for (const v of Object.values(dict)) {
        if (Array.isArray(v)) {
          candidates = v as PredictionItem[];
          break;
        }
      }
    }

    for (const it of candidates) {
      if (typeof it === "object" && it !== null) {
        const pt = extractPoint(it);
        if (pt) forecast.push(pt);
      }
    }

    if (Array.isArray(dict.history)) {
      for (const h of dict.history) {
        const pt = extractPoint(h);
        if (pt) history.push(pt);
      }
    }

    return { forecast, history };
  }

  // String fallback
  if (typeof raw === "string") {
    const lines = raw.split("\n").map((l) => l.trim()).filter(Boolean);
    for (const line of lines) {
      if (line.includes(":") && line.includes("$")) {
        try {
          const [left, right] = [
            line.substring(0, line.lastIndexOf(":")),
            line.substring(line.lastIndexOf(":") + 1),
          ];
          const parts = left.trim().split(/\s+/);
          const dStr = parts[parts.length - 1];
          const pStr = right.trim().replace(/[$,]/g, "");
          if (dStr && pStr) {
            forecast.push({ date: dStr, close: parseFloat(pStr) });
          }
        } catch {
          // skip
        }
      }
    }
    return { forecast, history };
  }

  return { forecast, history };
}
