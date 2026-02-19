import Plot from "react-plotly.js";
import type { ForecastPoint } from "../../types/api";

interface Props {
  forecast: ForecastPoint[];
  history: ForecastPoint[];
  ticker: string;
}

export default function ForecastChart({ forecast, history, ticker }: Props) {
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

    if (forecast.length > 0) {
      const fSorted = [...forecast].sort(
        (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
      );
      const lastHist = sorted[sorted.length - 1];
      const comboX = [lastHist.date, ...fSorted.map((p) => p.date)];
      const comboY = [lastHist.close, ...fSorted.map((p) => p.close)];

      traces.push({
        x: comboX,
        y: comboY,
        mode: "lines",
        name: "Forecast",
        line: { color: "#10b981", width: 2, dash: "dash" },
        fill: "tozeroy",
        fillcolor: "rgba(16, 185, 129, 0.05)",
      });
    }
  } else if (forecast.length > 0) {
    const fSorted = [...forecast].sort(
      (a, b) => new Date(a.date).getTime() - new Date(b.date).getTime(),
    );
    traces.push({
      x: fSorted.map((p) => p.date),
      y: fSorted.map((p) => p.close),
      mode: "lines",
      name: "Forecast",
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
        <div className="flex gap-3 text-xs text-[#8b949e]">
          <span className="flex items-center gap-1.5">
            <span className="h-0.5 w-4 rounded bg-[#3b82f6]" />
            History
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-0.5 w-4 rounded bg-[#10b981]" style={{ borderTop: "1px dashed #10b981", height: 0 }} />
            Forecast
          </span>
        </div>
      </div>
      <div className="p-2">
        <Plot
          data={traces}
          layout={{
            template: "plotly_dark" as unknown as Plotly.Template,
            plot_bgcolor: "rgba(0,0,0,0)",
            paper_bgcolor: "rgba(0,0,0,0)",
            yaxis: {
              title: { text: "Price (USD)" },
              gridcolor: "rgba(255,255,255,0.03)",
              zerolinecolor: "rgba(255,255,255,0.05)",
            },
            xaxis: {
              gridcolor: "rgba(255,255,255,0.03)",
              zerolinecolor: "rgba(255,255,255,0.05)",
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
            font: { color: "#8b949e", family: "system-ui" },
          }}
          useResizeHandler
          style={{ width: "100%" }}
          config={{ displayModeBar: false }}
        />
      </div>
    </div>
  );
}
