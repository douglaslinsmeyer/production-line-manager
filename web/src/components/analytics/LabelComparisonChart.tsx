import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import type { LabelMetrics } from '@/api/types';

interface LabelComparisonChartProps {
  labelMetrics: LabelMetrics[];
}

export default function LabelComparisonChart({ labelMetrics }: LabelComparisonChartProps) {
  if (labelMetrics.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">No label data available</p>
      </div>
    );
  }

  const chartData = labelMetrics.map((lm) => ({
    label: lm.label.name,
    uptime: parseFloat(lm.average_uptime_percentage.toFixed(1)),
    lineCount: lm.line_count,
    mttr: lm.mttr_hours.toFixed(2),
  }));

  return (
    <div>
      <h3 className="text-lg font-semibold mb-4">Uptime Comparison by Label</h3>
      <ResponsiveContainer width="100%" height={Math.max(300, labelMetrics.length * 60)}>
        <BarChart
          data={chartData}
          layout="vertical"
          margin={{ top: 5, right: 30, left: 100, bottom: 5 }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis type="number" domain={[0, 100]} unit="%" />
          <YAxis dataKey="label" type="category" width={90} />
          <Tooltip
            content={({ active, payload }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload;
                return (
                  <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                    <p className="text-sm font-semibold">{data.label}</p>
                    <p className="text-sm">Uptime: {data.uptime}%</p>
                    <p className="text-sm">Lines: {data.lineCount}</p>
                    <p className="text-sm">MTTR: {data.mttr}h</p>
                  </div>
                );
              }
              return null;
            }}
          />
          <Legend />
          <Bar dataKey="uptime" fill="#3b82f6" name="Uptime %" radius={[0, 4, 4, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
