import MetricCard from "../common/MetricCard";
import type { ForecastPoint } from "../../types/api";

interface Props {
  recommendation: string;
  confidence: string;
  history: ForecastPoint[];
  forecast: ForecastPoint[];
}

export default function KPIRow({ recommendation, confidence, history, forecast }: Props) {
  let latestPrice = "N/A";
  if (history.length > 0) {
    latestPrice = `$${history[history.length - 1].close.toFixed(2)}`;
  } else if (forecast.length > 0) {
    latestPrice = `$${forecast[0].close.toFixed(2)}`;
  }

  const rec = recommendation.toUpperCase();
  const isBullish = rec.includes("BUY") || rec.includes("BULL");
  const recColor = isBullish ? "#4ade80" : "#f87171";

  return (
    <div className="grid grid-cols-3 gap-4">
      <MetricCard
        label="Latest Price"
        value={latestPrice}
        icon={
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        }
      />
      <MetricCard
        label="Recommendation"
        value={recommendation}
        borderColor={recColor}
        valueColor={recColor}
        icon={
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5} style={{ color: recColor }}>
            <path strokeLinecap="round" strokeLinejoin="round" d={isBullish ? "M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" : "M13 17h8m0 0V9m0 8l-8-8-4 4-6-6"} />
          </svg>
        }
      />
      <MetricCard
        label="Confidence"
        value={confidence}
        icon={
          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        }
      />
    </div>
  );
}
