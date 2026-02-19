import type { DriftReport } from "../../types/api";
import StatusBadge from "../common/StatusBadge";

interface Props {
  data: DriftReport;
}

function DriftBar({ score, max = 1 }: { score: number; max?: number }) {
  const pct = Math.min((score / max) * 100, 100);
  const color =
    pct < 30 ? "#3fb950" : pct < 60 ? "#d29922" : "#f85149";

  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-white/[0.04]">
      <div
        className="h-full rounded-full transition-all duration-700"
        style={{
          width: `${pct}%`,
          background: color,
          boxShadow: `0 0 8px ${color}40`,
        }}
      />
    </div>
  );
}

export default function DriftStatus({ data }: Props) {
  const health = data.health ?? "Unknown";
  const driftScore = data.drift_score ?? 0;
  const volatility = data.volatility_index ?? 0;
  const featureMetrics = data.feature_metrics ?? {};

  return (
    <div className="glass-card p-5">
      <div className="mb-4 flex items-center gap-2">
        <svg className="h-4 w-4 text-yellow-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        <h3 className="text-sm font-semibold text-white">Data Drift</h3>
        <span className="ml-auto">
          <StatusBadge status={health} />
        </span>
      </div>

      <div className="mb-4 space-y-3">
        <div>
          <div className="mb-1 flex items-center justify-between text-xs">
            <span className="text-[#8b949e]">Drift Score</span>
            <span className="font-medium text-white">{driftScore.toFixed(3)}</span>
          </div>
          <DriftBar score={driftScore} />
        </div>

        <div className="flex items-center justify-between rounded-lg bg-white/[0.02] px-3 py-2">
          <span className="text-xs text-[#8b949e]">Volatility Index</span>
          <span className="text-sm font-semibold text-white">{volatility.toFixed(3)}</span>
        </div>
      </div>

      {Object.keys(featureMetrics).length > 0 && (
        <details>
          <summary className="cursor-pointer text-xs text-[#8b949e] hover:text-gray-300 transition-colors">
            Feature Breakdown ({Object.keys(featureMetrics).length} features)
          </summary>
          <div className="mt-2 overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-white/[0.06]">
                  <th className="py-1.5 text-left font-medium text-[#8b949e]">Feature</th>
                  {Object.keys(Object.values(featureMetrics)[0] ?? {}).map(
                    (col) => (
                      <th key={col} className="py-1.5 text-right font-medium text-[#8b949e]">
                        {col}
                      </th>
                    ),
                  )}
                </tr>
              </thead>
              <tbody>
                {Object.entries(featureMetrics).map(([feat, vals]) => (
                  <tr key={feat} className="border-b border-white/[0.03]">
                    <td className="py-1.5 text-gray-300">{feat}</td>
                    {Object.values(vals).map((v, i) => (
                      <td key={i} className="py-1.5 text-right text-gray-400">
                        {typeof v === "number" ? v.toFixed(4) : String(v)}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </details>
      )}
    </div>
  );
}
