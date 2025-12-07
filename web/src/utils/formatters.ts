import { format, formatDistanceToNow, parseISO } from 'date-fns';

// Format ISO date string to readable format
export function formatDate(dateString: string): string {
  try {
    const date = parseISO(dateString);
    return format(date, 'MMM d, yyyy h:mm a');
  } catch {
    return dateString;
  }
}

// Format date as relative time (e.g., "2 hours ago")
export function formatRelativeTime(dateString: string): string {
  try {
    const date = parseISO(dateString);
    return formatDistanceToNow(date, { addSuffix: true });
  } catch {
    return dateString;
  }
}

// Format duration in milliseconds to human-readable string
export function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
}

// Calculate uptime percentage from status history
export function calculateUptime(history: { new_status: string; time: string }[]): number {
  if (history.length === 0) return 0;

  // Sort by time (newest first)
  const sorted = [...history].sort((a, b) =>
    new Date(b.time).getTime() - new Date(a.time).getTime()
  );

  let totalTime = 0;
  let uptimeMs = 0;

  for (let i = 0; i < sorted.length - 1; i++) {
    const current = sorted[i];
    const next = sorted[i + 1];

    const currentTime = new Date(current.time).getTime();
    const nextTime = new Date(next.time).getTime();
    const duration = currentTime - nextTime;

    totalTime += duration;
    if (next.new_status === 'on') {
      uptimeMs += duration;
    }
  }

  if (totalTime === 0) return 0;
  return (uptimeMs / totalTime) * 100;
}
