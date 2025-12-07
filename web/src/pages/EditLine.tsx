import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useLine, useUpdateLine } from '@/hooks/useLines';
import type { UpdateLineRequest } from '@/api/types';
import Loading from '@/components/common/Loading';
import Card from '@/components/common/Card';
import Button from '@/components/common/Button';
import LineForm from '@/components/lines/LineForm';

export default function EditLine() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: line, isLoading, error: fetchError } = useLine(id!);
  const updateLine = useUpdateLine(id!);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const handleSubmit = async (data: UpdateLineRequest) => {
    setError(null);
    setSuccessMessage(null);

    try {
      await updateLine.mutateAsync(data);
      setSuccessMessage('Production line updated successfully!');
      // Navigate back to detail page after a brief delay
      setTimeout(() => {
        navigate(`/lines/${id}`, { replace: true });
      }, 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update production line');
    }
  };

  const handleCancel = () => {
    navigate(`/lines/${id}`, { replace: true });
  };

  if (isLoading) {
    return <Loading message="Loading production line..." />;
  }

  if (fetchError) {
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
            {fetchError instanceof Error ? fetchError.message : 'An unexpected error occurred'}
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
      <div className="flex items-center space-x-4">
        <Link to={`/lines/${id}`}>
          <Button variant="secondary" size="sm">
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Edit Production Line</h1>
          <p className="mt-1 text-gray-600">{line.name} (Code: {line.code})</p>
        </div>
      </div>

      {/* Success message */}
      {successMessage && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <p className="text-green-800">{successMessage}</p>
        </div>
      )}

      {/* Error message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <p className="text-red-600">{error}</p>
        </div>
      )}

      {/* Form card */}
      <Card>
        <LineForm
          mode="edit"
          initialData={{
            name: line.name,
            description: line.description,
          }}
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          isSubmitting={updateLine.isPending}
        />
      </Card>
    </div>
  );
}
