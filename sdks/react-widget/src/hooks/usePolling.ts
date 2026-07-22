import { useEffect, useRef, useCallback } from 'react';

interface UsePollingOptions {
  fn: () => Promise<boolean>;
  interval: number;
  enabled: boolean;
  maxAttempts?: number;
}

export function usePolling({
  fn,
  interval,
  enabled,
  maxAttempts = 40,
}: UsePollingOptions) {
  const attempts = useRef(0);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  const poll = useCallback(async () => {
    if (attempts.current >= maxAttempts) return;
    attempts.current++;

    const done = await fn();
    if (!done) {
      timerRef.current = setTimeout(poll, interval);
    }
  }, [fn, interval, maxAttempts]);

  useEffect(() => {
    if (!enabled) return;
    attempts.current = 0;
    poll();
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [enabled, poll]);
}
