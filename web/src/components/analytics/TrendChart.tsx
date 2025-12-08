import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { DailyKPI } from '@/api/types';

interface TrendChartProps {
  dailyKPIs: DailyKPI[];
  metric?: 'uptime' | 'mttr' | 'interruptions';
}

export default function TrendChart({ dailyKPIs, metric = 'uptime' }: TrendChartProps) {
  if (dailyKPIs.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">No trend data available</p>
      </div>
    );
  }

  const chartData = dailyKPIs.map((kpi) => ({
    date: new Date(kpi.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    uptime: parseFloat(kpi.uptime_percentage.toFixed(1)),
    mttr: parseFloat(kpi.mttr_hours.toFixed(2)),
    interruptions: kpi.interruption_count,
  }));

  const getMetricConfig = () => {
    switch (metric) {
      case 'uptime':
        return {
          dataKey: 'uptime',
          name: 'Uptime %',
          color: '#10b981',
          unit: '%',
        };
      case 'mttr':
        return {
          dataKey: 'mttr',
          name: 'MTTR (hours)',
          color: '#3b82f6',
          unit: 'h',
        };
      case 'interruptions':
        return {
          dataKey: 'interruptions',
          name: 'Interruptions',
          color: '#ef4444',
          unit: '',
        };
    }
  };

  const config = getMetricConfig();

  return (
    <div>
      <h3 className="text-lg font-semibold mb-4">Trend Over Time</h3>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="date" />
          <YAxis unit={config.unit} />
          <Tooltip
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                    <p className="text-sm font-semibold">{data.date}</p>
                    <p className="text-sm">{config.name}: {payload[0].value}{config.unit}</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey={config.dataKey}
            stroke={config.color}
            strokeWidth={2}
            name={config.name}
            dot={{ fill: config.color, r: 4 }}
            activeDot={{ r: 6 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
