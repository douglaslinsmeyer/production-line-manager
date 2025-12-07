import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { PencilIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useLine } from '@/hooks/useLines';
import Loading from '@/components/common/Loading';
import Card from '@/components/common/Card';
import StatusBadge from '@/components/common/StatusBadge';
import Button from '@/components/common/Button';
import StatusForm from '@/components/lines/StatusForm';
import DeleteConfirmModal from '@/components/lines/DeleteConfirmModal';
import { formatDate, formatRelativeTime } from '@/utils/formatters';

export default function LineDetail() {
  const { id } = useParams<{ id: string }>();
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  const { data: line, isLoading, error } = useLine(id!);

  if (isLoading) {
    return <Loading message="Loading production line..." />;
  }

  if (error) {
    return (
      <div className="space-y-4">
        <Link to="/">
          <Button variant="secondary" size="sm">
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back to Dashboard
          </Button>
        </Link>
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h3 className="text-red-800 font-semibold mb-2">Error Loading Line</h3>
          <p className="text-red-600">
            {error instanceof Error ? error.message : 'An unexpected error occurred'}
          </p>
        </div>
      </div>
    );
  }

  if (!line) {
    return (
      <div className="space-y-4">
        <Link to="/">
          <Button variant="secondary" size="sm">
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back to Dashboard
          </Button>
        </Link>
        <div className="bg-gray-50 border border-gray-200 rounded-lg p-12 text-center">
          <p className="text-gray-600 text-lg">Production line not found</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Link to="/">
            <Button variant="secondary" size="sm">
              <ArrowLeftIcon className="h-4 w-4 mr-2" />
              Back
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{line.name}</h1>
            <p className="mt-1 text-gray-600">Code: {line.code}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          <Link to={`/lines/${line.id}/edit`}>
            <Button variant="secondary">
              <PencilIcon className="h-5 w-5 mr-2" />
              Edit
            </Button>
          </Link>
          <Button variant="danger" onClick={() => setShowDeleteModal(true)}>
            Delete
          </Button>
        </div>
      </div>

      {/* Line information */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Details card */}
        <Card title="Line Details">
          <dl className="space-y-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">Status</dt>
              <dd className="mt-1">
                <StatusBadge status={line.status} />
              </dd>
            </div>

            {line.description && (
              <div>
                <dt className="text-sm font-medium text-gray-500">Description</dt>
                <dd className="mt-1 text-gray-900">{line.description}</dd>
              </div>
            )}

            <div>
              <dt className="text-sm font-medium text-gray-500">Created</dt>
              <dd className="mt-1 text-gray-900">
                {formatDate(line.created_at)}
                <span className="text-gray-500 text-sm ml-2">
                  ({formatRelativeTime(line.created_at)})
                </span>
              </dd>
            </div>

            <div>
              <dt className="text-sm font-medium text-gray-500">Last Updated</dt>
              <dd className="mt-1 text-gray-900">
                {formatDate(line.updated_at)}
                <span className="text-gray-500 text-sm ml-2">
                  ({formatRelativeTime(line.updated_at)})
                </span>
              </dd>
            </div>
          </dl>
        </Card>

        {/* Status control card */}
        <Card title="Status Control">
          <StatusForm lineId={line.id} currentStatus={line.status} />
        </Card>
      </div>

      {/* Analytics placeholder - will be implemented in Phase 6 */}
      <Card title="Analytics">
        <div className="bg-gray-50 rounded-lg p-8 text-center">
          <p className="text-gray-600">
            Production analytics and metrics will be displayed here
          </p>
          <p className="text-gray-500 text-sm mt-2">Coming in Phase 6</p>
        </div>
      </Card>

      {/* Delete confirmation modal */}
      <DeleteConfirmModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        line={line}
        redirectAfterDelete={true}
      />
    </div>
  );
}
