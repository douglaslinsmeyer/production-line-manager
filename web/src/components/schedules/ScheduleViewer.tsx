import type { Schedule, ScheduleDay, DayOfWeek } from '@/api/types';
import { DAY_LABELS, formatTime } from './scheduleUtils';

interface ScheduleViewerProps {
  schedule: Schedule;
}

// Sort days to start from Monday
function sortDays(days: ScheduleDay[]): ScheduleDay[] {
  const dayOrder: DayOfWeek[] = [1, 2, 3, 4, 5, 6, 0]; // Mon-Sun
  return [...days].sort((a, b) => dayOrder.indexOf(a.day_of_week) - dayOrder.indexOf(b.day_of_week));
}

export default function ScheduleViewer({ schedule }: ScheduleViewerProps) {
  const sortedDays = sortDays(schedule.days);

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-28">
              Day
            </th>
            <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider w-24">
              Status
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
          {sortedDays.map((day) => (
            <tr key={day.id} className={day.is_working_day ? 'bg-white' : 'bg-gray-50'}>
              <td className="px-4 py-3 whitespace-nowrap">
                <span className="font-medium text-gray-900">
                  {DAY_LABELS[day.day_of_week]}
                </span>
              </td>
              <td className="px-4 py-3 whitespace-nowrap">
                {day.is_working_day ? (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                    Working
                  </span>
                ) : (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
                    Off
                  </span>
                )}
              </td>
              <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                {day.is_working_day ? formatTime(day.shift_start) : '—'}
              </td>
              <td className="px-4 py-3 whitespace-nowrap text-sm text-gray-900">
                {day.is_working_day ? formatTime(day.shift_end) : '—'}
              </td>
              <td className="px-4 py-3 text-sm text-gray-900">
                {day.is_working_day && day.breaks && day.breaks.length > 0 ? (
                  <div className="space-y-1">
                    {day.breaks.map((brk) => (
                      <span
                        key={brk.id}
                        className="inline-flex items-center px-2 py-0.5 rounded text-xs bg-blue-50 text-blue-700 mr-2"
                      >
                        {brk.name}: {formatTime(brk.break_start)} - {formatTime(brk.break_end)}
                      </span>
                    ))}
                  </div>
                ) : day.is_working_day ? (
                  <span className="text-gray-400 italic">No breaks</span>
                ) : (
                  <span className="text-gray-400">—</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
