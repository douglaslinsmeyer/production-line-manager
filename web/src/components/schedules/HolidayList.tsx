import { useState } from 'react';
import { PlusIcon, TrashIcon, PencilIcon } from '@heroicons/react/24/outline';
import type { ScheduleHoliday } from '@/api/types';
import { formatDate } from './scheduleUtils';
import HolidayModal from './HolidayModal';

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
  const [showModal, setShowModal] = useState(false);
  const [editingHoliday, setEditingHoliday] = useState<ScheduleHoliday | null>(null);

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

  // Sort holidays by date
  const sortedHolidays = [...holidays].sort((a, b) => a.holiday_date.localeCompare(b.holiday_date));

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h4 className="text-sm font-medium text-gray-700">Holidays</h4>
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

      {sortedHolidays.length === 0 ? (
        <p className="text-sm text-gray-500 italic">No holidays configured</p>
      ) : (
        <div className="space-y-2">
          {sortedHolidays.map((holiday) => (
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
