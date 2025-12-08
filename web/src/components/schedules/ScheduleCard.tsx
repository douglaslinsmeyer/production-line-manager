import { Link } from 'react-router-dom';
import { CalendarDaysIcon, UserGroupIcon } from '@heroicons/react/24/outline';
import type { ScheduleSummary } from '@/api/types';
import Card from '@/components/common/Card';

interface ScheduleCardProps {
  schedule: ScheduleSummary;
}

export default function ScheduleCard({ schedule }: ScheduleCardProps) {
  return (
    <Card className="hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between">
        <div className="flex-1 min-w-0">
          <Link to={`/schedules/${schedule.id}`}>
            <h3 className="text-lg font-semibold text-gray-900 hover:text-blue-600 truncate">
              {schedule.name}
            </h3>
          </Link>
          {schedule.description && (
            <p className="text-sm text-gray-500 mt-1 truncate">{schedule.description}</p>
          )}
          <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
            <span className="flex items-center gap-1">
              <CalendarDaysIcon className="h-4 w-4" />
              {schedule.timezone}
            </span>
            <span className="flex items-center gap-1">
              <UserGroupIcon className="h-4 w-4" />
              {schedule.line_count} {schedule.line_count === 1 ? 'line' : 'lines'}
            </span>
          </div>
        </div>
        <Link
          to={`/schedules/${schedule.id}`}
          className="flex-shrink-0 px-4 py-2 text-sm font-medium text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded-md transition-colors"
        >
          View Details
        </Link>
      </div>
    </Card>
  );
}
