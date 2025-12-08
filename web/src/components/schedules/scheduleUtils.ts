import type { DayOfWeek } from '@/api/types';

// Day of week labels
export const DAY_LABELS: Record<DayOfWeek, string> = {
  0: 'Sunday',
  1: 'Monday',
  2: 'Tuesday',
  3: 'Wednesday',
  4: 'Thursday',
  5: 'Friday',
  6: 'Saturday',
};

export const DAY_SHORT_LABELS: Record<DayOfWeek, string> = {
  0: 'Sun',
  1: 'Mon',
  2: 'Tue',
  3: 'Wed',
  4: 'Thu',
  5: 'Fri',
  6: 'Sat',
};

// Convert HH:MM:SS to display format HH:MM
export function formatTime(time: string | undefined): string {
  if (!time) return '--:--';
  const parts = time.split(':');
  if (parts.length >= 2) {
    return `${parts[0]}:${parts[1]}`;
  }
  return time;
}

// Convert display time HH:MM to storage format HH:MM:SS
export function toStorageTime(time: string): string {
  if (!time) return '';
  const parts = time.split(':');
  if (parts.length === 2) {
    return `${parts[0]}:${parts[1]}:00`;
  }
  return time;
}

// Format date for display
export function formatDate(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  return date.toLocaleDateString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

// Create default days array for new schedule
export function createDefaultDays() {
  const defaultWorkingDays: DayOfWeek[] = [1, 2, 3, 4, 5]; // Mon-Fri
  return Array.from({ length: 7 }, (_, i) => {
    const dayOfWeek = i as DayOfWeek;
    const isWorkingDay = defaultWorkingDays.includes(dayOfWeek);
    return {
      day_of_week: dayOfWeek,
      is_working_day: isWorkingDay,
      shift_start: isWorkingDay ? '08:00:00' : undefined,
      shift_end: isWorkingDay ? '17:00:00' : undefined,
      breaks: isWorkingDay ? [{ name: 'Lunch', break_start: '12:00:00', break_end: '13:00:00' }] : [],
    };
  });
}

// Validate time range
export function isValidTimeRange(start: string, end: string): boolean {
  if (!start || !end) return false;
  return start < end;
}

// Calculate hours between two times
export function calculateHours(start: string, end: string): number {
  if (!start || !end) return 0;
  const [startH, startM] = start.split(':').map(Number);
  const [endH, endM] = end.split(':').map(Number);
  const startMinutes = startH * 60 + startM;
  const endMinutes = endH * 60 + endM;
  return (endMinutes - startMinutes) / 60;
}
