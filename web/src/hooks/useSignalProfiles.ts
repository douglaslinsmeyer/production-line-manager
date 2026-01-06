import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { signalProfilesApi } from '../api/signal-profiles';
import { useSSE } from './useSSE';
import type { CreateSignalProfileRequest, UpdateSignalProfileRequest } from '../api/types';
import toast from 'react-hot-toast';

// Query keys
const signalProfileKeys = {
  all: ['signal-profiles'] as const,
  lists: () => [...signalProfileKeys.all, 'list'] as const,
  list: () => [...signalProfileKeys.lists()] as const,
  details: () => [...signalProfileKeys.all, 'detail'] as const,
  detail: (id: string) => [...signalProfileKeys.details(), id] as const,
  versions: (id: string) => [...signalProfileKeys.all, 'versions', id] as const,
  deviceStatus: (id: string) => [...signalProfileKeys.all, 'device-status', id] as const,
};

// Get all profiles
export function useSignalProfiles() {
  const result = useQuery({
    queryKey: signalProfileKeys.list(),
    queryFn: signalProfilesApi.getAll,
    staleTime: 60000,
  });

  // Real-time updates via SSE
  const queryClient = useQueryClient();
  useSSE('profile.created', () => {
    queryClient.invalidateQueries({ queryKey: signalProfileKeys.list() });
  }, true);
  useSSE('profile.updated', () => {
    queryClient.invalidateQueries({ queryKey: signalProfileKeys.all });
  }, true);
  useSSE('profile.deleted', () => {
    queryClient.invalidateQueries({ queryKey: signalProfileKeys.list() });
  }, true);

  return result;
}

// Get single profile
export function useSignalProfile(id: string) {
  return useQuery({
    queryKey: signalProfileKeys.detail(id),
    queryFn: () => signalProfilesApi.getOne(id),
    staleTime: 30000,
    enabled: !!id,
  });
}

// Create profile
export function useCreateSignalProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSignalProfileRequest) => signalProfilesApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: signalProfileKeys.list() });
      toast.success('Profile created successfully');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to create profile');
    },
  });
}

// Update profile
export function useUpdateSignalProfile(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateSignalProfileRequest) => signalProfilesApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: signalProfileKeys.all });
      toast.success('Profile updated successfully');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to update profile');
    },
  });
}

// Delete profile
export function useDeleteSignalProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => signalProfilesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: signalProfileKeys.list() });
      toast.success('Profile deleted successfully');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete profile');
    },
  });
}

// Get version history
export function useProfileVersions(id: string) {
  return useQuery({
    queryKey: signalProfileKeys.versions(id),
    queryFn: () => signalProfilesApi.getVersions(id),
    enabled: !!id,
  });
}

// Get device sync status
export function useProfileDeviceStatus(id: string) {
  return useQuery({
    queryKey: signalProfileKeys.deviceStatus(id),
    queryFn: () => signalProfilesApi.getDeviceStatus(id),
    refetchInterval: 30000, // Refresh every 30s
    enabled: !!id,
  });
}

// Assign profile to line
export function useAssignProfileToLine() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ lineId, profileId }: { lineId: string; profileId: string }) =>
      signalProfilesApi.assignToLine(lineId, profileId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lines'] });
      toast.success('Profile assigned to line');
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to assign profile');
    },
  });
}

// Get line's assigned profile
export function useLineProfile(lineId: string) {
  return useQuery({
    queryKey: ['lines', lineId, 'profile'],
    queryFn: () => signalProfilesApi.getLineProfile(lineId),
    enabled: !!lineId,
    retry: false, // Don't retry if line has no profile
  });
}
