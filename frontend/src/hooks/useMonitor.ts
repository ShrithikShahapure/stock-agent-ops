import { useState } from "react";
import { postMonitorParent, postMonitorTicker, getDrift, getEval } from "../api/monitor";
import type { DriftReport, EvalReport } from "../types/api";

interface MonitorResult {
  eval: EvalReport | null;
  drift: DriftReport | null;
}

export function useMonitor() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<MonitorResult | null>(null);
  const [ticker, setTicker] = useState<string | null>(null);

  async function monitorTicker(t: string) {
    setLoading(true);
    setError(null);
    setResult(null);
    setTicker(t);

    try {
      const isParent = t === "^GSPC";
      if (isParent) {
        await postMonitorParent();
      } else {
        await postMonitorTicker(t);
      }

      // Fetch eval and drift reports
      const [evalData, driftData] = await Promise.allSettled([
        getEval(t),
        isParent ? getDrift(t) : Promise.reject("skip"),
      ]);

      setResult({
        eval: evalData.status === "fulfilled" ? evalData.value : null,
        drift: driftData.status === "fulfilled" ? driftData.value : null,
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : "API error");
    } finally {
      setLoading(false);
    }
  }

  function reset() {
    setResult(null);
    setTicker(null);
    setError(null);
  }

  return { loading, error, result, ticker, monitorTicker, reset };
}
