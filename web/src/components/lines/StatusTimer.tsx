import { ClockIcon } from '@heroicons/react/24/outline';
import { useStatusHistory } from '@/hooks/useLines';
import { useStatusTimer } from '@/hooks/useStatusTimer';
import type { Status } from '@/api/types';

interface StatusTimerProps {
  lineId: string;
  currentStatus?: Status;
  className?: string;
}

/**
 * Formats elapsed seconds as HH:MM:SS
 */
function formatElapsedTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

export default function StatusTimer({ lineId, className = '' }: StatusTimerProps) {
  // Fetch only the most recent status change
  const { data: history, isLoading } = useStatusHistory(lineId, 1);

  // Calculate elapsed time from the last status change
  const lastStatusChange = history?.[0];
  const elapsedSeconds = useStatusTimer(lastStatusChange?.time);

  if (isLoading) {
    return (
      <div className={`text-sm text-gray-400 font-mono ${className}`}>
        <div className="flex items-center gap-1">
          <ClockIcon className="h-4 w-4" />
          <span>--:--:--</span>
        </div>
      </div>
    );
  }

  return (
    <div className={`text-sm text-gray-600 font-mono ${className}`}>
      <div className="flex items-center gap-1">
        <ClockIcon className="h-4 w-4" />
        <span>{formatElapsedTime(elapsedSeconds)}</span>
      </div>
    </div>
  );
}
