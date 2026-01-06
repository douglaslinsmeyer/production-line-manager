import client from './client';
import type {
  APIResponse,
  SignalProfile,
  CreateSignalProfileRequest,
  UpdateSignalProfileRequest,
  ProfileVersion,
  ProfileDeviceStatusResponse
} from './types';

export const signalProfilesApi = {
  // Profile CRUD
  getAll: async (): Promise<SignalProfile[]> => {
    const response = await client.get<APIResponse<SignalProfile[]>>('/profiles');
    return response.data.data || [];
  },

  getOne: async (id: string): Promise<SignalProfile> => {
    const response = await client.get<APIResponse<SignalProfile>>(`/profiles/${id}`);
    return response.data.data!;
  },

  create: async (data: CreateSignalProfileRequest): Promise<SignalProfile> => {
    const response = await client.post<APIResponse<SignalProfile>>('/profiles', data);
    return response.data.data!;
  },

  update: async (id: string, data: UpdateSignalProfileRequest): Promise<SignalProfile> => {
    const response = await client.put<APIResponse<SignalProfile>>(`/profiles/${id}`, data);
    return response.data.data!;
  },

  delete: async (id: string): Promise<void> => {
    await client.delete(`/profiles/${id}`);
  },

  // Version management
  getVersions: async (id: string): Promise<ProfileVersion[]> => {
    const response = await client.get<APIResponse<ProfileVersion[]>>(`/profiles/${id}/versions`);
    return response.data.data || [];
  },

  rollback: async (id: string, targetVersion: number, reason: string): Promise<SignalProfile> => {
    const response = await client.post<APIResponse<SignalProfile>>(`/profiles/${id}/rollback`, {
      targetVersion,
      reason
    });
    return response.data.data!;
  },

  getDeviceStatus: async (id: string): Promise<ProfileDeviceStatusResponse> => {
    const response = await client.get<APIResponse<ProfileDeviceStatusResponse>>(`/profiles/${id}/device-status`);
    return response.data.data!;
  },

  // Line assignment
  assignToLine: async (lineId: string, profileId: string): Promise<void> => {
    await client.put(`/lines/${lineId}/profile`, { profile_id: profileId });
  },

  getLineProfile: async (lineId: string): Promise<SignalProfile> => {
    const response = await client.get<APIResponse<SignalProfile>>(`/lines/${lineId}/profile`);
    return response.data.data!;
  }
};
