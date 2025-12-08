import { useNavigate, Link } from 'react-router-dom';
import toast from 'react-hot-toast';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useCreateLine } from '@/hooks/useLines';
import { linesApi } from '@/api/lines';
import Card from '@/components/common/Card';
import Button from '@/components/common/Button';
import LineForm from '@/components/lines/LineForm';

export default function CreateLine() {
  const navigate = useNavigate();
  const createLine = useCreateLine();

  const handleSubmit = async (data: any) => {
    try {
      // Step 1: Create line
      const createdLine = await createLine.mutateAsync({
        code: data.code,
        name: data.name,
        description: data.description,
      });

      // Step 2: Assign labels if any
      if (data.labels && data.labels.length > 0) {
        await linesApi.assignLabels(
          createdLine.id,
          data.labels.map((l: any) => l.id)
        );
      }

      toast.success(`Production line "${data.name}" created successfully!`);
      navigate('/', { replace: true });
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create production line');
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
