import { PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import type { CreateBreakInput } from '@/api/types';
import TimeInput from './TimeInput';

interface BreakEditorProps {
  breaks: CreateBreakInput[];
  onChange: (breaks: CreateBreakInput[]) => void;
  disabled?: boolean;
}

export default function BreakEditor({ breaks, onChange, disabled = false }: BreakEditorProps) {
  const addBreak = () => {
    onChange([
      ...breaks,
      { name: `Break ${breaks.length + 1}`, break_start: '12:00:00', break_end: '12:30:00' },
    ]);
  };

  const removeBreak = (index: number) => {
    onChange(breaks.filter((_, i) => i !== index));
  };

  const updateBreak = (index: number, field: keyof CreateBreakInput, value: string) => {
    const updated = [...breaks];
    updated[index] = { ...updated[index], [field]: value };
    onChange(updated);
  };

  return (
    <div className="space-y-2">
      {breaks.map((brk, index) => (
        <div key={index} className="flex items-center gap-2">
          <input
            type="text"
            value={brk.name}
            onChange={(e) => updateBreak(index, 'name', e.target.value)}
            disabled={disabled}
            placeholder="Break name"
            className="flex-1 px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm disabled:bg-gray-100"
          />
          <TimeInput
            value={brk.break_start}
            onChange={(v) => updateBreak(index, 'break_start', v)}
            disabled={disabled}
            className="w-28"
          />
          <span className="text-gray-400">-</span>
          <TimeInput
            value={brk.break_end}
            onChange={(v) => updateBreak(index, 'break_end', v)}
            disabled={disabled}
            className="w-28"
          />
          {!disabled && (
            <button
              type="button"
              onClick={() => removeBreak(index)}
              className="p-1 text-red-500 hover:text-red-700 hover:bg-red-50 rounded"
            >
              <TrashIcon className="h-4 w-4" />
            </button>
          )}
        </div>
      ))}
      {!disabled && (
        <button
          type="button"
          onClick={addBreak}
          className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700"
        >
          <PlusIcon className="h-4 w-4" />
          Add Break
        </button>
      )}
    </div>
  );
}
