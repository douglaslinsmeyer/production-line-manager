import { useState, useEffect } from 'react';
import { PlusIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import Card from '@/components/common/Card';
import Button from '@/components/common/Button';
import Loading from '@/components/common/Loading';
import Modal from '@/components/common/Modal';
import LabelBadge from '@/components/labels/LabelBadge';
import { useLabels } from '@/hooks/useLines';
import { labelsApi } from '@/api/lines';
import { useQueryClient } from '@tanstack/react-query';
import { lineKeys } from '@/hooks/useLines';
import type { Label, CreateLabelRequest, UpdateLabelRequest } from '@/api/types';

export default function LabelsPage() {
  const { data: labels = [], isLoading } = useLabels();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingLabel, setEditingLabel] = useState<Label | null>(null);
  const [deletingLabel, setDeletingLabel] = useState<Label | null>(null);
  const queryClient = useQueryClient();

  if (isLoading) {
    return <Loading message="Loading labels..." />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Label Management</h1>
          <p className="mt-1 text-gray-500">Create and manage labels for organizing production lines</p>
        </div>
        <Button variant="primary" onClick={() => setShowCreateModal(true)}>
          <PlusIcon className="h-5 w-5 mr-2" />
          Create Label
        </Button>
      </div>

      {/* Labels list */}
      <Card>
        {labels.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 text-lg">No labels created yet</p>
            <Button variant="primary" onClick={() => setShowCreateModal(true)} className="mt-4">
              Create Your First Label
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Label
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {labels.map((label) => (
                  <tr key={label.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <LabelBadge label={label} />
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">
                      {label.description || '-'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(label.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <button
                        onClick={() => setEditingLabel(label)}
                        className="text-blue-600 hover:text-blue-900 mr-4"
                      >
                        <PencilIcon className="h-5 w-5 inline" />
                      </button>
                      <button
                        onClick={() => setDeletingLabel(label)}
                        className="text-red-600 hover:text-red-900"
                      >
                        <TrashIcon className="h-5 w-5 inline" />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Create Modal */}
      <LabelFormModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSuccess={() => {
          setShowCreateModal(false);
          queryClient.invalidateQueries({ queryKey: lineKeys.labels() });
        }}
        title="Create Label"
        mode="create"
      />

      {/* Edit Modal */}
      {editingLabel && (
        <LabelFormModal
          isOpen={!!editingLabel}
          onClose={() => setEditingLabel(null)}
          onSuccess={() => {
            setEditingLabel(null);
            queryClient.invalidateQueries({ queryKey: lineKeys.labels() });
          }}
          title="Edit Label"
          mode="edit"
          initialData={editingLabel}
        />
      )}

      {/* Delete Confirmation Modal */}
      {deletingLabel && (
        <DeleteLabelModal
          isOpen={!!deletingLabel}
          onClose={() => setDeletingLabel(null)}
          label={deletingLabel}
          onSuccess={() => {
            setDeletingLabel(null);
            queryClient.invalidateQueries({ queryKey: lineKeys.labels() });
          }}
        />
      )}
    </div>
  );
}

// Label Form Modal Component
interface LabelFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  title: string;
  mode: 'create' | 'edit';
  initialData?: Label;
}

function LabelFormModal({ isOpen, onClose, onSuccess, title, mode, initialData }: LabelFormModalProps) {
  // Default to blue if no color is set
  const defaultColor = '#3b82f6';

  const [name, setName] = useState('');
  const [color, setColor] = useState(defaultColor);
  const [description, setDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Reset form when modal opens/closes or initialData changes
  useEffect(() => {
    if (isOpen) {
      const initialColor = initialData?.color?.startsWith('#') ? initialData.color : defaultColor;
      setName(initialData?.name || '');
      setColor(initialColor);
      setDescription(initialData?.description || '');
    }
  }, [isOpen, initialData]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Label name is required');
      return;
    }

    setIsSubmitting(true);
    try {
      if (mode === 'create') {
        const request: CreateLabelRequest = {
          name: name.trim(),
          color: color,
          description: description || undefined,
        };
        await labelsApi.createLabel(request);
        toast.success('Label created successfully');
      } else if (initialData) {
        const request: UpdateLabelRequest = {
          name: name.trim(),
          color: color,
          description: description || undefined,
        };
        await labelsApi.updateLabel(initialData.id, request);
        toast.success('Label updated successfully');
      }
      onSuccess();
    } catch (error) {
      toast.error(mode === 'create' ? 'Failed to create label' : 'Failed to update label');
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={title}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
            Name <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="e.g., production"
            required
          />
        </div>

        <div>
          <label htmlFor="color" className="block text-sm font-medium text-gray-700 mb-2">
            Color
          </label>
          <div className="flex items-center gap-3">
            <input
              type="color"
              id="color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              className="h-10 w-20 rounded border border-gray-300 cursor-pointer"
            />
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-600 font-mono">{color}</span>
            </div>
          </div>
          <p className="mt-1 text-xs text-gray-500">Pick a color for this label badge</p>
          {name && (
            <div className="mt-3 pt-3 border-t border-gray-200">
              <p className="text-xs font-medium text-gray-700 mb-2">Preview:</p>
              <LabelBadge label={{ id: 'preview', name, color, description, created_at: new Date().toISOString(), updated_at: new Date().toISOString() }} />
            </div>
          )}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
            Description (optional)
          </label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Optional description..."
          />
        </div>

        <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
          <Button type="button" variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" loading={isSubmitting}>
            {mode === 'create' ? 'Create Label' : 'Update Label'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

// Delete Confirmation Modal
interface DeleteLabelModalProps {
  isOpen: boolean;
  onClose: () => void;
  label: Label;
  onSuccess: () => void;
}

function DeleteLabelModal({ isOpen, onClose, label, onSuccess }: DeleteLabelModalProps) {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      await labelsApi.deleteLabel(label.id);
      toast.success(`Label "${label.name}" deleted`);
      onSuccess();
    } catch (error) {
      toast.error('Failed to delete label');
      console.error(error);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Delete Label">
      <div className="space-y-4">
        <p className="text-gray-700">
          Are you sure you want to delete the label <strong>"{label.name}"</strong>?
        </p>
        <p className="text-sm text-gray-500">
          This will remove the label from all production lines. This action cannot be undone.
        </p>

        <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
          <Button variant="secondary" onClick={onClose} disabled={isDeleting}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDelete} loading={isDeleting}>
            Delete Label
          </Button>
        </div>
      </div>
    </Modal>
  );
}
