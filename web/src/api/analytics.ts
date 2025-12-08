import client from './client';
import type {
  APIResponse,
  AnalyticsFilters,
  AggregateMetrics,
  LineMetrics,
  LabelMetrics,
  DailyKPI,
} from './types';

// Analytics API
export const analyticsApi = {
  // Get aggregate metrics across all or filtered lines
  getAggregateMetrics: async (filters: AnalyticsFilters): Promise<AggregateMetrics> => {
    const params = buildAnalyticsParams(filters);
    const response = await client.get<APIResponse<AggregateMetrics>>(
      `/analytics/aggregate?${params.toString()}`
    );
    if (!response.data.data) {
      throw new Error('Failed to fetch aggregate metrics');
    }
    return response.data.data;
  },

  // Get metrics for each production line
  getLineMetrics: async (filters: AnalyticsFilters): Promise<LineMetrics[]> => {
    const params = buildAnalyticsParams(filters);
    const response = await client.get<APIResponse<LineMetrics[]>>(
      `/analytics/lines?${params.toString()}`
    );
    return response.data.data || [];
  },

  // Get metrics grouped by label
  getLabelMetrics: async (filters: AnalyticsFilters): Promise<LabelMetrics[]> => {
    const params = buildAnalyticsParams(filters);
    const response = await client.get<APIResponse<LabelMetrics[]>>(
      `/analytics/labels?${params.toString()}`
    );
    return response.data.data || [];
  },

  // Get daily KPIs for a specific line
  getDailyKPIs: async (lineId: string, filters: AnalyticsFilters): Promise<DailyKPI[]> => {
    const params = buildAnalyticsParams(filters);
    const response = await client.get<APIResponse<DailyKPI[]>>(
      `/analytics/lines/${lineId}/daily?${params.toString()}`
    );
    return response.data.data || [];
  },
};

// Helper to build query parameters for analytics requests
function buildAnalyticsParams(filters: AnalyticsFilters): URLSearchParams {
  const params = new URLSearchParams();

  // Add timeframe
  if (filters.timeframe) {
    params.set('timeframe', filters.timeframe);
  }

  // Add custom time range if applicable
  if (filters.timeframe === 'custom') {
    if (filters.start_time) {
      params.set('start_time', filters.start_time);
    }
    if (filters.end_time) {
      params.set('end_time', filters.end_time);
    }
  }

  // Add label filters (comma-separated)
  if (filters.label_ids && filters.label_ids.length > 0) {
    params.set('label_ids', filters.label_ids.join(','));
  }

  return params;
}
