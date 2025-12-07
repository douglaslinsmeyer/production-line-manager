import { useState } from 'react';
import toast from 'react-hot-toast';
import { useSetStatus } from '@/hooks/useLines';
import type { Status } from '@/api/types';
import { STATUS_OPTIONS } from '@/utils/constants';
import Button from '@/components/common/Button';

interface StatusFormProps {
  lineId: string;
  currentStatus: Status;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export default function StatusForm({ lineId, currentStatus, onSuccess, onCancel }: StatusFormProps) {
  const [selectedStatus, setSelectedStatus] = useState<Status>(currentStatus);
  const setStatus = useSetStatus(lineId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (selectedStatus === currentStatus) {
      toast('Status is already set to ' + selectedStatus, { icon: 'ℹ️' });
      onCancel?.();
      return;
    }

    try {
      await setStatus.mutateAsync(selectedStatus);
      toast.success(`Status changed to ${selectedStatus}`);
      onSuccess?.();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update status');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-2">
          Select New Status
        </label>
        <select
          id="status"
          value={selectedStatus}
          onChange={(e) => setSelectedStatus(e.target.value as Status)}
          className="
            w-full px-4 py-2 rounded-lg border border-gray-300
            focus:outline-none focus:ring-2 focus:ring-blue-500
          "
          disabled={setStatus.isPending}
        >
          {STATUS_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      <div className="flex items-center justify-end space-x-2">
        {onCancel && (
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={onCancel}
            disabled={setStatus.isPending}
          >
            Cancel
          </Button>
        )}
        <Button
          type="submit"
          variant="primary"
          size="sm"
          loading={setStatus.isPending}
          disabled={selectedStatus === currentStatus}
        >
          Update Status
        </Button>
      </div>
    </form>
  );
}
