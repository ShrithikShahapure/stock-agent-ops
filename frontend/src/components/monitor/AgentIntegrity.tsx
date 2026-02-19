import type { EvalReport } from "../../types/api";

interface Props {
  data: EvalReport;
}

function ScoreGauge({ score }: { score: number }) {
  const pct = score * 100;
  const radius = 40;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (score * circumference);
  const color = pct >= 80 ? "#3fb950" : pct >= 50 ? "#d29922" : "#f85149";

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width="100" height="100" viewBox="0 0 100 100">
        <circle
          cx="50" cy="50" r={radius}
          fill="none"
          stroke="rgba(255,255,255,0.04)"
          strokeWidth="6"
        />
        <circle
          cx="50" cy="50" r={radius}
          fill="none"
          stroke={color}
          strokeWidth="6"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="gauge-ring"
          transform="rotate(-90 50 50)"
          style={{ filter: `drop-shadow(0 0 4px ${color}40)` }}
        />
      </svg>
      <div className="absolute text-center">
        <div className="text-xl font-bold text-white">{pct.toFixed(0)}%</div>
        <div className="text-[9px] uppercase tracking-wider text-[#8b949e]">Trust</div>
      </div>
    </div>
  );
}

export default function AgentIntegrity({ data }: Props) {
  const metrics = data.metrics;
  if (!metrics) {
    return (
      <div className="glass-card p-5">
        <h3 className="mb-3 text-sm font-semibold text-white">Agent Integrity</h3>
        <p className="text-sm text-[#8b949e]">No evaluation data available.</p>
      </div>
    );
  }

  const score = metrics.overall_score ?? 0;
  const statusLabel = metrics.status ?? "N/A";
  const checks = metrics.checks ?? {};
  const preview = data.output_preview_text ?? "";
  const isTrainingPlaceholder =
    preview.includes("status") && preview.includes("training");

  return (
    <div className="glass-card p-5">
      <div className="mb-4 flex items-center gap-2">
        <svg className="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
        <h3 className="text-sm font-semibold text-white">Agent Integrity</h3>
        <span
          className={`ml-auto rounded-full px-2 py-0.5 text-[10px] font-medium border ${
            statusLabel.toLowerCase().includes("pass")
              ? "border-green-500/20 bg-green-500/10 text-green-400"
              : "border-yellow-500/20 bg-yellow-500/10 text-yellow-400"
          }`}
        >
          {statusLabel}
        </span>
      </div>

      <div className="mb-4 flex justify-center">
        <ScoreGauge score={score} />
      </div>

      {Object.keys(checks).length > 0 ? (
        <div className="space-y-1.5">
          {Object.entries(checks).map(([key, passed]) => (
            <div
              key={key}
              className={`flex items-center justify-between rounded-lg px-3 py-2 text-xs ${
                passed ? "bg-green-500/[0.06]" : "bg-red-500/[0.06]"
              }`}
            >
              <span className="text-gray-300">
                {key.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())}
              </span>
              <span className={passed ? "text-green-400" : "text-red-400"}>
                {passed ? "\u2713 Pass" : "\u2717 Fail"}
              </span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-center text-xs text-yellow-400">
          No heuristic checks performed.
        </p>
      )}

      {preview && (
        <details className="mt-4">
          <summary className="cursor-pointer text-xs text-[#8b949e] hover:text-gray-300 transition-colors">
            View Report Analysis
          </summary>
          <div className="mt-2 terminal p-3 text-xs whitespace-pre-wrap max-h-48 overflow-y-auto">
            {isTrainingPlaceholder && (
              <p className="mb-2 text-yellow-400">
                Agent reported model training in progress. Evaluation is a placeholder.
              </p>
            )}
            <span className="text-gray-400">{preview}</span>
          </div>
        </details>
      )}
    </div>
  );
}
