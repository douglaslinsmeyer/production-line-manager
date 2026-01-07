import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { PlusIcon } from '@heroicons/react/24/outline';
import type { SignalProfile, ProfileState, ButtonBehavior } from '../../api/types';
import StateEditor from './StateEditor';
import ButtonCycleEditor from './ButtonCycleEditor';
import Button from '../common/Button';

// Validation schemas
const stateOutputsSchema = z.object({
  redLight: z.enum(['off', 'on', 'shortBlink', 'longBlink']),
  yellowLight: z.enum(['off', 'on', 'shortBlink', 'longBlink']),
  greenLight: z.enum(['off', 'on', 'shortBlink', 'longBlink']),
  primaryBuzzer: z.enum(['off', 'on', 'chirp']),
  towerBuzzer: z.enum(['off', 'on', 'chirp']),
});

const profileStateSchema = z.object({
  name: z.string().min(1, 'State name required').max(50, 'State name too long'),
  outputs: stateOutputsSchema,
});

const buttonBehaviorSchema = z.object({
  shortPressCycle: z.array(z.string()).min(1, 'Short press cycle must have at least one state'),
  longPressCycle: z.array(z.string()).min(1, 'Long press cycle must have at least one state'),
});

const profileFormSchema = z.object({
  name: z.string().min(1, 'Name required').max(100, 'Name too long'),
  description: z.string().optional(),
  states: z.array(profileStateSchema).min(1, 'At least one state required'),
  buttonBehavior: buttonBehaviorSchema,
  defaultState: z.string().min(1, 'Default state required'),
}).refine(
  (data) => data.states.some(s => s.name === data.defaultState),
  { message: 'Default state must exist in states list', path: ['defaultState'] }
).refine(
  (data) => {
    const stateNames = data.states.map(s => s.name);
    return data.buttonBehavior.shortPressCycle.every(s => stateNames.includes(s));
  },
  { message: 'All short press states must exist in states', path: ['buttonBehavior', 'shortPressCycle'] }
).refine(
  (data) => {
    const stateNames = data.states.map(s => s.name);
    return data.buttonBehavior.longPressCycle.every(s => stateNames.includes(s));
  },
  { message: 'All long press states must exist in states', path: ['buttonBehavior', 'longPressCycle'] }
);

type ProfileFormData = z.infer<typeof profileFormSchema>;

interface SignalProfileFormProps {
  initialData?: Partial<SignalProfile>;
  onSubmit: (data: ProfileFormData) => void;
  isSubmitting?: boolean;
  submitLabel?: string;
}

export default function SignalProfileForm({
  initialData,
  onSubmit,
  isSubmitting = false,
  submitLabel = 'Save Profile',
}: SignalProfileFormProps) {
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ProfileFormData>({
    resolver: zodResolver(profileFormSchema),
    defaultValues: initialData || {
      name: '',
      description: '',
      states: [
        {
          name: 'On',
          outputs: { redLight: 'off', yellowLight: 'off', greenLight: 'on', primaryBuzzer: 'off', towerBuzzer: 'off' },
        },
        {
          name: 'Off',
          outputs: { redLight: 'on', yellowLight: 'off', greenLight: 'off', primaryBuzzer: 'off', towerBuzzer: 'off' },
        },
      ],
      buttonBehavior: {
        shortPressCycle: ['On', 'Off'],
        longPressCycle: ['Maintenance'],
      },
      defaultState: 'Off',
    },
  });

  const states = watch('states');
  const buttonBehavior = watch('buttonBehavior');
  const stateNames = states.map((s) => s.name);

  const addState = () => {
    const newStateName = `State ${states.length + 1}`;
    setValue('states', [
      ...states,
      {
        name: newStateName,
        outputs: { redLight: 'off', yellowLight: 'off', greenLight: 'off', primaryBuzzer: 'off', towerBuzzer: 'off' },
      },
    ]);
  };

  const updateState = (index: number, state: ProfileState) => {
    const updatedStates = [...states];
    updatedStates[index] = state;
    setValue('states', updatedStates);
  };

  const removeState = (index: number) => {
    const updatedStates = states.filter((_, i) => i !== index);
    setValue('states', updatedStates);
  };

  const updateButtonBehavior = (newBehavior: ButtonBehavior) => {
    setValue('buttonBehavior', newBehavior);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
      {/* Basic Information */}
      <div className="bg-white p-6 rounded-lg shadow space-y-4">
        <h2 className="text-lg font-semibold text-gray-900">Basic Information</h2>

        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
            Profile Name *
          </label>
          <input
            {...register('name')}
            type="text"
            id="name"
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
            placeholder="e.g., Assembly Line Profile"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea
            {...register('description')}
            id="description"
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
            placeholder="Optional description of this profile"
          />
        </div>
      </div>

      {/* States Configuration */}
      <div className="bg-white p-6 rounded-lg shadow space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">States</h2>
          <button
            type="button"
            onClick={addState}
            className="flex items-center gap-1 px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors text-sm"
          >
            <PlusIcon className="w-4 h-4" />
            Add State
          </button>
        </div>

        <div className="space-y-3">
          {states.map((state, index) => (
            <StateEditor
              key={index}
              state={state}
              index={index}
              onChange={updateState}
              onRemove={removeState}
              canRemove={states.length > 1}
            />
          ))}
        </div>

        {errors.states && (
          <p className="text-sm text-red-600">{errors.states.message}</p>
        )}
      </div>

      {/* Button Behavior */}
      <div className="bg-white p-6 rounded-lg shadow space-y-4">
        <h2 className="text-lg font-semibold text-gray-900">Button Behavior</h2>

        <ButtonCycleEditor
          buttonBehavior={buttonBehavior}
          availableStates={stateNames}
          onChange={updateButtonBehavior}
        />

        {errors.buttonBehavior && (
          <p className="text-sm text-red-600">{errors.buttonBehavior.message}</p>
        )}
      </div>

      {/* Default State */}
      <div className="bg-white p-6 rounded-lg shadow space-y-4">
        <h2 className="text-lg font-semibold text-gray-900">Default State</h2>

        <div>
          <label htmlFor="defaultState" className="block text-sm font-medium text-gray-700 mb-1">
            Default State *
          </label>
          <select
            {...register('defaultState')}
            id="defaultState"
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">Select default state...</option>
            {stateNames.map((stateName) => (
              <option key={stateName} value={stateName}>
                {stateName}
              </option>
            ))}
          </select>
          {errors.defaultState && (
            <p className="mt-1 text-sm text-red-600">{errors.defaultState.message}</p>
          )}
          <p className="mt-1 text-sm text-gray-500">
            State that devices will use when first assigned or after override reset
          </p>
        </div>
      </div>

      {/* Submit Button */}
      <div className="flex justify-end gap-3">
        <Button
          type="submit"
          variant="primary"
          disabled={isSubmitting}
          className="px-6"
        >
          {isSubmitting ? 'Saving...' : submitLabel}
        </Button>
      </div>
    </form>
  );
}
