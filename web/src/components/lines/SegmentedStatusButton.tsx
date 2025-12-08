import { PlayIcon, StopIcon, WrenchScrewdriverIcon } from '@heroicons/react/24/outline';
import { useSetStatus } from '@/hooks/useLines';
import type { Status } from '@/api/types';
import toast from 'react-hot-toast';

interface SegmentedStatusButtonProps {
  lineId: string;
  currentStatus: Status;
  disabled?: boolean;
}

interface StatusSegment {
  status: Status;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
}

interface StatusSegmentConfig extends StatusSegment {
  activeColor: string;
}

const SEGMENTS: StatusSegmentConfig[] = [
  { status: 'on', label: 'Start', icon: PlayIcon, activeColor: 'bg-green-600 text-white' },
  { status: 'off', label: 'Stop', icon: StopIcon, activeColor: 'bg-red-600 text-white' },
  { status: 'maintenance', label: 'Maintenance', icon: WrenchScrewdriverIcon, activeColor: 'bg-blue-600 text-white' },
];

export default function SegmentedStatusButton({
  lineId,
  currentStatus,
  disabled = false,
}: SegmentedStatusButtonProps) {
  const setStatus = useSetStatus(lineId);

  const handleStatusChange = async (newStatus: Status) => {
    // Don't do anything if already in this status
    if (newStatus === currentStatus) return;

    try {
      await setStatus.mutateAsync(newStatus);
      toast.success(`Status changed to ${newStatus}`);
    } catch (err) {
      toast.error('Failed to update status');
      console.error('Status update error:', err);
    }
  };

  // Disable all buttons if status is 'error' (system-set only)
  const isErrorState = currentStatus === 'error';
  const isDisabled = disabled || isErrorState || setStatus.isPending;

  return (
    <div className="inline-flex rounded-lg border border-gray-300 overflow-hidden">
      {SEGMENTS.map((segment, index) => {
        const isActive = currentStatus === segment.status;
        const Icon = segment.icon;

        return (
          <button
            key={segment.status}
            onClick={() => handleStatusChange(segment.status)}
            disabled={isDisabled}
            className={`
              px-4 py-2 text-sm font-medium transition-colors
              flex items-center gap-1.5
              ${
                isActive
                  ? segment.activeColor
                  : 'bg-white text-gray-700 hover:bg-gray-50'
              }
              ${index > 0 ? 'border-l border-gray-300' : ''}
              disabled:opacity-50 disabled:cursor-not-allowed
            `}
            aria-label={`Set status to ${segment.label}`}
            aria-pressed={isActive}
          >
            <Icon className="h-4 w-4" />
            <span>{segment.label}</span>
          </button>
        );
      })}
    </div>
  );
}
