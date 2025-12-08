import type { CreateScheduleDayInput, DayOfWeek } from '@/api/types';
import DayScheduleRow from './DayScheduleRow';

interface WeeklyScheduleEditorProps {
  days: CreateScheduleDayInput[];
  onChange: (days: CreateScheduleDayInput[]) => void;
  disabled?: boolean;
}

// Ensure days are sorted by day_of_week and start from Monday (1)
function sortDays(days: CreateScheduleDayInput[]): CreateScheduleDayInput[] {
  const dayOrder: DayOfWeek[] = [1, 2, 3, 4, 5, 6, 0]; // Mon-Sun
  return [...days].sort((a, b) => dayOrder.indexOf(a.day_of_week) - dayOrder.indexOf(b.day_of_week));
}

export default function WeeklyScheduleEditor({
  days,
  onChange,
  disabled = false,
}: WeeklyScheduleEditorProps) {
  const sortedDays = sortDays(days);

  const handleDayChange = (index: number, updatedDay: CreateScheduleDayInput) => {
    const newDays = [...days];
    const originalIndex = days.findIndex((d) => d.day_of_week === sortedDays[index].day_of_week);
    if (originalIndex !== -1) {
      newDays[originalIndex] = updatedDay;
      onChange(newDays);
    }
  };

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-28">
              Day
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-24">
              Working
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-32">
              Shift Start
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-32">
              Shift End
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Breaks
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {sortedDays.map((day, index) => (
            <DayScheduleRow
              key={day.day_of_week}
              day={day}
              onChange={(updatedDay) => handleDayChange(index, updatedDay)}
              disabled={disabled}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}
