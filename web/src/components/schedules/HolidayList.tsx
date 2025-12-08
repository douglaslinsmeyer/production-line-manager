import { useState, useMemo } from 'react';
import { PlusIcon, TrashIcon, PencilIcon, SparklesIcon, ChevronDownIcon } from '@heroicons/react/24/outline';
import type { ScheduleHoliday, SuggestedHoliday } from '@/api/types';
import { formatDate } from './scheduleUtils';
import HolidayModal from './HolidayModal';
import SuggestedHolidaysPanel from './SuggestedHolidaysPanel';
import { useSuggestedHolidays } from '@/hooks/useSchedules';

interface HolidayListProps {
  holidays: ScheduleHoliday[];
  scheduleId: string;
  onAdd: (data: { holiday_date: string; name: string }) => void;
  onUpdate: (holidayId: string, data: { holiday_date?: string; name?: string }) => void;
  onDelete: (holidayId: string) => void;
  isLoading?: boolean;
}

export default function HolidayList({
  holidays,
  onAdd,
  onUpdate,
  onDelete,
  isLoading = false,
}: HolidayListProps) {
  const currentYear = new Date().getFullYear();
  const [selectedYear, setSelectedYear] = useState<number>(currentYear);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [editingHoliday, setEditingHoliday] = useState<ScheduleHoliday | null>(null);

  // Year options: current year and next year
  const yearOptions = [currentYear, currentYear + 1];

  // Get suggested holidays when panel is shown
  const { data: suggestedData, isLoading: suggestionsLoading } = useSuggestedHolidays(
    selectedYear,
    showSuggestions
  );

  // Create Set of existing holiday dates for quick lookup
  const existingDates = useMemo(() => {
    return new Set(holidays.map((h) => h.holiday_date));
  }, [holidays]);

  // Filter holidays by selected year
  const filteredHolidays = useMemo(() => {
    return holidays
      .filter((h) => h.holiday_date.startsWith(`${selectedYear}-`))
      .sort((a, b) => a.holiday_date.localeCompare(b.holiday_date));
  }, [holidays, selectedYear]);

  const handleAdd = () => {
    setEditingHoliday(null);
    setShowModal(true);
  };

  const handleEdit = (holiday: ScheduleHoliday) => {
    setEditingHoliday(holiday);
    setShowModal(true);
  };

  const handleSubmit = (data: { holiday_date: string; name: string }) => {
    if (editingHoliday) {
      onUpdate(editingHoliday.id, data);
    } else {
      onAdd(data);
    }
    setShowModal(false);
    setEditingHoliday(null);
  };

  const handleDelete = (holidayId: string) => {
    if (window.confirm('Are you sure you want to delete this holiday?')) {
      onDelete(holidayId);
    }
  };

  // Handler to add suggested holiday
  const handleAddSuggested = (suggested: SuggestedHoliday) => {
    onAdd({ holiday_date: suggested.date, name: suggested.name });
  };

  return (
    <div>
      {/* Header with Year Selector */}
      <div className="flex flex-wrap items-center justify-between gap-3 mb-4">
        <div className="flex items-center gap-3">
          <h4 className="text-sm font-medium text-gray-700">Holidays</h4>
          <div className="relative">
            <select
              value={selectedYear}
              onChange={(e) => setSelectedYear(Number(e.target.value))}
              className="appearance-none bg-white border border-gray-300 rounded-md pl-3 pr-8 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              {yearOptions.map((year) => (
                <option key={year} value={year}>
                  {year}
                </option>
              ))}
            </select>
            <ChevronDownIcon className="absolute right-2 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 pointer-events-none" />
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setShowSuggestions(!showSuggestions)}
            className={`flex items-center gap-1 text-sm px-3 py-1.5 rounded-md border transition-colors ${
              showSuggestions
                ? 'bg-blue-50 border-blue-200 text-blue-700'
                : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
            }`}
          >
            <SparklesIcon className="h-4 w-4" />
            {showSuggestions ? 'Hide Suggestions' : 'Show Suggestions'}
          </button>
          <button
            type="button"
            onClick={handleAdd}
            disabled={isLoading}
            className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 disabled:opacity-50"
          >
            <PlusIcon className="h-4 w-4" />
            Add Holiday
          </button>
        </div>
      </div>

      {/* Suggested Holidays Section (collapsible) */}
      {showSuggestions && (
        <SuggestedHolidaysPanel
          suggestions={suggestedData?.holidays || []}
          existingDates={existingDates}
          onAdd={handleAddSuggested}
          isLoading={suggestionsLoading}
          error={suggestedData?.error}
          cached={suggestedData?.cached}
        />
      )}

      {/* Existing Holidays List */}
      {filteredHolidays.length === 0 ? (
        <p className="text-sm text-gray-500 italic">No holidays configured for {selectedYear}</p>
      ) : (
        <div className="space-y-2">
          {filteredHolidays.map((holiday) => (
            <div
              key={holiday.id}
              className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div>
                <p className="font-medium text-gray-900">{holiday.name}</p>
                <p className="text-sm text-gray-500">{formatDate(holiday.holiday_date)}</p>
              </div>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => handleEdit(holiday)}
                  disabled={isLoading}
                  className="p-1 text-gray-400 hover:text-blue-600 rounded disabled:opacity-50"
                >
                  <PencilIcon className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  onClick={() => handleDelete(holiday.id)}
                  disabled={isLoading}
                  className="p-1 text-gray-400 hover:text-red-600 rounded disabled:opacity-50"
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      <HolidayModal
        isOpen={showModal}
        onClose={() => {
          setShowModal(false);
          setEditingHoliday(null);
        }}
        onSubmit={handleSubmit}
        holiday={editingHoliday}
        isLoading={isLoading}
      />
    </div>
  );
}
