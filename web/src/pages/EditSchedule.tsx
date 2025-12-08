import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import Loading from '@/components/common/Loading';
import Button from '@/components/common/Button';
import { ScheduleForm } from '@/components/schedules';
import { useSchedule, useUpdateSchedule } from '@/hooks/useSchedules';
import type { CreateScheduleRequest } from '@/api/types';

export default function EditSchedule() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: schedule, isLoading } = useSchedule(id!);
  const updateSchedule = useUpdateSchedule(id!);

  if (isLoading) {
    return <Loading message="Loading schedule..." />;
  }

  if (!schedule) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 text-lg">Schedule not found</p>
        <Link to="/schedules">
          <Button variant="secondary" className="mt-4">
            Back to Schedules
          </Button>
        </Link>
      </div>
    );
  }

  const handleSubmit = async (data: CreateScheduleRequest) => {
    try {
      await updateSchedule.mutateAsync({
        name: data.name,
        description: data.description,
        timezone: data.timezone,
      });
      toast.success('Schedule updated successfully');
      navigate(`/schedules/${id}`);
    } catch (error) {
      toast.error('Failed to update schedule');
      console.error(error);
    }
  };

  const handleCancel = () => {
    navigate(`/schedules/${id}`);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <Link
          to={`/schedules/${id}`}
          className="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1 mb-2"
        >
          <ArrowLeftIcon className="h-4 w-4" />
          Back to Schedule
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">Edit Schedule</h1>
        <p className="mt-1 text-gray-500">Update schedule details and weekly patterns</p>
      </div>

      {/* Form */}
      <ScheduleForm
        schedule={schedule}
        onSubmit={handleSubmit}
        onCancel={handleCancel}
        isLoading={updateSchedule.isPending}
      />
    </div>
  );
}
