import { useState, useEffect } from 'react';

/**
 * Custom hook that calculates and updates elapsed time from an ISO timestamp.
 * Returns elapsed time in seconds, updating every second.
 *
 * @param timestampISO - ISO 8601 formatted timestamp string
 * @returns Elapsed seconds since the timestamp
 */
export function useStatusTimer(timestampISO?: string): number {
  const [elapsedSeconds, setElapsedSeconds] = useState(0);

  useEffect(() => {
    if (!timestampISO) {
      setElapsedSeconds(0);
      return;
    }

    // Calculate elapsed time from timestamp to now
    const calculateElapsed = () => {
      const statusChangeTime = new Date(timestampISO).getTime();
      const now = Date.now();
      const diffMs = now - statusChangeTime;
      return Math.max(0, Math.floor(diffMs / 1000)); // Ensure non-negative
    };

    // Set initial value
    setElapsedSeconds(calculateElapsed());

    // Update every second
    const intervalId = setInterval(() => {
      setElapsedSeconds(calculateElapsed());
    }, 1000);

    // Cleanup interval on unmount or timestamp change
    return () => clearInterval(intervalId);
  }, [timestampISO]);

  return elapsedSeconds;
}
