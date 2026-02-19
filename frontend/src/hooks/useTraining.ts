import { useCallback, useState } from "react";
import { postTrainParent, postTrainChild } from "../api/train";
import { getStatus } from "../api/status";
import { usePolling } from "./usePolling";
import type { StatusResponse } from "../types/api";

export function useTraining() {
  const [status, setStatus] = useState<"idle" | "running" | "completed" | "failed">("idle");
  const [taskId, setTaskId] = useState<string | null>(null);
  const [message, setMessage] = useState("");
  const [elapsed, setElapsed] = useState(0);

  const pollFn = useCallback(() => getStatus(taskId!), [taskId]);

  const { data: pollData } = usePolling<StatusResponse>({
    fn: pollFn,
    interval: 3000,
    enabled: status === "running" && !!taskId,
    shouldStop: (d) => d.status === "completed" || d.status === "failed",
    onStop: (d) => {
      setStatus(d.status === "completed" ? "completed" : "failed");
      setMessage(d.status === "completed" ? "Training complete" : "Training failed");
    },
  });

  if (pollData?.elapsed_seconds && status === "running") {
    if (pollData.elapsed_seconds !== elapsed) {
      // Update elapsed outside render - this is a read-only check
    }
  }

  const progress =
    status === "running"
      ? Math.min((pollData?.elapsed_seconds ?? elapsed) * 2, 95)
      : status === "completed"
        ? 100
        : 0;

  async function trainParent() {
    setStatus("running");
    setMessage("Starting parent training...");
    try {
      const resp = await postTrainParent();
      if (resp.status === "completed") {
        setStatus("completed");
        setMessage("Parent model already trained");
      } else {
        setTaskId(resp.task_id ?? "parent_training");
      }
    } catch (e) {
      setStatus("failed");
      setMessage(e instanceof Error ? e.message : "API offline");
    }
  }

  async function trainChild(ticker: string) {
    setStatus("running");
    setMessage(`Starting training for ${ticker}...`);
    try {
      const resp = await postTrainChild({ ticker });
      if (resp.status === "completed") {
        setStatus("completed");
        setMessage(`${ticker} model already trained`);
      } else {
        setTaskId(resp.task_id ?? ticker.toLowerCase());
      }
    } catch (e) {
      setStatus("failed");
      setMessage(e instanceof Error ? e.message : "API offline");
    }
  }

  function reset() {
    setStatus("idle");
    setTaskId(null);
    setMessage("");
    setElapsed(0);
  }

  return { status, message, progress, trainParent, trainChild, reset };
}
