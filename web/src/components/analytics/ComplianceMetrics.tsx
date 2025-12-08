import { ClockIcon, CheckCircleIcon, ExclamationTriangleIcon, ArrowTrendingUpIcon } from '@heroicons/react/24/outline';
import type { AggregateComplianceMetrics } from '@/api/types';

interface ComplianceMetricsProps {
  metrics: AggregateComplianceMetrics;
}

export default function ComplianceMetrics({ metrics }: ComplianceMetricsProps) {
  const formatHours = (hours: number): string => {
    if (hours < 1) {
      return `${Math.round(hours * 60)}m`;
    }
    return `${hours.toFixed(1)}h`;
  };

  const getComplianceColor = (percentage: number): string => {
    if (percentage >= 90) return 'text-green-600';
    if (percentage >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getComplianceBgColor = (percentage: number): string => {
    if (percentage >= 90) return 'bg-green-100';
    if (percentage >= 70) return 'bg-yellow-100';
    return 'bg-red-100';
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {/* Average Compliance */}
      <div className={`p-4 rounded-lg ${getComplianceBgColor(metrics.average_compliance_percentage)}`}>
        <div className="flex items-center gap-3">
          <div className={`p-2 rounded-full ${metrics.average_compliance_percentage >= 90 ? 'bg-green-200' : metrics.average_compliance_percentage >= 70 ? 'bg-yellow-200' : 'bg-red-200'}`}>
            <CheckCircleIcon className={`h-6 w-6 ${getComplianceColor(metrics.average_compliance_percentage)}`} />
          </div>
          <div>
            <p className="text-sm text-gray-600">Average Compliance</p>
            <p className={`text-2xl font-bold ${getComplianceColor(metrics.average_compliance_percentage)}`}>
              {metrics.average_compliance_percentage.toFixed(1)}%
            </p>
          </div>
        </div>
        <p className="text-xs text-gray-500 mt-2">
          {metrics.lines_with_schedule} of {metrics.total_lines} lines have schedules
        </p>
      </div>

      {/* Scheduled vs Actual Hours */}
      <div className="p-4 rounded-lg bg-blue-50">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-full bg-blue-200">
            <ClockIcon className="h-6 w-6 text-blue-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">Scheduled Uptime</p>
            <p className="text-2xl font-bold text-blue-600">
              {formatHours(metrics.total_scheduled_hours)}
            </p>
          </div>
        </div>
        <div className="mt-2 text-sm">
          <span className="text-gray-600">Actual: </span>
          <span className="font-medium text-blue-700">
            {formatHours(metrics.total_actual_hours)}
          </span>
        </div>
      </div>

      {/* Unplanned Downtime */}
      <div className="p-4 rounded-lg bg-red-50">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-full bg-red-200">
            <ExclamationTriangleIcon className="h-6 w-6 text-red-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">Unplanned Downtime</p>
            <p className="text-2xl font-bold text-red-600">
              {formatHours(metrics.total_unplanned_downtime_hours)}
            </p>
          </div>
        </div>
        <p className="text-xs text-gray-500 mt-2">
          During scheduled production hours
        </p>
      </div>

      {/* Overtime */}
      <div className="p-4 rounded-lg bg-purple-50">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-full bg-purple-200">
            <ArrowTrendingUpIcon className="h-6 w-6 text-purple-600" />
          </div>
          <div>
            <p className="text-sm text-gray-600">Overtime Production</p>
            <p className="text-2xl font-bold text-purple-600">
              {formatHours(metrics.total_overtime_hours)}
            </p>
          </div>
        </div>
        <p className="text-xs text-gray-500 mt-2">
          Outside scheduled hours
        </p>
      </div>
    </div>
  );
}
