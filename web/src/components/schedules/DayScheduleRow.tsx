import type { DayOfWeek, CreateScheduleDayInput, CreateBreakInput } from '@/api/types';
import { DAY_LABELS } from './scheduleUtils';
import TimeInput from './TimeInput';
import BreakEditor from './BreakEditor';

interface DayScheduleRowProps {
  day: CreateScheduleDayInput;
  onChange: (day: CreateScheduleDayInput) => void;
  disabled?: boolean;
}

export default function DayScheduleRow({ day, onChange, disabled = false }: DayScheduleRowProps) {
  const handleWorkingDayChange = (isWorkingDay: boolean) => {
    if (isWorkingDay) {
      onChange({
        ...day,
        is_working_day: true,
        shift_start: day.shift_start || '08:00:00',
        shift_end: day.shift_end || '17:00:00',
        breaks: day.breaks || [],
      });
    } else {
      onChange({
        ...day,
        is_working_day: false,
        shift_start: undefined,
        shift_end: undefined,
        breaks: [],
      });
    }
  };

  const handleTimeChange = (field: 'shift_start' | 'shift_end', value: string) => {
    onChange({ ...day, [field]: value });
  };

  const handleBreaksChange = (breaks: CreateBreakInput[]) => {
    onChange({ ...day, breaks });
  };

  return (
    <tr className={day.is_working_day ? 'bg-white' : 'bg-gray-50'}>
      {/* Day Name */}
      <td className="px-4 py-3 whitespace-nowrap">
        <span className="font-medium text-gray-900">{DAY_LABELS[day.day_of_week as DayOfWeek]}</span>
      </td>

      {/* Working Day Toggle */}
      <td className="px-4 py-3 whitespace-nowrap">
        <label className="inline-flex items-center cursor-pointer">
          <input
            type="checkbox"
            checked={day.is_working_day}
            onChange={(e) => handleWorkingDayChange(e.target.checked)}
            disabled={disabled}
            className="sr-only peer"
          />
          <div className="relative w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
        </label>
      </td>

      {/* Shift Start */}
      <td className="px-4 py-3 whitespace-nowrap">
        <TimeInput
          value={day.shift_start || ''}
          onChange={(v) => handleTimeChange('shift_start', v)}
          disabled={disabled || !day.is_working_day}
        />
      </td>

      {/* Shift End */}
      <td className="px-4 py-3 whitespace-nowrap">
        <TimeInput
          value={day.shift_end || ''}
          onChange={(v) => handleTimeChange('shift_end', v)}
          disabled={disabled || !day.is_working_day}
        />
      </td>

      {/* Breaks */}
      <td className="px-4 py-3">
        {day.is_working_day ? (
          <BreakEditor
            breaks={day.breaks || []}
            onChange={handleBreaksChange}
            disabled={disabled}
          />
        ) : (
          <span className="text-sm text-gray-400 italic">No breaks (non-working day)</span>
        )}
      </td>
    </tr>
  );
}
