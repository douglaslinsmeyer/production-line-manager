import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Menu } from '@headlessui/react';
import { EllipsisVerticalIcon, PencilIcon, TrashIcon } from '@heroicons/react/24/outline';
import type { ProductionLine } from '@/api/types';
import Card from '@/components/common/Card';
import StatusBadge from '@/components/common/StatusBadge';
import Button from '@/components/common/Button';
import StatusForm from './StatusForm';
import DeleteConfirmModal from './DeleteConfirmModal';
import LabelBadge from '@/components/labels/LabelBadge';

interface LineCardProps {
  line: ProductionLine;
}

export default function LineCard({ line }: LineCardProps) {
  const [showStatusForm, setShowStatusForm] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  return (
    <>
      <Card className="hover:shadow-md transition-shadow">
        <div className="space-y-4">
          {/* Header with status badge */}
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <Link to={`/lines/${line.id}`}>
                <h3 className="text-lg font-semibold text-gray-900 hover:text-blue-600">
                  {line.name}
                </h3>
              </Link>
              <p className="text-sm text-gray-500 mt-1">Code: {line.code}</p>
            </div>
            <StatusBadge status={line.status} />
          </div>

          {/* Description */}
          {line.description && (
            <p className="text-gray-600 text-sm line-clamp-2">{line.description}</p>
          )}

          {/* Labels */}
          {line.labels && line.labels.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {line.labels.map((label) => (
                <LabelBadge key={label.id} label={label} size="sm" />
              ))}
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center justify-between pt-4 border-t border-gray-100">
            <div className="flex items-center space-x-2">
              <Button
                size="sm"
                variant="secondary"
                onClick={() => setShowStatusForm(!showStatusForm)}
              >
                Change Status
              </Button>
              <Link to={`/lines/${line.id}/edit`}>
                <Button size="sm" variant="secondary">
                  <PencilIcon className="h-4 w-4 mr-1" />
                  Edit
                </Button>
              </Link>
            </div>

            {/* More actions menu */}
            <Menu as="div" className="relative">
              <Menu.Button className="p-2 text-gray-400 hover:text-gray-600 rounded-lg">
                <EllipsisVerticalIcon className="h-5 w-5" />
              </Menu.Button>
              <Menu.Items className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-10">
                <Menu.Item>
                  {({ active }) => (
                    <button
                      onClick={() => setShowDeleteModal(true)}
                      className={`
                        w-full px-4 py-2 text-left text-sm flex items-center
                        ${active ? 'bg-red-50 text-red-700' : 'text-red-600'}
                      `}
                    >
                      <TrashIcon className="h-4 w-4 mr-2" />
                      Delete Line
                    </button>
                  )}
                </Menu.Item>
              </Menu.Items>
            </Menu>
          </div>

          {/* Status form (inline) */}
          {showStatusForm && (
            <div className="pt-4 border-t border-gray-100">
              <StatusForm
                lineId={line.id}
                currentStatus={line.status}
                onSuccess={() => setShowStatusForm(false)}
                onCancel={() => setShowStatusForm(false)}
              />
            </div>
          )}
        </div>
      </Card>

      {/* Delete confirmation modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        line={line}
      />
    </>
  );
}
