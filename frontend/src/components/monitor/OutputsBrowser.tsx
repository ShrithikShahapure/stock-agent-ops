import { useEffect, useState } from "react";
import { getOutputs } from "../../api/outputs";
import type { OutputsListResponse } from "../../types/api";

export default function OutputsBrowser() {
  const [data, setData] = useState<OutputsListResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getOutputs()
      .then(setData)
      .catch(() => setError("Could not fetch outputs"));
  }, []);

  if (error || !data || data.contents.length === 0) {
    return null;
  }

  const tickers = data.contents
    .filter((c) => c.type === "directory")
    .map((c) => c.name.toUpperCase())
    .sort();

  if (tickers.length === 0) return null;

  return (
    <div className="glass-card p-5">
      <div className="mb-3 flex items-center gap-2">
        <svg className="h-4 w-4 text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <h3 className="text-xs font-semibold uppercase tracking-wider text-[#8b949e]">
          Recent Activities
        </h3>
      </div>
      <div className="flex flex-wrap gap-2">
        {tickers.map((t) => (
          <span
            key={t}
            className="rounded-lg border border-white/[0.06] bg-white/[0.03] px-3 py-1.5 text-xs font-medium text-gray-300"
          >
            {t}
          </span>
        ))}
      </div>
    </div>
  );
}
