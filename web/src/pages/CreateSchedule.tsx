import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import { ScheduleForm } from '@/components/schedules';
import { useCreateSchedule } from '@/hooks/useSchedules';
import type { CreateScheduleRequest } from '@/api/types';

export default function CreateSchedule() {
  const navigate = useNavigate();
  const createSchedule = useCreateSchedule();

  const handleSubmit = async (data: CreateScheduleRequest) => {
    try {
      const schedule = await createSchedule.mutateAsync(data);
      toast.success('Schedule created successfully');
      navigate(`/schedules/${schedule.id}`);
    } catch (error) {
      toast.error('Failed to create schedule');
      console.error(error);
    }
  };

  const handleCancel = () => {
    navigate('/schedules');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Create Schedule</h1>
        <p className="mt-1 text-gray-500">
          Define a new production schedule with weekly shift patterns and breaks
        </p>
      </div>

      {/* Form */}
      <ScheduleForm
        onSubmit={handleSubmit}
        onCancel={handleCancel}
        isLoading={createSchedule.isPending}
      />
    </div>
  );
}
