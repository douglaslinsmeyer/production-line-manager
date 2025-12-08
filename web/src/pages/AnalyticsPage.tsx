import { useState } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import Card from '@/components/common/Card';
import Button from '@/components/common/Button';
import Loading from '@/components/common/Loading';
import LabelFilter from '@/components/labels/LabelFilter';
import TimeRangeSelector from '@/components/analytics/TimeRangeSelector';
import SystemMetrics from '@/components/analytics/SystemMetrics';
import LabelComparisonChart from '@/components/analytics/LabelComparisonChart';
import { useAggregateMetrics, useLabelMetrics, useLineMetrics } from '@/hooks/useAnalytics';
import { useLabels } from '@/hooks/useLines';
import { formatDuration } from '@/utils/formatters';
import type { AnalyticsFilters, TimeRange, Status, Label } from '@/api/types';

const statusColors: Record<Status, string> = {
  on: '#10b981',
  off: '#6b7280',
  maintenance: '#f59e0b',
  error: '#ef4444',
};

export default function AnalyticsPage() {
  const [filters, setFilters] = useState<AnalyticsFilters>({
    label_ids: [],
    timeframe: '7d',
  });

  const { data: availableLabels = [], isLoading: labelsLoading } = useLabels();
  const { data: aggregateMetrics, isLoading: aggregateLoading, error: aggregateError } = useAggregateMetrics(filters);
  const { data: labelMetrics = [], isLoading: labelMetricsLoading } = useLabelMetrics(filters);
  const { data: lineMetrics = [], isLoading: lineMetricsLoading } = useLineMetrics(filters);

  const selectedLabels = availableLabels.filter((label) => filters.label_ids.includes(label.id));

  const handleLabelChange = (labels: Label[]) => {
    setFilters((prev) => ({
      ...prev,
      label_ids: labels.map((l) => l.id),
    }));
  };

  const handleTimeRangeChange = (timeRange: TimeRange) => {
    setFilters((prev) => ({
      ...prev,
      timeframe: timeRange,
    }));
  };

  const handleClearFilters = () => {
    setFilters({
      label_ids: [],
      timeframe: '7d',
    });
  };

  const isLoading = aggregateLoading || labelMetricsLoading || lineMetricsLoading || labelsLoading;

  if (isLoading) {
    return <Loading message="Loading analytics..." />;
  }

  if (aggregateError) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-6">
        <h3 className="text-red-800 font-semibold mb-2">Error Loading Analytics</h3>
        <p className="text-red-600">
          {aggregateError instanceof Error ? aggregateError.message : 'Failed to load analytics'}
        </p>
      </div>
    );
  }

  if (!aggregateMetrics) {
    return <div>No data available</div>;
  }

  // Prepare status distribution data for pie chart
  const statusDistribution = Object.entries(aggregateMetrics.status_distribution || {})
    .filter(([_, value]) => value > 0)
    .map(([status, value]) => ({
      name: status.charAt(0).toUpperCase() + status.slice(1),
      value: value,
      color: statusColors[status as Status],
    }));

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Analytics Dashboard</h1>
          <p className="mt-1 text-gray-500">Monitor performance across all production lines</p>
        </div>
      </div>

      {/* Filters */}
      <Card>
        <div className="space-y-4">
          <h3 className="text-lg font-semibold">Filters</h3>
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">Labels</label>
              <LabelFilter
                selectedLabels={selectedLabels}
                onLabelsChange={handleLabelChange}
                availableLabels={availableLabels}
              />
            </div>
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">Time Range</label>
              <TimeRangeSelector
                value={filters.timeframe}
                onChange={handleTimeRangeChange}
              />
            </div>
            <div className="flex items-end">
              <Button variant="secondary" onClick={handleClearFilters}>
                Clear Filters
              </Button>
            </div>
          </div>
        </div>
      </Card>

      {/* System Metrics */}
      <Card>
        <h2 className="text-xl font-semibold mb-4">System Overview</h2>
        <SystemMetrics metrics={aggregateMetrics} />
      </Card>

      {/* Status Distribution & MTTR */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <h3 className="text-lg font-semibold mb-4">Status Distribution</h3>
          {statusDistribution.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={statusDistribution}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, value }) => `${name}: ${value.toFixed(1)}%`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {statusDistribution.map((entry, index) => (
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
                          <p className="text-sm">Percentage: {data.value.toFixed(1)}%</p>
                        </div>
                      );
                    }
                    return null;
                  }}
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center py-12">
              <p className="text-gray-500">No status data available</p>
            </div>
          )}
        </Card>

        <Card>
          <h3 className="text-lg font-semibold mb-4">Average MTTR</h3>
          <div className="flex items-center justify-center py-12">
            <div className="text-center">
              <p className="text-5xl font-bold text-blue-600">
                {formatDuration(aggregateMetrics.mttr_hours * 3600000)}
              </p>
              <p className="text-sm text-gray-600 mt-2">Mean Time To Repair</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Label Comparison */}
      {labelMetrics.length > 0 && (
        <Card>
          <LabelComparisonChart labelMetrics={labelMetrics} />
        </Card>
      )}

      {/* Line Metrics Table */}
      {lineMetrics.length > 0 && (
        <Card>
          <h3 className="text-lg font-semibold mb-4">Production Line Performance</h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Line
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Labels
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Uptime
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    MTTR
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Interruptions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {lineMetrics.map((line) => (
                  <tr key={line.line_id}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{line.line_name}</div>
                        <div className="text-sm text-gray-500">{line.line_code}</div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {line.labels.map((label) => (
                          <span
                            key={label.id}
                            className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800"
                          >
                            {label.name}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {line.uptime_percentage.toFixed(1)}%
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {formatDuration(line.mttr_hours * 3600000)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {line.interruption_count}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {/* Empty state when no data matches filters */}
      {filters.label_ids.length > 0 && lineMetrics.length === 0 && (
        <Card>
          <div className="text-center py-12">
            <p className="text-gray-500 text-lg">No lines match the selected labels</p>
            <Button variant="secondary" onClick={handleClearFilters} className="mt-4">
              Clear Filters
            </Button>
          </div>
        </Card>
      )}
    </div>
  );
}
