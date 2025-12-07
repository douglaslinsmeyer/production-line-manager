import type { Status } from '@/api/types';

// Status display configuration
export const STATUS_CONFIG: Record<Status, { label: string; color: string; bgColor: string }> = {
  on: {
    label: 'Running',
    color: 'text-green-800',
    bgColor: 'bg-green-100',
  },
  off: {
    label: 'Stopped',
    color: 'text-gray-800',
    bgColor: 'bg-gray-100',
  },
  maintenance: {
    label: 'Maintenance',
    color: 'text-yellow-800',
    bgColor: 'bg-yellow-100',
  },
  error: {
    label: 'Error',
    color: 'text-red-800',
    bgColor: 'bg-red-100',
  },
};

// Status options for select dropdowns
export const STATUS_OPTIONS: { value: Status; label: string }[] = [
  { value: 'on', label: 'Running' },
  { value: 'off', label: 'Stopped' },
  { value: 'maintenance', label: 'Maintenance' },
  { value: 'error', label: 'Error' },
];
