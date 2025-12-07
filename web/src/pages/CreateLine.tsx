import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useCreateLine } from '@/hooks/useLines';
import type { CreateLineRequest } from '@/api/types';
import Card from '@/components/common/Card';
import Button from '@/components/common/Button';
import LineForm from '@/components/lines/LineForm';

export default function CreateLine() {
  const navigate = useNavigate();
  const createLine = useCreateLine();
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const handleSubmit = async (data: CreateLineRequest) => {
    setError(null);
    setSuccessMessage(null);

    try {
      await createLine.mutateAsync(data);
      setSuccessMessage(`Production line "${data.name}" created successfully!`);
      // Navigate to dashboard after a brief delay to show success message
      setTimeout(() => {
        navigate('/', { replace: true });
      }, 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create production line');
    }
  };

  const handleCancel = () => {
    navigate('/', { replace: true });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-4">
        <Link to="/">
          <Button variant="secondary" size="sm">
            <ArrowLeftIcon className="h-4 w-4 mr-2" />
            Back
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Create Production Line</h1>
          <p className="mt-1 text-gray-600">Add a new production line to the system</p>
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
          mode="create"
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          isSubmitting={createLine.isPending}
        />
      </Card>
    </div>
  );
}
