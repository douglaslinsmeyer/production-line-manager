import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Button from '@/components/common/Button';

// Validation schema
const lineSchema = z.object({
  code: z
    .string()
    .min(1, 'Code is required')
    .max(50, 'Code must be 50 characters or less'),
  name: z
    .string()
    .min(1, 'Name is required')
    .max(255, 'Name must be 255 characters or less'),
  description: z.string().optional(),
});

type LineFormData = z.infer<typeof lineSchema>;

interface LineFormProps {
  initialData?: Partial<LineFormData>;
  onSubmit: (data: LineFormData) => Promise<void>;
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
    formState: { errors },
  } = useForm<LineFormData>({
    resolver: zodResolver(lineSchema),
    defaultValues: initialData,
  });

  const handleFormSubmit = async (data: LineFormData) => {
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
            <p className="mt-1 text-sm text-red-600">{errors.code.message}</p>
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
          <p className="mt-1 text-sm text-red-600">{errors.name.message}</p>
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
          <p className="mt-1 text-sm text-red-600">{errors.description.message}</p>
        )}
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
