import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useState } from 'react';
import type { Schedule, CreateScheduleRequest, CreateScheduleDayInput } from '@/api/types';
import WeeklyScheduleEditor from './WeeklyScheduleEditor';
import { createDefaultDays } from './scheduleUtils';

const scheduleSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255),
  description: z.string().max(1000).optional(),
  timezone: z.string().min(1, 'Timezone is required'),
});

type ScheduleFormData = z.infer<typeof scheduleSchema>;

interface ScheduleFormProps {
  schedule?: Schedule;
  onSubmit: (data: CreateScheduleRequest) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

// Common timezones
const TIMEZONES = [
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'America/Phoenix',
  'America/Anchorage',
  'Pacific/Honolulu',
  'UTC',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Asia/Tokyo',
  'Asia/Shanghai',
  'Asia/Singapore',
  'Australia/Sydney',
];

export default function ScheduleForm({
  schedule,
  onSubmit,
  onCancel,
  isLoading = false,
}: ScheduleFormProps) {
  const isEditing = !!schedule;

  // Initialize days from schedule or defaults
  const [days, setDays] = useState<CreateScheduleDayInput[]>(() => {
    if (schedule?.days) {
      return schedule.days.map((d) => ({
        day_of_week: d.day_of_week,
        is_working_day: d.is_working_day,
        shift_start: d.shift_start,
        shift_end: d.shift_end,
        breaks: d.breaks?.map((b) => ({
          name: b.name,
          break_start: b.break_start,
          break_end: b.break_end,
        })) || [],
      }));
    }
    return createDefaultDays();
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ScheduleFormData>({
    resolver: zodResolver(scheduleSchema),
    defaultValues: {
      name: schedule?.name || '',
      description: schedule?.description || '',
      timezone: schedule?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone,
    },
  });

  const onFormSubmit = (data: ScheduleFormData) => {
    const request: CreateScheduleRequest = {
      name: data.name,
      description: data.description || undefined,
      timezone: data.timezone,
      days,
    };
    onSubmit(request);
  };

  return (
    <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-6">
      {/* Basic Info */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Basic Information</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Schedule Name *
            </label>
            <input
              id="name"
              type="text"
              {...register('name')}
              className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="e.g., Standard Shift"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
            )}
          </div>

          <div>
            <label htmlFor="timezone" className="block text-sm font-medium text-gray-700 mb-1">
              Timezone *
            </label>
            <select
              id="timezone"
              {...register('timezone')}
              className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {TIMEZONES.map((tz) => (
                <option key={tz} value={tz}>
                  {tz}
                </option>
              ))}
            </select>
            {errors.timezone && (
              <p className="mt-1 text-sm text-red-600">{errors.timezone.message}</p>
            )}
          </div>

          <div className="md:col-span-2">
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
              Description
            </label>
            <textarea
              id="description"
              {...register('description')}
              rows={2}
              className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Optional description"
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
            )}
          </div>
        </div>
      </div>

      {/* Weekly Schedule */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Weekly Schedule</h3>
        <WeeklyScheduleEditor days={days} onChange={setDays} disabled={isLoading} />
      </div>

      {/* Actions */}
      <div className="flex justify-end gap-3">
        <button
          type="button"
          onClick={onCancel}
          disabled={isLoading}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isLoading}
          className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {isLoading ? 'Saving...' : isEditing ? 'Update Schedule' : 'Create Schedule'}
        </button>
      </div>
    </form>
  );
}
