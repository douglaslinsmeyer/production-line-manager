import { useMemo } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import type { StatusChange, Status } from '@/api/types';
import { formatDuration } from '@/utils/formatters';

interface UptimeMetricsProps {
  history: StatusChange[];
}

const statusColors: Record<Status, string> = {
  on: '#10b981',
  off: '#6b7280',
  maintenance: '#f59e0b',
  error: '#ef4444',
};

export default function UptimeMetrics({ history }: UptimeMetricsProps) {
  const metrics = useMemo(() => {
    if (!history || history.length < 2) {
      return {
        uptime: 0,
        mttr: 0,
        distribution: [],
      };
    }

    // Sort by time (oldest first)
    const sorted = [...history].sort(
      (a, b) => new Date(a.time).getTime() - new Date(b.time).getTime()
    );

    let totalTime = 0;
    let uptimeMs = 0;
    let errorTime = 0;
    let maintenanceTime = 0;
    let errorCount = 0;
    let maintenanceCount = 0;

    const statusDurations: Record<Status, number> = {
      on: 0,
      off: 0,
      maintenance: 0,
      error: 0,
    };

    // Calculate time in each status
    for (let i = 0; i < sorted.length - 1; i++) {
      const current = sorted[i];
      const next = sorted[i + 1];

      const duration = new Date(next.time).getTime() - new Date(current.time).getTime();
      totalTime += duration;

      const status = current.new_status;
      statusDurations[status] += duration;

      if (status === 'on') {
        uptimeMs += duration;
      } else if (status === 'error') {
        errorTime += duration;
        errorCount++;
      } else if (status === 'maintenance') {
        maintenanceTime += duration;
        maintenanceCount++;
      }
    }

    // Add time from last status change to now
    const lastChange = sorted[sorted.length - 1];
    const timeSinceLastChange = Date.now() - new Date(lastChange.time).getTime();
    totalTime += timeSinceLastChange;
    statusDurations[lastChange.new_status] += timeSinceLastChange;

    if (lastChange.new_status === 'on') {
      uptimeMs += timeSinceLastChange;
    }

    // Calculate metrics
    const uptimePercent = totalTime > 0 ? (uptimeMs / totalTime) * 100 : 0;
    const mttr =
      errorCount + maintenanceCount > 0
        ? (errorTime + maintenanceTime) / (errorCount + maintenanceCount)
        : 0;

    // Prepare distribution data for pie chart
    const distribution = Object.entries(statusDurations)
      .filter(([_, duration]) => duration > 0)
      .map(([status, duration]) => ({
        name: status.charAt(0).toUpperCase() + status.slice(1),
        value: duration,
        percent: ((duration / totalTime) * 100).toFixed(1),
        color: statusColors[status as Status],
      }))
      .sort((a, b) => b.value - a.value);

    return {
      uptime: uptimePercent,
      mttr,
      distribution,
      totalTime,
    };
  }, [history]);

  if (history.length < 2) {
    return (
      <div className="flex items-center justify-center py-8">
        <p className="text-gray-500">Not enough data for metrics</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Uptime Percentage */}
        <div className="bg-gradient-to-br from-green-50 to-green-100 p-6 rounded-lg">
          <h3 className="text-sm font-medium text-green-900 mb-2">Uptime</h3>
          <p className="text-3xl font-bold text-green-700">
            {metrics.uptime.toFixed(1)}%
          </p>
          <p className="text-xs text-green-600 mt-1">Time in running status</p>
        </div>

        {/* Mean Time To Repair */}
        <div className="bg-gradient-to-br from-blue-50 to-blue-100 p-6 rounded-lg">
          <h3 className="text-sm font-medium text-blue-900 mb-2">MTTR</h3>
          <p className="text-3xl font-bold text-blue-700">
            {formatDuration(metrics.mttr)}
          </p>
          <p className="text-xs text-blue-600 mt-1">Mean time to repair</p>
        </div>
      </div>

      {/* Status Distribution Pie Chart */}
      <div>
        <h3 className="text-lg font-semibold mb-4">Status Distribution</h3>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={metrics.distribution}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name}: ${percent}%`}
              outerRadius={100}
              fill="#8884d8"
              dataKey="value"
            >
              {metrics.distribution.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip
              content={({ active, payload }) => {
                if (active && payload && payload.length) {
                  const data = payload[0].payload;
                  return (
                    <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                      <p className="text-sm font-semibold">{data.name}</p>
                      <p className="text-sm">Duration: {formatDuration(data.value)}</p>
                      <p className="text-sm">Percentage: {data.percent}%</p>
                    </div>
                  );
                }
                return null;
              }}
            />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
