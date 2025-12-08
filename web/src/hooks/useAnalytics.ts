import { useQuery } from '@tanstack/react-query';
import { analyticsApi } from '@/api/analytics';
import type { AnalyticsFilters } from '@/api/types';

// Query keys for analytics cache management
export const analyticsKeys = {
  all: ['analytics'] as const,
  aggregate: (filters: AnalyticsFilters) => [...analyticsKeys.all, 'aggregate', filters] as const,
  lineMetrics: (filters: AnalyticsFilters) => [...analyticsKeys.all, 'line-metrics', filters] as const,
  labelMetrics: (filters: AnalyticsFilters) => [...analyticsKeys.all, 'label-metrics', filters] as const,
  dailyKPIs: (lineId: string, filters: AnalyticsFilters) =>
    [...analyticsKeys.all, 'daily-kpis', lineId, filters] as const,
};

// Get aggregate metrics with filters
export function useAggregateMetrics(filters: AnalyticsFilters) {
  return useQuery({
    queryKey: analyticsKeys.aggregate(filters),
    queryFn: () => analyticsApi.getAggregateMetrics(filters),
    staleTime: 10000, // Analytics are less time-sensitive
    refetchOnWindowFocus: false, // Don't refetch on window focus for performance
  });
}

// Get metrics for each production line
export function useLineMetrics(filters: AnalyticsFilters) {
  return useQuery({
    queryKey: analyticsKeys.lineMetrics(filters),
    queryFn: () => analyticsApi.getLineMetrics(filters),
    staleTime: 10000,
    refetchOnWindowFocus: false,
  });
}

// Get metrics grouped by label
export function useLabelMetrics(filters: AnalyticsFilters) {
  return useQuery({
    queryKey: analyticsKeys.labelMetrics(filters),
    queryFn: () => analyticsApi.getLabelMetrics(filters),
    staleTime: 10000,
    refetchOnWindowFocus: false,
  });
}

// Get daily KPIs for a specific line
export function useDailyKPIs(lineId: string, filters: AnalyticsFilters) {
  return useQuery({
    queryKey: analyticsKeys.dailyKPIs(lineId, filters),
    queryFn: () => analyticsApi.getDailyKPIs(lineId, filters),
    enabled: !!lineId,
    staleTime: 10000,
    refetchOnWindowFocus: false,
  });
}
