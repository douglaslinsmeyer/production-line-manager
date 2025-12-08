import type { Status } from '@/api/types';

interface StatusCircleProps {
  status: Status;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const STATUS_CIRCLE_COLORS: Record<Status, string> = {
  on: 'bg-green-500',
  off: 'bg-red-500',
  maintenance: 'bg-yellow-500',
  error: 'bg-red-500',
};

const SIZE_CLASSES = {
  sm: 'w-3 h-3',
  md: 'w-4 h-4',
  lg: 'w-6 h-6',
};

export default function StatusCircle({ status, size = 'md', className = '' }: StatusCircleProps) {
  return (
    <div className={`relative ${className}`}>
      <div
        className={`
          rounded-full ${STATUS_CIRCLE_COLORS[status]} ${SIZE_CLASSES[size]}
          ${status === 'on' ? 'animate-pulse' : ''}
        `}
        aria-label={`Status: ${status}`}
      />
    </div>
  );
}
