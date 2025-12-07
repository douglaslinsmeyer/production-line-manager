import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDeleteLine } from '@/hooks/useLines';
import type { ProductionLine } from '@/api/types';
import Modal from '@/components/common/Modal';
import Button from '@/components/common/Button';

interface DeleteConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  line: ProductionLine;
  redirectAfterDelete?: boolean;
}

export default function DeleteConfirmModal({
  isOpen,
  onClose,
  line,
  redirectAfterDelete = false,
}: DeleteConfirmModalProps) {
  const [error, setError] = useState<string | null>(null);
  const deleteLine = useDeleteLine();
  const navigate = useNavigate();

  const handleDelete = async () => {
    setError(null);

    try {
      await deleteLine.mutateAsync(line.id);
      onClose();
      if (redirectAfterDelete) {
        navigate('/');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete line');
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Delete Production Line">
      <div className="space-y-4">
        <p className="text-gray-700">
          Are you sure you want to delete <strong>{line.name}</strong> (Code: {line.code})?
        </p>
        <p className="text-sm text-gray-600">
          This action cannot be undone. All status history for this line will also be deleted.
        </p>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        <div className="flex items-center justify-end space-x-3 pt-4">
          <Button variant="secondary" onClick={onClose} disabled={deleteLine.isPending}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDelete} loading={deleteLine.isPending}>
            Delete Line
          </Button>
        </div>
      </div>
    </Modal>
  );
}
