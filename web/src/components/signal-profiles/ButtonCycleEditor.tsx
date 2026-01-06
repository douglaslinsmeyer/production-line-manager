import { ArrowRightIcon, XMarkIcon } from '@heroicons/react/24/outline';
import type { ButtonBehavior } from '../../api/types';

interface ButtonCycleEditorProps {
  buttonBehavior: ButtonBehavior;
  availableStates: string[];
  onChange: (buttonBehavior: ButtonBehavior) => void;
}

export default function ButtonCycleEditor({
  buttonBehavior,
  availableStates,
  onChange,
}: ButtonCycleEditorProps) {
  const addToShortCycle = (stateName: string) => {
    if (!buttonBehavior.shortPressCycle.includes(stateName)) {
      onChange({
        ...buttonBehavior,
        shortPressCycle: [...buttonBehavior.shortPressCycle, stateName],
      });
    }
  };

  const removeFromShortCycle = (index: number) => {
    onChange({
      ...buttonBehavior,
      shortPressCycle: buttonBehavior.shortPressCycle.filter((_, i) => i !== index),
    });
  };

  const addToLongCycle = (stateName: string) => {
    if (!buttonBehavior.longPressCycle.includes(stateName)) {
      onChange({
        ...buttonBehavior,
        longPressCycle: [...buttonBehavior.longPressCycle, stateName],
      });
    }
  };

  const removeFromLongCycle = (index: number) => {
    onChange({
      ...buttonBehavior,
      longPressCycle: buttonBehavior.longPressCycle.filter((_, i) => i !== index),
    });
  };

  const renderCycle = (
    cycle: string[],
    onRemove: (index: number) => void,
    onAdd: (state: string) => void,
    label: string
  ) => {
    const availableToAdd = availableStates.filter((s) => !cycle.includes(s));

    return (
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">{label}</label>

        {/* Cycle Visualization */}
        <div className="flex flex-wrap items-center gap-2 mb-3 p-3 bg-gray-50 rounded-md min-h-[3rem]">
          {cycle.length === 0 ? (
            <span className="text-sm text-gray-400">No states in cycle</span>
          ) : (
            cycle.map((stateName, index) => (
              <div key={index} className="flex items-center gap-2">
                <div className="flex items-center gap-2 px-3 py-1 bg-blue-100 text-blue-800 rounded-md">
                  <span className="text-sm font-medium">{stateName}</span>
                  <button
                    type="button"
                    onClick={() => onRemove(index)}
                    className="text-blue-600 hover:text-blue-800"
                  >
                    <XMarkIcon className="w-4 h-4" />
                  </button>
                </div>
                {index < cycle.length - 1 && (
                  <ArrowRightIcon className="w-4 h-4 text-gray-400" />
                )}
              </div>
            ))
          )}
          {cycle.length > 1 && (
            <>
              <ArrowRightIcon className="w-4 h-4 text-gray-400" />
              <span className="text-sm text-gray-500">(cycles back to first)</span>
            </>
          )}
        </div>

        {/* Add State Dropdown */}
        {availableToAdd.length > 0 && (
          <div className="flex gap-2">
            <select
              onChange={(e) => {
                if (e.target.value) {
                  onAdd(e.target.value);
                  e.target.value = '';
                }
              }}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 text-sm"
              defaultValue=""
            >
              <option value="" disabled>
                Add state to cycle...
              </option>
              {availableToAdd.map((stateName) => (
                <option key={stateName} value={stateName}>
                  {stateName}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-4">
      <div className="text-sm text-gray-600 mb-4">
        Define how button presses cycle through states. Short press is released before 1 second, long
        press is held for 1+ seconds.
      </div>

      {/* Short Press Cycle */}
      {renderCycle(
        buttonBehavior.shortPressCycle,
        removeFromShortCycle,
        addToShortCycle,
        'Short Press Cycle (< 1s)'
      )}

      {/* Long Press Cycle */}
      {renderCycle(
        buttonBehavior.longPressCycle,
        removeFromLongCycle,
        addToLongCycle,
        'Long Press Cycle (≥ 1s)'
      )}
    </div>
  );
}
