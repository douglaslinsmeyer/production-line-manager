import { Fragment, useState } from 'react';
import { Listbox, Transition } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid';
import type { Label } from '@/api/types';

interface LabelFilterProps {
  selectedLabels: Label[];
  onLabelsChange: (labels: Label[]) => void;
  availableLabels: Label[];
}

export default function LabelFilter({
  selectedLabels,
  onLabelsChange,
  availableLabels,
}: LabelFilterProps) {
  const [searchQuery, setSearchQuery] = useState('');

  const filteredLabels = searchQuery
    ? availableLabels.filter((label) =>
        label.name.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : availableLabels;

  const handleSelectAll = () => {
    onLabelsChange(availableLabels);
  };

  const handleClearAll = () => {
    onLabelsChange([]);
  };

  const handleToggleLabel = (label: Label) => {
    const isSelected = selectedLabels.some((l) => l.id === label.id);
    if (isSelected) {
      onLabelsChange(selectedLabels.filter((l) => l.id !== label.id));
    } else {
      onLabelsChange([...selectedLabels, label]);
    }
  };

  return (
    <Listbox value={selectedLabels} onChange={onLabelsChange} multiple>
      <div className="relative">
        <Listbox.Button className="relative w-full cursor-pointer rounded-lg bg-white py-2 pl-3 pr-10 text-left border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm">
          <span className="block truncate">
            {selectedLabels.length === 0
              ? 'Filter by labels'
              : `${selectedLabels.length} label${selectedLabels.length > 1 ? 's' : ''} selected`}
          </span>
          <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
            <ChevronUpDownIcon className="h-5 w-5 text-gray-400" aria-hidden="true" />
          </span>
        </Listbox.Button>

        <Transition
          as={Fragment}
          leave="transition ease-in duration-100"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <Listbox.Options className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {/* Search input */}
            <div className="sticky top-0 bg-white px-2 py-2 border-b border-gray-200">
              <input
                type="text"
                className="w-full rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                placeholder="Search labels..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onClick={(e) => e.stopPropagation()}
              />
            </div>

            {/* Select All / Clear All buttons */}
            <div className="sticky top-[52px] bg-white px-2 py-2 border-b border-gray-200 flex gap-2">
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  handleSelectAll();
                }}
                className="flex-1 px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 rounded"
              >
                Select All
              </button>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  handleClearAll();
                }}
                className="flex-1 px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-50 rounded"
              >
                Clear All
              </button>
            </div>

            {/* Label options */}
            {filteredLabels.length === 0 ? (
              <div className="px-3 py-2 text-sm text-gray-500">No labels found</div>
            ) : (
              filteredLabels.map((label) => {
                const isSelected = selectedLabels.some((l) => l.id === label.id);
                return (
                  <Listbox.Option
                    key={label.id}
                    value={label}
                    className={({ active }) =>
                      `relative cursor-pointer select-none py-2 pl-10 pr-4 ${
                        active ? 'bg-blue-50 text-blue-900' : 'text-gray-900'
                      }`
                    }
                    onClick={(e) => {
                      e.stopPropagation();
                      handleToggleLabel(label);
                    }}
                  >
                    {() => (
                      <>
                        <span className={`block truncate ${isSelected ? 'font-semibold' : 'font-normal'}`}>
                          {label.name}
                        </span>
                        {isSelected && (
                          <span className="absolute inset-y-0 left-0 flex items-center pl-3 text-blue-600">
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        )}
                      </>
                    )}
                  </Listbox.Option>
                );
              })
            )}
          </Listbox.Options>
        </Transition>
      </div>
    </Listbox>
  );
}
