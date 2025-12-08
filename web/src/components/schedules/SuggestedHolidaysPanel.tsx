import { PlusIcon, CheckIcon } from '@heroicons/react/24/outline';
import type { SuggestedHoliday } from '@/api/types';
import { formatDate } from './scheduleUtils';

interface SuggestedHolidaysPanelProps {
  suggestions: SuggestedHoliday[];
  existingDates: Set<string>;
  onAdd: (holiday: SuggestedHoliday) => void;
  isLoading: boolean;
  error?: string;
  cached?: boolean;
}

export default function SuggestedHolidaysPanel({
  suggestions,
  existingDates,
  onAdd,
  isLoading,
  error,
  cached,
}: SuggestedHolidaysPanelProps) {
  if (isLoading) {
    return (
      <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <div className="flex items-center gap-2 text-blue-700">
          <div className="animate-spin h-4 w-4 border-2 border-blue-600 border-t-transparent rounded-full" />
          <span className="text-sm">Loading suggested holidays...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mb-6 p-4 bg-amber-50 rounded-lg border border-amber-200">
        <p className="text-sm text-amber-700">Unable to load suggested holidays: {error}</p>
      </div>
    );
  }

  if (suggestions.length === 0) {
    return (
      <div className="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
        <p className="text-sm text-gray-500 italic">No suggested holidays available for this year</p>
      </div>
    );
  }

  return (
    <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
      <div className="flex items-center justify-between mb-3">
        <h5 className="text-sm font-medium text-blue-800">
          Suggested Public Holidays
          {cached && <span className="ml-2 text-xs text-blue-600">(cached)</span>}
        </h5>
        <span className="text-xs text-blue-600">{suggestions.length} holidays</span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
        {suggestions.map((holiday) => {
          const isAdded = existingDates.has(holiday.date);
          return (
            <div
              key={holiday.date}
              className="flex items-center justify-between p-2 bg-white rounded border border-blue-100"
            >
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium text-gray-900 truncate">{holiday.name}</p>
                <p className="text-xs text-gray-500">{formatDate(holiday.date)}</p>
              </div>
              {isAdded ? (
                <div className="flex items-center gap-1 text-green-600 ml-2">
                  <CheckIcon className="h-4 w-4" />
                  <span className="text-xs">Added</span>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => onAdd(holiday)}
                  className="flex items-center gap-1 text-blue-600 hover:text-blue-700 text-sm ml-2"
                >
                  <PlusIcon className="h-4 w-4" />
                  <span className="text-xs">Add</span>
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
