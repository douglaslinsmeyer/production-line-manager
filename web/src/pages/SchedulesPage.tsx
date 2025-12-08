import { Link } from 'react-router-dom';
import { PlusIcon } from '@heroicons/react/24/outline';
import Button from '@/components/common/Button';
import Loading from '@/components/common/Loading';
import { ScheduleCard } from '@/components/schedules';
import { useSchedules } from '@/hooks/useSchedules';

export default function SchedulesPage() {
  const { data: schedules = [], isLoading } = useSchedules();

  if (isLoading) {
    return <Loading message="Loading schedules..." />;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Schedules</h1>
          <p className="mt-1 text-gray-500">
            Manage production schedules with shift times, breaks, and exceptions
          </p>
        </div>
        <Link to="/schedules/new">
          <Button variant="primary">
            <PlusIcon className="h-5 w-5 mr-2" />
            Create Schedule
          </Button>
        </Link>
      </div>

      {/* Schedules list */}
      {schedules.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
          <p className="text-gray-500 text-lg">No schedules created yet</p>
          <Link to="/schedules/new">
            <Button variant="primary" className="mt-4">
              Create Your First Schedule
            </Button>
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {schedules.map((schedule) => (
            <ScheduleCard key={schedule.id} schedule={schedule} />
          ))}
        </div>
      )}
    </div>
  );
}
