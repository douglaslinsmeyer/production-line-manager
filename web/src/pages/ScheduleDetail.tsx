import { useParams, Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { PencilIcon, TrashIcon, ArrowLeftIcon } from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';
import Button from '@/components/common/Button';
import Loading from '@/components/common/Loading';
import Card from '@/components/common/Card';
import Modal from '@/components/common/Modal';
import {
  ScheduleViewer,
  HolidayList,
  ExceptionList,
  LineExceptionList,
} from '@/components/schedules';
import {
  useSchedule,
  useScheduleLines,
  useHolidays,
  useExceptions,
  useLineExceptions,
  useCreateHoliday,
  useUpdateHoliday,
  useDeleteHoliday,
  useCreateException,
  useDeleteException,
  useCreateLineException,
  useDeleteLineException,
  useDeleteSchedule,
} from '@/hooks/useSchedules';

export default function ScheduleDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [activeTab, setActiveTab] = useState<'schedule' | 'holidays' | 'exceptions' | 'line-exceptions'>('schedule');

  const { data: schedule, isLoading: scheduleLoading } = useSchedule(id!);
  const { data: scheduleLines = [], isLoading: linesLoading } = useScheduleLines(id!);
  const { data: holidays = [], isLoading: holidaysLoading } = useHolidays(id!);
  const { data: exceptions = [], isLoading: exceptionsLoading } = useExceptions(id!);
  const { data: lineExceptions = [], isLoading: lineExceptionsLoading } = useLineExceptions(id!);

  const createHoliday = useCreateHoliday(id!);
  const updateHoliday = useUpdateHoliday(id!, '');
  const deleteHoliday = useDeleteHoliday(id!);
  const createException = useCreateException(id!);
  const deleteException = useDeleteException(id!);
  const createLineException = useCreateLineException(id!);
  const deleteLineException = useDeleteLineException(id!);
  const deleteSchedule = useDeleteSchedule();

  if (scheduleLoading || linesLoading) {
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

  const handleDelete = async () => {
    try {
      await deleteSchedule.mutateAsync(id!);
      toast.success('Schedule deleted successfully');
      navigate('/schedules');
    } catch (error) {
      toast.error('Failed to delete schedule');
      console.error(error);
    }
  };

  const tabs = [
    { id: 'schedule', name: 'Weekly Schedule' },
    { id: 'holidays', name: `Holidays (${holidays.length})` },
    { id: 'exceptions', name: `Exceptions (${exceptions.length})` },
    { id: 'line-exceptions', name: `Line Exceptions (${lineExceptions.length})` },
  ] as const;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <Link
            to="/schedules"
            className="text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1 mb-2"
          >
            <ArrowLeftIcon className="h-4 w-4" />
            Back to Schedules
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">{schedule.name}</h1>
          {schedule.description && (
            <p className="mt-1 text-gray-500">{schedule.description}</p>
          )}
          <p className="mt-2 text-sm text-gray-400">
            Timezone: {schedule.timezone} | {scheduleLines.length}{' '}
            {scheduleLines.length === 1 ? 'line' : 'lines'} using this schedule
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Link to={`/schedules/${id}/edit`}>
            <Button variant="secondary">
              <PencilIcon className="h-4 w-4 mr-2" />
              Edit
            </Button>
          </Link>
          <Button variant="danger" onClick={() => setShowDeleteModal(true)}>
            <TrashIcon className="h-4 w-4 mr-2" />
            Delete
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.name}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <Card>
        {activeTab === 'schedule' && <ScheduleViewer schedule={schedule} />}

        {activeTab === 'holidays' && (
          <HolidayList
            holidays={holidays}
            scheduleId={id!}
            onAdd={(data) => createHoliday.mutate(data)}
            onUpdate={(_holidayId, data) => {
              // TODO: Refactor to support dynamic holidayId in the mutation
              updateHoliday.mutate(data);
            }}
            onDelete={(holidayId) => deleteHoliday.mutate(holidayId)}
            isLoading={holidaysLoading || createHoliday.isPending || deleteHoliday.isPending}
          />
        )}

        {activeTab === 'exceptions' && (
          <ExceptionList
            exceptions={exceptions}
            scheduleId={id!}
            onAdd={(data) => createException.mutate(data)}
            onDelete={(exceptionId) => deleteException.mutate(exceptionId)}
            isLoading={exceptionsLoading || createException.isPending || deleteException.isPending}
          />
        )}

        {activeTab === 'line-exceptions' && (
          <LineExceptionList
            exceptions={lineExceptions}
            scheduleId={id!}
            availableLines={scheduleLines}
            onAdd={(data) => createLineException.mutate(data)}
            onDelete={(exceptionId) => deleteLineException.mutate(exceptionId)}
            isLoading={
              lineExceptionsLoading || createLineException.isPending || deleteLineException.isPending
            }
          />
        )}
      </Card>

      {/* Lines using this schedule */}
      {scheduleLines.length > 0 && (
        <Card title="Production Lines Using This Schedule">
          <div className="space-y-2">
            {scheduleLines.map((line) => (
              <Link
                key={line.id}
                to={`/lines/${line.id}`}
                className="block p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <span className="font-medium text-gray-900">{line.name}</span>
                <span className="text-gray-500 ml-2">({line.code})</span>
              </Link>
            ))}
          </div>
        </Card>
      )}

      {/* Delete Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Schedule"
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete <strong>"{schedule.name}"</strong>?
          </p>
          {scheduleLines.length > 0 && (
            <p className="text-sm text-amber-600 bg-amber-50 p-3 rounded-lg">
              Warning: {scheduleLines.length} production{' '}
              {scheduleLines.length === 1 ? 'line is' : 'lines are'} currently using this schedule.
              They will be unassigned.
            </p>
          )}
          <p className="text-sm text-gray-500">This action cannot be undone.</p>

          <div className="flex items-center justify-end space-x-3 pt-4 border-t border-gray-200">
            <Button
              variant="secondary"
              onClick={() => setShowDeleteModal(false)}
              disabled={deleteSchedule.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              loading={deleteSchedule.isPending}
            >
              Delete Schedule
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
