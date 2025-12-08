import type { AggregateMetrics } from '@/api/types';

interface SystemMetricsProps {
  metrics: AggregateMetrics;
}

export default function SystemMetrics({ metrics }: SystemMetricsProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
      {/* Total Lines */}
      <div className="bg-gradient-to-br from-blue-50 to-blue-100 p-6 rounded-lg">
        <h3 className="text-sm font-medium text-blue-900 mb-2">Total Lines</h3>
        <p className="text-3xl font-bold text-blue-700">{metrics.total_lines}</p>
        <p className="text-xs text-blue-600 mt-1">Production lines in system</p>
      </div>

      {/* Overall Uptime */}
      <div className="bg-gradient-to-br from-green-50 to-green-100 p-6 rounded-lg">
        <h3 className="text-sm font-medium text-green-900 mb-2">Overall Uptime</h3>
        <p className="text-3xl font-bold text-green-700">
          {metrics.average_uptime_percentage.toFixed(1)}%
        </p>
        <p className="text-xs text-green-600 mt-1">Average across all lines</p>
      </div>

      {/* Total Interruptions */}
      <div className="bg-gradient-to-br from-purple-50 to-purple-100 p-6 rounded-lg">
        <h3 className="text-sm font-medium text-purple-900 mb-2">Interruptions</h3>
        <p className="text-3xl font-bold text-purple-700">{metrics.total_interruptions}</p>
        <p className="text-xs text-purple-600 mt-1">Total status changes</p>
      </div>
    </div>
  );
}
