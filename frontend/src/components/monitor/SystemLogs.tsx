import { useState } from "react";
import { getLogs } from "../../api/system";

export default function SystemLogs() {
  const [lines, setLines] = useState(50);
  const [logs, setLogs] = useState<string | null>(null);
  const [filename, setFilename] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function fetchLogs() {
    setLoading(true);
    setError(null);
    try {
      const data = await getLogs(lines);
      setLogs(data.logs);
      setFilename(data.filename);
    } catch {
      setError("API offline");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="glass-card p-5">
      <div className="mb-3 flex items-center gap-2">
        <svg className="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <h3 className="text-xs font-semibold uppercase tracking-wider text-[#8b949e]">
          System Logs
        </h3>
      </div>

      <div className="mb-3 flex items-center gap-3">
        <input
          type="range"
          min={10}
          max={200}
          value={lines}
          onChange={(e) => setLines(Number(e.target.value))}
          className="flex-1 accent-blue-500"
        />
        <span className="w-8 text-right text-xs text-[#8b949e]">{lines}</span>
        <button
          onClick={fetchLogs}
          disabled={loading}
          className="btn-secondary px-3 py-1.5 text-xs"
        >
          {loading ? "..." : "Fetch"}
        </button>
      </div>

      {error && <p className="text-xs text-red-400">{error}</p>}

      {logs !== null && (
        <div className="terminal overflow-hidden">
          <div className="terminal-header">
            <span className="terminal-dot" style={{ background: "#ff5f56" }} />
            <span className="terminal-dot" style={{ background: "#ffbd2e" }} />
            <span className="terminal-dot" style={{ background: "#27c93f" }} />
            <span className="ml-2 text-[10px] text-[#8b949e]">{filename}</span>
          </div>
          <pre className="max-h-52 overflow-auto p-3 text-xs leading-relaxed">
            {logs}
          </pre>
        </div>
      )}
    </div>
  );
}
