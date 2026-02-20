import { useState } from "react";
import Plot from "react-plotly.js";
import type { ForecastPoint } from "../../types/api";
import { useTheme } from "../../context/ThemeContext";

type Horizon = "1W" | "1M" | "1Q";

const HORIZON_DAYS: Record<Horizon, number> = {
  "1W": 5,
  "1M": 21,
  "1Q": 63,
};

const HORIZON_LABELS: Record<Horizon, string> = {
  "1W": "1 Week",
  "1M": "1 Month",
  "1Q": "1 Quarter",
};

interface Props {
  forecast: ForecastPoint[];
  history: ForecastPoint[];
  ticker: string;
}

export default function ForecastChart({ forecast, history, ticker }: Props) {
  const [horizon, setHorizon] = useState<Horizon>("1W");
  const { theme } = useTheme();

  const visibleForecast = forecast.slice(0, HORIZON_DAYS[horizon]);

  const traces: Plotly.Data[] = [];

  if (history.length > 0) {
    const sorted = [...history].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
    );
    traces.push({
      x: sorted.map((p) => p.date),
      y: sorted.map((p) => p.close),
      mode: "lines",
      name: "History",
      line: { color: "#3b82f6", width: 2 },
      fill: "tozeroy",
      fillcolor: "rgba(59, 130, 246, 0.05)",
    });

    if (visibleForecast.length > 0) {
      const fSorted = [...visibleForecast].sort(
        (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
      );
      const lastHist = sorted[sorted.length - 1];
      const comboX = [lastHist.date, ...fSorted.map((p) => p.date)];
      const comboY = [lastHist.close, ...fSorted.map((p) => p.close)];

      traces.push({
        x: comboX,
        y: comboY,
        mode: "lines",
        name: `Forecast (${HORIZON_LABELS[horizon]})`,
        line: { color: "#10b981", width: 2, dash: "dash" },
        fill: "tozeroy",
        fillcolor: "rgba(16, 185, 129, 0.05)",
      });
    }
  } else if (visibleForecast.length > 0) {
    const fSorted = [...visibleForecast].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
    );
    traces.push({
      x: fSorted.map((p) => p.date),
      y: fSorted.map((p) => p.close),
      mode: "lines",
      name: `Forecast (${HORIZON_LABELS[horizon]})`,
      line: { color: "#10b981", width: 2, dash: "dash" },
    });
  }

  return (
    <div className="glass-card overflow-hidden">
      <div className="flex items-center justify-between border-b border-white/[0.06] px-5 py-3">
        <div className="flex items-center gap-2">
          <svg className="h-4 w-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
          </svg>
          <span className="text-sm font-medium text-white">{ticker} Price Forecast</span>
        </div>

        <div className="flex items-center gap-3">
          {/* Horizon selector */}
          <div className="flex rounded-md border border-white/[0.08] bg-white/[0.03] p-0.5">
            {(["1W", "1M", "1Q"] as Horizon[]).map((h) => (
              <button
                key={h}
                onClick={() => setHorizon(h)}
                className={`rounded px-3 py-1 text-xs font-medium transition-all ${
                  horizon === h
                    ? "bg-blue-500/20 text-blue-300"
                    : "text-[#8b949e] hover:text-gray-300"
                }`}
              >
                {h}
              </button>
            ))}
          </div>

          {/* Legend */}
          <div className="flex gap-3 text-xs text-[#8b949e]">
            <span className="flex items-center gap-1.5">
              <span className="h-0.5 w-4 rounded bg-[#3b82f6]" />
              History
            </span>
            <span className="flex items-center gap-1.5">
              <span className="h-0.5 w-4 rounded bg-[#10b981]" />
              Forecast
            </span>
          </div>
        </div>
      </div>

      <div className="p-2">
        <Plot
          data={traces}
          layout={{
            template: (theme === "dark" ? "plotly_dark" : "plotly_white") as unknown as Plotly.Template,
            plot_bgcolor: "rgba(0,0,0,0)",
            paper_bgcolor: "rgba(0,0,0,0)",
            yaxis: {
              title: { text: "Price (USD)" },
              gridcolor: theme === "dark" ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.06)",
              zerolinecolor: theme === "dark" ? "rgba(255,255,255,0.05)" : "rgba(0,0,0,0.08)",
              color: theme === "dark" ? "#8b949e" : "#57606a",
            },
            xaxis: {
              gridcolor: theme === "dark" ? "rgba(255,255,255,0.03)" : "rgba(0,0,0,0.06)",
              zerolinecolor: theme === "dark" ? "rgba(255,255,255,0.05)" : "rgba(0,0,0,0.08)",
              color: theme === "dark" ? "#8b949e" : "#57606a",
            },
            hovermode: "x unified",
            margin: { l: 55, r: 15, t: 10, b: 40 },
            height: 420,
            legend: {
              orientation: "h",
              yanchor: "bottom",
              y: 1.02,
              xanchor: "right",
              x: 1,
              font: { color: "#8b949e", size: 11 },
            },
            font: { color: theme === "dark" ? "#8b949e" : "#57606a", family: "system-ui" },
          }}
          useResizeHandler
          style={{ width: "100%" }}
          config={{ displayModeBar: false }}
        />
      </div>
    </div>
  );
}
