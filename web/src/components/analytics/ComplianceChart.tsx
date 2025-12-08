import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import type { LineComplianceMetrics } from '@/api/types';

interface ComplianceChartProps {
  lineMetrics: LineComplianceMetrics[];
}

export default function ComplianceChart({ lineMetrics }: ComplianceChartProps) {
  // Filter to only lines with schedules and sort by compliance
  const chartData = lineMetrics
    .filter((m) => m.schedule_id)
    .map((m) => ({
      name: m.line_code,
      lineName: m.line_name,
      scheduled: m.scheduled_uptime_hours,
      actual: m.actual_uptime_hours,
      compliance: m.compliance_percentage,
      unplanned: m.unplanned_downtime_hours,
      overtime: m.overtime_hours,
    }))
    .sort((a, b) => b.compliance - a.compliance)
    .slice(0, 10); // Show top 10 lines

  if (chartData.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">No lines with schedules to display</p>
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={400}>
      <BarChart
        data={chartData}
        margin={{
          top: 20,
          right: 30,
          left: 20,
          bottom: 5,
        }}
      >
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="name" />
        <YAxis
          label={{ value: 'Hours', angle: -90, position: 'insideLeft' }}
        />
        <Tooltip
          content={({ active, payload, label }) => {
            if (active && payload && payload.length) {
              const data = payload[0].payload;
              return (
                <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
                  <p className="font-semibold">{data.lineName}</p>
                  <p className="text-sm text-gray-600">Code: {label}</p>
                  <div className="mt-2 space-y-1 text-sm">
                    <p>
                      <span className="text-blue-600">Scheduled:</span>{' '}
                      {data.scheduled.toFixed(1)}h
                    </p>
                    <p>
                      <span className="text-green-600">Actual:</span>{' '}
                      {data.actual.toFixed(1)}h
                    </p>
                    <p>
                      <span className="text-gray-600">Compliance:</span>{' '}
                      <span className={data.compliance >= 90 ? 'text-green-600' : data.compliance >= 70 ? 'text-yellow-600' : 'text-red-600'}>
                        {data.compliance.toFixed(1)}%
                      </span>
                    </p>
                  </div>
                </div>
              );
            }
            return null;
          }}
        />
        <Legend />
        <Bar
          dataKey="scheduled"
          name="Scheduled Hours"
          fill="#3b82f6"
          radius={[4, 4, 0, 0]}
        />
        <Bar
          dataKey="actual"
          name="Actual Hours"
          fill="#10b981"
          radius={[4, 4, 0, 0]}
        />
      </BarChart>
    </ResponsiveContainer>
  );
}
