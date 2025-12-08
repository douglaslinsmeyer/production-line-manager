import { useQuery } from '@tanstack/react-query';
import { complianceApi } from '@/api/compliance';
import type {
  AggregateComplianceMetrics,
  LineComplianceMetrics,
  DailyComplianceKPI,
} from '@/api/types';

// Query keys for cache management
export const complianceKeys = {
  all: ['compliance'] as const,
  aggregate: (startDate: string, endDate: string, lineIds?: string[], labelIds?: string[]) =>
    [...complianceKeys.all, 'aggregate', startDate, endDate, lineIds, labelIds] as const,
  lines: (startDate: string, endDate: string, lineIds?: string[], labelIds?: string[]) =>
    [...complianceKeys.all, 'lines', startDate, endDate, lineIds, labelIds] as const,
  daily: (lineId: string, startDate: string, endDate: string) =>
    [...complianceKeys.all, 'daily', lineId, startDate, endDate] as const,
};

// Hooks

export function useAggregateCompliance(
  startDate: string,
  endDate: string,
  lineIds?: string[],
  labelIds?: string[],
  enabled = true
) {
  return useQuery<AggregateComplianceMetrics, Error>({
    queryKey: complianceKeys.aggregate(startDate, endDate, lineIds, labelIds),
    queryFn: () => complianceApi.getAggregateCompliance(startDate, endDate, lineIds, labelIds),
    enabled: enabled && !!startDate && !!endDate,
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function useLineCompliance(
  startDate: string,
  endDate: string,
  lineIds?: string[],
  labelIds?: string[],
  enabled = true
) {
  return useQuery<LineComplianceMetrics[], Error>({
    queryKey: complianceKeys.lines(startDate, endDate, lineIds, labelIds),
    queryFn: () => complianceApi.getLineCompliance(startDate, endDate, lineIds, labelIds),
    enabled: enabled && !!startDate && !!endDate,
    staleTime: 30 * 1000, // 30 seconds
  });
}

export function useDailyComplianceKPIs(
  lineId: string,
  startDate: string,
  endDate: string,
  enabled = true
) {
  return useQuery<DailyComplianceKPI[], Error>({
    queryKey: complianceKeys.daily(lineId, startDate, endDate),
    queryFn: () => complianceApi.getDailyComplianceKPIs(lineId, startDate, endDate),
    enabled: enabled && !!lineId && !!startDate && !!endDate,
    staleTime: 30 * 1000, // 30 seconds
  });
}
