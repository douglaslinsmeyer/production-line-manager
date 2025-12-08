import client from './client';
import type {
  AggregateComplianceMetrics,
  LineComplianceMetrics,
  DailyComplianceKPI,
} from './types';

// Compliance API functions

export const complianceApi = {
  // Get aggregate compliance metrics for all lines
  getAggregateCompliance: async (
    startDate: string,
    endDate: string,
    lineIds?: string[],
    labelIds?: string[]
  ): Promise<AggregateComplianceMetrics> => {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    if (lineIds && lineIds.length > 0) {
      params.append('line_ids', lineIds.join(','));
    }

    if (labelIds && labelIds.length > 0) {
      params.append('label_ids', labelIds.join(','));
    }

    const response = await client.get<AggregateComplianceMetrics>(`/analytics/compliance?${params.toString()}`);
    return response.data;
  },

  // Get compliance metrics per line
  getLineCompliance: async (
    startDate: string,
    endDate: string,
    lineIds?: string[],
    labelIds?: string[]
  ): Promise<LineComplianceMetrics[]> => {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    if (lineIds && lineIds.length > 0) {
      params.append('line_ids', lineIds.join(','));
    }

    if (labelIds && labelIds.length > 0) {
      params.append('label_ids', labelIds.join(','));
    }

    const response = await client.get<LineComplianceMetrics[]>(`/analytics/compliance/lines?${params.toString()}`);
    return response.data;
  },

  // Get daily compliance KPIs for a specific line
  getDailyComplianceKPIs: async (
    lineId: string,
    startDate: string,
    endDate: string
  ): Promise<DailyComplianceKPI[]> => {
    const params = new URLSearchParams({
      start_date: startDate,
      end_date: endDate,
    });

    const response = await client.get<DailyComplianceKPI[]>(`/analytics/lines/${lineId}/compliance/daily?${params.toString()}`);
    return response.data;
  },
};
