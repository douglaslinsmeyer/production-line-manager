import { useState } from 'react';
import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import type { LineScheduleException, CreateLineScheduleExceptionRequest, ProductionLine } from '@/api/types';
import { formatDate } from './scheduleUtils';
import LineExceptionModal from './LineExceptionModal';

interface LineExceptionListProps {
  exceptions: LineScheduleException[];
  scheduleId: string;
  availableLines: ProductionLine[];
  onAdd: (data: CreateLineScheduleExceptionRequest) => void;
  onDelete: (exceptionId: string) => void;
  isLoading?: boolean;
}

export default function LineExceptionList({
  exceptions,
  availableLines,
  onAdd,
  onDelete,
  isLoading = false,
}: LineExceptionListProps) {
  const [showModal, setShowModal] = useState(false);

  const handleAdd = () => {
    setShowModal(true);
  };

  const handleSubmit = (data: CreateLineScheduleExceptionRequest) => {
    onAdd(data);
    setShowModal(false);
  };

  const handleDelete = (exceptionId: string) => {
    if (window.confirm('Are you sure you want to delete this line exception?')) {
      onDelete(exceptionId);
    }
  };

  // Create a map of line ID to line for quick lookup
  const lineMap = new Map(availableLines.map(line => [line.id, line]));

  // Get line names for display
  const getLineNames = (lineIds: string[]): string => {
    return lineIds
      .map(id => lineMap.get(id)?.name || lineMap.get(id)?.code || 'Unknown')
      .join(', ');
  };

  // Sort exceptions by start date
  const sortedExceptions = [...exceptions].sort((a, b) => a.start_date.localeCompare(b.start_date));

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <div>
          <h4 className="text-sm font-medium text-gray-700">Line-Specific Exceptions</h4>
          <p className="text-xs text-gray-500">Apply to selected lines only</p>
        </div>
        <button
          type="button"
          onClick={handleAdd}
          disabled={isLoading || availableLines.length === 0}
          className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700 disabled:opacity-50"
        >
          <PlusIcon className="h-4 w-4" />
          Add Line Exception
        </button>
      </div>

      {availableLines.length === 0 ? (
        <p className="text-sm text-gray-500 italic">No lines assigned to this schedule yet</p>
      ) : sortedExceptions.length === 0 ? (
        <p className="text-sm text-gray-500 italic">No line-specific exceptions configured</p>
      ) : (
        <div className="space-y-3">
          {sortedExceptions.map((exception) => (
            <div
              key={exception.id}
              className="p-4 bg-amber-50 rounded-lg border border-amber-200"
            >
              <div className="flex items-start justify-between">
                <div>
                  <p className="font-medium text-gray-900">{exception.name}</p>
                  <p className="text-sm text-gray-500">
                    {formatDate(exception.start_date)} - {formatDate(exception.end_date)}
                  </p>
                  <p className="text-xs text-amber-700 mt-1">
                    Lines: {getLineNames(exception.line_ids)}
                  </p>
                  <p className="text-xs text-gray-400 mt-1">
                    {exception.days.filter(d => d.is_working_day).length} working days configured
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => handleDelete(exception.id)}
                    disabled={isLoading}
                    className="p-1 text-gray-400 hover:text-red-600 rounded disabled:opacity-50"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      <LineExceptionModal
        isOpen={showModal}
        onClose={() => setShowModal(false)}
        onSubmit={handleSubmit}
        availableLines={availableLines}
        isLoading={isLoading}
      />
    </div>
  );
}
