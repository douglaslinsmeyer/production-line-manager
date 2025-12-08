import { useState } from 'react';
import { formatDuration } from '@/utils/formatters';
import type { DailyKPI } from '@/api/types';

interface DailyKPIsTableProps {
  dailyKPIs: DailyKPI[];
}

type SortField = 'date' | 'uptime_percentage' | 'maintenance_hours' | 'interruption_count' | 'mttr_hours';
type SortDirection = 'asc' | 'desc';

export default function DailyKPIsTable({ dailyKPIs }: DailyKPIsTableProps) {
  const [sortField, setSortField] = useState<SortField>('date');
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

  if (dailyKPIs.length === 0) {
    return (
      <div className="flex items-center justify-center py-12">
        <p className="text-gray-500">No daily KPI data available</p>
      </div>
    );
  }

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const sortedKPIs = [...dailyKPIs].sort((a, b) => {
    let aVal: number | string = a[sortField];
    let bVal: number | string = b[sortField];

    if (sortField === 'date') {
      aVal = new Date(a.date).getTime();
      bVal = new Date(b.date).getTime();
    }

    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : -1;
    } else {
      return aVal < bVal ? 1 : -1;
    }
  });

  const getUptimeColor = (uptime: number): string => {
    if (uptime >= 95) return 'text-green-600 bg-green-50';
    if (uptime >= 80) return 'text-yellow-600 bg-yellow-50';
    return 'text-red-600 bg-red-50';
  };

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th
              onClick={() => handleSort('date')}
              className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
            >
              Date {sortField === 'date' && (sortDirection === 'asc' ? '↑' : '↓')}
            </th>
            <th
              onClick={() => handleSort('uptime_percentage')}
              className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
            >
              Uptime {sortField === 'uptime_percentage' && (sortDirection === 'asc' ? '↑' : '↓')}
            </th>
            <th
              onClick={() => handleSort('maintenance_hours')}
              className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
            >
              Maintenance {sortField === 'maintenance_hours' && (sortDirection === 'asc' ? '↑' : '↓')}
            </th>
            <th
              onClick={() => handleSort('interruption_count')}
              className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
            >
              Interruptions {sortField === 'interruption_count' && (sortDirection === 'asc' ? '↑' : '↓')}
            </th>
            <th
              onClick={() => handleSort('mttr_hours')}
              className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
            >
              MTTR {sortField === 'mttr_hours' && (sortDirection === 'asc' ? '↑' : '↓')}
            </th>
          </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
          {sortedKPIs.map((kpi, index) => (
            <tr key={kpi.date} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
              <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                {new Date(kpi.date).toLocaleDateString('en-US', {
                  month: 'short',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full font-medium ${getUptimeColor(kpi.uptime_percentage)}`}>
                  {kpi.uptime_percentage.toFixed(1)}%
                </span>
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {formatDuration(kpi.maintenance_hours * 3600000)}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {kpi.interruption_count}
              </td>
              <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {kpi.mttr_hours > 0 ? formatDuration(kpi.mttr_hours * 3600000) : '-'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
