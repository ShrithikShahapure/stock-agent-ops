import { useEffect, useRef, useState } from "react";

interface UsePollingOptions<T> {
  fn: () => Promise<T>;
  interval?: number;
  enabled: boolean;
  shouldStop: (data: T) => boolean;
  onStop?: (data: T) => void;
}

export function usePolling<T>({
  fn,
  interval = 3000,
  enabled,
  shouldStop,
  onStop,
}: UsePollingOptions<T>) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const stoppedRef = useRef(false);

  useEffect(() => {
    if (!enabled) {
      if (timerRef.current) clearInterval(timerRef.current);
      stoppedRef.current = false;
      return;
    }

    stoppedRef.current = false;

    async function poll() {
      if (stoppedRef.current) return;
      try {
        const result = await fn();
        setData(result);
        setError(null);
        if (shouldStop(result)) {
          stoppedRef.current = true;
          if (timerRef.current) clearInterval(timerRef.current);
          onStop?.(result);
        }
      } catch (e) {
        setError(e instanceof Error ? e.message : "Polling error");
      }
    }

    poll();
    timerRef.current = setInterval(poll, interval);

    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [enabled]); // eslint-disable-line react-hooks/exhaustive-deps

  return { data, error };
}
