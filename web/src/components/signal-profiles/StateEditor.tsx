import { TrashIcon } from '@heroicons/react/24/outline';
import type { ProfileState, LightMode, BuzzerMode } from '../../api/types';
import StateOutputPreview from './StateOutputPreview';

interface StateEditorProps {
  state: ProfileState;
  index: number;
  onChange: (index: number, state: ProfileState) => void;
  onRemove: (index: number) => void;
  canRemove: boolean;
}

const lightModeOptions: { value: LightMode; label: string }[] = [
  { value: 'off', label: 'Off' },
  { value: 'on', label: 'On (Steady)' },
  { value: 'shortBlink', label: 'Short Blink (500ms)' },
  { value: 'longBlink', label: 'Long Blink (1500ms)' },
];

const buzzerModeOptions: { value: BuzzerMode; label: string }[] = [
  { value: 'off', label: 'Off' },
  { value: 'on', label: 'On (Continuous)' },
  { value: 'chirp', label: 'Chirp Pattern' },
];

export default function StateEditor({
  state,
  index,
  onChange,
  onRemove,
  canRemove,
}: StateEditorProps) {
  const handleFieldChange = (field: string, value: string) => {
    if (field === 'name') {
      onChange(index, { ...state, name: value });
    } else {
      onChange(index, {
        ...state,
        outputs: { ...state.outputs, [field]: value },
      });
    }
  };

  return (
    <div className="border border-gray-200 rounded-lg p-4 bg-white">
      <div className="flex items-start gap-4">
        {/* Preview */}
        <div className="flex-shrink-0">
          <StateOutputPreview outputs={state.outputs} />
        </div>

        {/* Configuration */}
        <div className="flex-1 space-y-3">
          {/* State Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              State Name
            </label>
            <input
              type="text"
              value={state.name}
              onChange={(e) => handleFieldChange('name', e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="e.g., On, Off, Maintenance"
            />
          </div>

          {/* Output Controls Grid */}
          <div className="grid grid-cols-2 gap-3">
            {/* Red Light */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Red Light
              </label>
              <select
                value={state.outputs.redLight}
                onChange={(e) => handleFieldChange('redLight', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              >
                {lightModeOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Yellow Light */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Yellow Light
              </label>
              <select
                value={state.outputs.yellowLight}
                onChange={(e) => handleFieldChange('yellowLight', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              >
                {lightModeOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Green Light */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Green Light
              </label>
              <select
                value={state.outputs.greenLight}
                onChange={(e) => handleFieldChange('greenLight', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              >
                {lightModeOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Buzzer */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Buzzer
              </label>
              <select
                value={state.outputs.buzzer}
                onChange={(e) => handleFieldChange('buzzer', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              >
                {buzzerModeOptions.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Remove Button */}
        {canRemove && (
          <button
            type="button"
            onClick={() => onRemove(index)}
            className="flex-shrink-0 p-2 text-red-600 hover:bg-red-50 rounded-md transition-colors"
            title="Remove state"
          >
            <TrashIcon className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}
