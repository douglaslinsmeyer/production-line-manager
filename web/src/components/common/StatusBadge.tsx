import type { Status } from '@/api/types';
import { STATUS_CONFIG } from '@/utils/constants';

interface StatusBadgeProps {
  status: Status;
  className?: string;
}

export default function StatusBadge({ status, className = '' }: StatusBadgeProps) {
  const config = STATUS_CONFIG[status];

  return (
    <span
      className={`
        inline-flex items-center px-3 py-1 rounded-full text-sm font-medium
        ${config.color} ${config.bgColor}
        ${className}
      `}
    >
      {config.label}
    </span>
  );
}
