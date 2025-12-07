import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
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
  const deleteLine = useDeleteLine();
  const navigate = useNavigate();

  const handleDelete = async () => {
    try {
      await deleteLine.mutateAsync(line.id);
      toast.success(`Production line "${line.name}" deleted successfully`);
      onClose();
      if (redirectAfterDelete) {
        navigate('/');
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to delete line');
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
