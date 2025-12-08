import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Button from '@/components/common/Button';
import LabelInput from '@/components/labels/LabelInput';
import { useLabels } from '@/hooks/useLines';
import { useSchedules } from '@/hooks/useSchedules';
import type { Label } from '@/api/types';

// Validation schemas
const createLineSchema = z.object({
  code: z
    .string()
    .min(1, 'Code is required')
    .max(50, 'Code must be 50 characters or less'),
  name: z
    .string()
    .min(1, 'Name is required')
    .max(255, 'Name must be 255 characters or less'),
  description: z.string().optional(),
  labels: z.array(z.custom<Label>()).optional(),
  schedule_id: z.string().optional().nullable(),
});

const editLineSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(255, 'Name must be 255 characters or less'),
  description: z.string().optional(),
  labels: z.array(z.custom<Label>()).optional(),
  schedule_id: z.string().optional().nullable(),
});

type CreateLineFormData = z.infer<typeof createLineSchema>;
type EditLineFormData = z.infer<typeof editLineSchema>;
type LineFormData = CreateLineFormData | EditLineFormData;

interface LineFormProps {
  initialData?: Partial<LineFormData>;
  onSubmit: (data: any) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
  mode?: 'create' | 'edit';
}

export default function LineForm({
  initialData,
  onSubmit,
  onCancel,
  isSubmitting = false,
  mode = 'create',
}: LineFormProps) {
  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<any>({
    resolver: zodResolver(mode === 'create' ? createLineSchema : editLineSchema) as any,
    defaultValues: initialData,
  });

  const { data: availableLabels = [] } = useLabels();
  const { data: availableSchedules = [] } = useSchedules();

  const handleFormSubmit = async (data: any) => {
    try {
      await onSubmit(data);
    } catch (error) {
      // Error handling is done in the parent component
      console.error('Form submission error:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
      {/* Code field (only for create mode) */}
      {mode === 'create' && (
        <div>
          <label htmlFor="code" className="block text-sm font-medium text-gray-700 mb-2">
            Line Code <span className="text-red-500">*</span>
          </label>
          <input
            {...register('code')}
            type="text"
            id="code"
            className={`
              w-full px-4 py-2 rounded-lg border
              ${errors.code ? 'border-red-300' : 'border-gray-300'}
              focus:outline-none focus:ring-2 focus:ring-blue-500
            `}
            placeholder="e.g., LINE-001"
          />
          {errors.code && (
            <p className="mt-1 text-sm text-red-600">{String(errors.code.message)}</p>
          )}
        </div>
      )}

      {/* Name field */}
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
          Line Name <span className="text-red-500">*</span>
        </label>
        <input
          {...register('name')}
          type="text"
          id="name"
          className={`
            w-full px-4 py-2 rounded-lg border
            ${errors.name ? 'border-red-300' : 'border-gray-300'}
            focus:outline-none focus:ring-2 focus:ring-blue-500
          `}
          placeholder="e.g., Assembly Line A"
        />
        {errors.name && (
          <p className="mt-1 text-sm text-red-600">{String(errors.name.message)}</p>
        )}
      </div>

      {/* Description field */}
      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
          Description
        </label>
        <textarea
          {...register('description')}
          id="description"
          rows={4}
          className="
            w-full px-4 py-2 rounded-lg border border-gray-300
            focus:outline-none focus:ring-2 focus:ring-blue-500
          "
          placeholder="Optional description of the production line..."
        />
        {errors.description && (
          <p className="mt-1 text-sm text-red-600">{String(errors.description.message)}</p>
        )}
      </div>

      {/* Schedule field */}
      <div>
        <label htmlFor="schedule_id" className="block text-sm font-medium text-gray-700 mb-2">
          Schedule
        </label>
        <select
          {...register('schedule_id')}
          id="schedule_id"
          className="
            w-full px-4 py-2 rounded-lg border border-gray-300
            focus:outline-none focus:ring-2 focus:ring-blue-500
          "
        >
          <option value="">No schedule assigned</option>
          {availableSchedules.map((schedule) => (
            <option key={schedule.id} value={schedule.id}>
              {schedule.name}
            </option>
          ))}
        </select>
        <p className="mt-1 text-sm text-gray-500">
          Assign a schedule to track planned uptime vs actual
        </p>
      </div>

      {/* Labels field */}
      <div>
        <label htmlFor="labels" className="block text-sm font-medium text-gray-700 mb-2">
          Labels
        </label>
        <Controller
          name="labels"
          control={control}
          render={({ field }) => (
            <LabelInput
              value={field.value || []}
              onChange={field.onChange}
              availableLabels={availableLabels}
              placeholder="Add labels to categorize this line..."
            />
          )}
        />
        <p className="mt-1 text-sm text-gray-500">
          Add labels like "production", "assembly", "testing" to organize lines
        </p>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={isSubmitting}>
          Cancel
        </Button>
        <Button type="submit" variant="primary" loading={isSubmitting}>
          {mode === 'create' ? 'Create Line' : 'Update Line'}
        </Button>
      </div>
    </form>
  );
}
