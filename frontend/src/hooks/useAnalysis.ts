import { useCallback, useMemo, useRef, useState } from "react";
import { postAnalyze } from "../api/analyze";
import { getStatus } from "../api/status";
import { usePolling } from "./usePolling";
import type { AnalyzeResponse, StatusResponse } from "../types/api";

type Phase = "idle" | "analyzing" | "training" | "retrying" | "done" | "error";

export function useAnalysis() {
  const [phase, setPhase] = useState<Phase>("idle");
  const [result, setResult] = useState<AnalyzeResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [taskId, setTaskId] = useState<string | null>(null);
  const [trainingStart, setTrainingStart] = useState(0);
  const sessionId = useRef(crypto.randomUUID());
  const tickerRef = useRef("");

  const progress = useMemo(() => {
    if (phase !== "training" || !trainingStart) return 0;
    const elapsed = (Date.now() - trainingStart) / 1000;
    return Math.min(elapsed * 2, 95);
  }, [phase, trainingStart]);

  const pollStatus = useCallback(() => getStatus(taskId!), [taskId]);

  const { data: statusData } = usePolling<StatusResponse>({
    fn: pollStatus,
    interval: 3000,
    enabled: phase === "training" && !!taskId,
    shouldStop: (d) => d.status === "completed" || d.status === "failed",
    onStop: async (d) => {
      if (d.status === "completed") {
        setPhase("retrying");
        try {
          const resp = await postAnalyze({
            ticker: tickerRef.current,
            thread_id: sessionId.current,
          });
          if (resp.status === "error") {
            setError(resp.detail ?? "Analysis failed");
            setPhase("error");
          } else {
            setResult(resp);
            setPhase("done");
          }
        } catch (e) {
          setError(e instanceof Error ? e.message : "Retry failed");
          setPhase("error");
        }
      } else {
        setError("Model training failed");
        setPhase("error");
      }
    },
  });

  const trainingElapsed = statusData?.elapsed_seconds ?? 0;
  const computedProgress =
    phase === "training" ? Math.min(trainingElapsed * 2, 95) : progress;

  async function analyze(ticker: string) {
    tickerRef.current = ticker;
    setPhase("analyzing");
    setResult(null);
    setError(null);
    setTaskId(null);

    try {
      const resp = await postAnalyze({
        ticker,
        thread_id: sessionId.current,
      });

      if (resp.status === "training") {
        const tid = resp.task_id ?? ticker.toLowerCase();
        setTaskId(tid);
        setTrainingStart(Date.now());
        setPhase("training");
      } else if (resp.status === "error") {
        setError(resp.detail ?? "Analysis failed");
        setPhase("error");
      } else {
        setResult(resp);
        setPhase("done");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Connection failed");
      setPhase("error");
    }
  }

  function reset() {
    setPhase("idle");
    setResult(null);
    setError(null);
    setTaskId(null);
  }

  return {
    phase,
    result,
    error,
    progress: computedProgress,
    trainingElapsed,
    sessionId: sessionId.current,
    analyze,
    reset,
  };
}
