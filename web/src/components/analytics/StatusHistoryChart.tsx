import { useMemo } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { format, parseISO } from 'date-fns';
import type { StatusChange, Status } from '@/api/types';

interface StatusHistoryChartProps {
  history: StatusChange[];
}

// Map status to numeric values for chart
const statusToValue: Record<Status, number> = {
  error: 1,
  off: 2,
  maintenance: 3,
  on: 4,
};

const valueToStatus: Record<number, Status> = {
  1: 'error',
  2: 'off',
  3: 'maintenance',
  4: 'on',
};

export default function StatusHistoryChart({ history }: StatusHistoryChartProps) {
  // Transform status history into chart data
  const chartData = useMemo(() => {
    if (!history || history.length === 0) return [];

    // Sort by time (oldest first for chronological chart)
    const sorted = [...history].sort(
      (a, b) => new Date(a.time).getTime() - new Date(b.time).getTime()
    );

    return sorted.map((change) => ({
      time: new Date(change.time).getTime(),
      timeDisplay: format(parseISO(change.time), 'MMM d, HH:mm'),
      status: change.new_status,
      statusValue: statusToValue[change.new_status],
      source: change.source,
    }));
  }, [history]);

  if (chartData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-gray-50 rounded-lg">
        <p className="text-gray-500">No status history available</p>
      </div>
    );
  }

  return (
    <div className="w-full">
      <ResponsiveContainer width="100%" height={300}>
        <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="statusGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#10b981" stopOpacity={0.8} />
              <stop offset="95%" stopColor="#10b981" stopOpacity={0.1} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200" />
          <XAxis
            dataKey="timeDisplay"
            tick={{ fontSize: 12 }}
            tickLine={false}
          />
          <YAxis
            domain={[0.5, 4.5]}
            ticks={[1, 2, 3, 4]}
            tickFormatter={(value) => {
              const status = valueToStatus[value];
              return status ? status.charAt(0).toUpperCase() + status.slice(1) : '';
            }}
            tick={{ fontSize: 12 }}
            tickLine={false}
          />
          <Tooltip
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                    <p className="text-sm font-semibold">{data.timeDisplay}</p>
                    <p className="text-sm">
                      Status: <span className="font-medium capitalize">{data.status}</span>
                    </p>
                    <p className="text-xs text-gray-500">Source: {data.source}</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Legend
            formatter={(value) => {
              if (value === 'statusValue') return 'Status Level';
              return value;
            }}
          />
          <Area
            type="stepAfter"
            dataKey="statusValue"
            stroke="#10b981"
            strokeWidth={2}
            fill="url(#statusGradient)"
            name="statusValue"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}
