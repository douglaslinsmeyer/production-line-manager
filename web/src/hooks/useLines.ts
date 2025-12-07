import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { linesApi } from '@/api/lines';
import type { CreateLineRequest, UpdateLineRequest, Status } from '@/api/types';

// Query keys for cache management
export const lineKeys = {
  all: ['lines'] as const,
  lists: () => [...lineKeys.all, 'list'] as const,
  list: () => [...lineKeys.lists()] as const,
  details: () => [...lineKeys.all, 'detail'] as const,
  detail: (id: string) => [...lineKeys.details(), id] as const,
  history: (id: string) => [...lineKeys.detail(id), 'history'] as const,
};

// Get all production lines with auto-refetch
export function useLines() {
  return useQuery({
    queryKey: lineKeys.list(),
    queryFn: linesApi.getLines,
    refetchInterval: 5000, // Auto-refetch every 5 seconds for real-time updates
    staleTime: 3000, // Consider data stale after 3 seconds
  });
}

// Get single production line
export function useLine(id: string) {
  return useQuery({
    queryKey: lineKeys.detail(id),
    queryFn: () => linesApi.getLine(id),
    enabled: !!id, // Only run if ID is provided
    refetchInterval: 5000,
  });
}

// Create production line mutation
export function useCreateLine() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateLineRequest) => linesApi.createLine(data),
    onSuccess: () => {
      // Invalidate and refetch lines list
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
    },
  });
}

// Update production line mutation
export function useUpdateLine(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateLineRequest) => linesApi.updateLine(id, data),
    onSuccess: (updatedLine) => {
      // Update cache optimistically
      queryClient.setQueryData(lineKeys.detail(id), updatedLine);
      // Invalidate list to ensure consistency
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
    },
  });
}

// Delete production line mutation
export function useDeleteLine() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => linesApi.deleteLine(id),
    onSuccess: (_data, id) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: lineKeys.detail(id) });
      // Invalidate list
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
    },
  });
}

// Set production line status mutation
export function useSetStatus(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (status: Status) => linesApi.setStatus(id, status),
    onSuccess: (updatedLine) => {
      // Update cache optimistically
      queryClient.setQueryData(lineKeys.detail(id), updatedLine);
      // Invalidate list to update status badge
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
      // Invalidate history to show new entry
      queryClient.invalidateQueries({ queryKey: lineKeys.history(id) });
    },
  });
}

// Get status history
export function useStatusHistory(id: string, limit = 100) {
  return useQuery({
    queryKey: [...lineKeys.history(id), limit],
    queryFn: () => linesApi.getHistory(id, limit),
    enabled: !!id,
    refetchInterval: 10000, // Refetch every 10 seconds
  });
}
