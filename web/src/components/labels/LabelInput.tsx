import { useState } from 'react';
import { Combobox } from '@headlessui/react';
import { PlusIcon } from '@heroicons/react/20/solid';
import { useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import LabelBadge from './LabelBadge';
import { labelsApi } from '@/api/lines';
import { lineKeys } from '@/hooks/useLines';
import type { Label } from '@/api/types';

interface LabelInputProps {
  value: Label[];
  onChange: (labels: Label[]) => void;
  availableLabels?: Label[];
  placeholder?: string;
  maxLabels?: number;
}

export default function LabelInput({
  value = [],
  onChange,
  availableLabels = [],
  placeholder = 'Add labels...',
  maxLabels,
}: LabelInputProps) {
  const [query, setQuery] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const queryClient = useQueryClient();

  // Filter suggestions based on query
  const filteredLabels =
    query === ''
      ? availableLabels
      : availableLabels.filter((label) =>
          label.name.toLowerCase().includes(query.toLowerCase())
        );

  // Filter out already selected labels
  const suggestions = filteredLabels.filter(
    (label) => !value.some((v) => v.id === label.id)
  );

  const handleAddLabel = (label: Label | null) => {
    if (!label) return;
    if (maxLabels && value.length >= maxLabels) return;
    if (value.some((v) => v.id === label.id)) return;

    onChange([...value, label]);
    setQuery('');
  };

  const handleRemoveLabel = (labelId: string) => {
    onChange(value.filter((l) => l.id !== labelId));
  };

  const handleAddTypedLabel = async () => {
    const trimmed = query.trim();
    if (!trimmed) return;
    if (isCreating) return;
    if (maxLabels && value.length >= maxLabels) {
      toast.error(`Maximum ${maxLabels} labels allowed`);
      return;
    }

    // Check if already selected
    if (value.some((v) => v.name.toLowerCase() === trimmed.toLowerCase())) {
      setQuery('');
      toast.error('Label already added');
      return;
    }

    // Check if exists in available labels
    const existing = availableLabels.find(
      (l) => l.name.toLowerCase() === trimmed.toLowerCase()
    );

    if (existing) {
      // Add existing label
      if (!value.some((v) => v.id === existing.id)) {
        onChange([...value, existing]);
      }
      setQuery('');
    } else {
      // Create new label via API
      try {
        setIsCreating(true);
        const newLabel = await labelsApi.createLabel({ name: trimmed });

        // Invalidate labels cache to refresh dropdown
        queryClient.invalidateQueries({ queryKey: lineKeys.labels() });

        // Add to selection
        onChange([...value, newLabel]);
        setQuery('');
        toast.success(`Label "${trimmed}" created`);
      } catch (error) {
        console.error('Failed to create label:', error);
        toast.error('Failed to create label');
      } finally {
        setIsCreating(false);
      }
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddTypedLabel();
    } else if (e.key === 'Escape') {
      setQuery('');
    }
  };

  return (
    <div className="space-y-2">
      <Combobox value={null} onChange={handleAddLabel}>
        <div className="relative">
          <div className="relative">
            <Combobox.Input
              className="w-full rounded-lg border border-gray-300 py-2 pl-3 pr-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              placeholder={placeholder}
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={handleKeyDown}
              disabled={isCreating}
            />
            <button
              type="button"
              onClick={handleAddTypedLabel}
              disabled={isCreating || !query.trim()}
              className="absolute inset-y-0 right-0 flex items-center pr-2 disabled:opacity-50"
              title="Add label (or press Enter)"
            >
              {isCreating ? (
                <div className="animate-spin h-5 w-5 border-2 border-blue-600 border-t-transparent rounded-full" />
              ) : (
                <PlusIcon className="h-5 w-5 text-gray-400 hover:text-gray-600" aria-hidden="true" />
              )}
            </button>
          </div>

          {suggestions.length > 0 && query !== '' && (
            <Combobox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
              {suggestions.map((label) => (
                <Combobox.Option
                  key={label.id}
                  value={label}
                  className={({ active }) =>
                    `relative cursor-pointer select-none py-2 px-3 ${
                      active ? 'bg-blue-50 text-blue-900' : 'text-gray-900'
                    }`
                  }
                >
                  {({ selected }) => (
                    <div className="flex items-center justify-between">
                      <span className={selected ? 'font-semibold' : 'font-normal'}>
                        {label.name}
                      </span>
                      {label.description && (
                        <span className="text-xs text-gray-500 ml-2">
                          {label.description}
                        </span>
                      )}
                    </div>
                  )}
                </Combobox.Option>
              ))}
            </Combobox.Options>
          )}
        </div>
      </Combobox>

      {/* Display current labels */}
      {value.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {value.map((label) => (
            <LabelBadge
              key={label.id}
              label={label}
              onRemove={() => handleRemoveLabel(label.id)}
            />
          ))}
        </div>
      )}

      {maxLabels && (
        <p className="text-xs text-gray-500">
          {value.length} / {maxLabels} labels
        </p>
      )}
    </div>
  );
}
