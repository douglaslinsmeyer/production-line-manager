import client from './client';
import type {
  APIResponse,
  ProductionLine,
  CreateLineRequest,
  UpdateLineRequest,
  SetStatusRequest,
  StatusChange,
  Status,
} from './types';

// Production Lines API
export const linesApi = {
  // Get all production lines
  getLines: async (): Promise<ProductionLine[]> => {
    const response = await client.get<APIResponse<ProductionLine[]>>('/lines');
    return response.data.data || [];
  },

  // Get single production line by ID
  getLine: async (id: string): Promise<ProductionLine> => {
    const response = await client.get<APIResponse<ProductionLine>>(`/lines/${id}`);
    if (!response.data.data) {
      throw new Error('Production line not found');
    }
    return response.data.data;
  },

  // Create new production line
  createLine: async (data: CreateLineRequest): Promise<ProductionLine> => {
    const response = await client.post<APIResponse<ProductionLine>>('/lines', data);
    if (!response.data.data) {
      throw new Error('Failed to create production line');
    }
    return response.data.data;
  },

  // Update production line
  updateLine: async (id: string, data: UpdateLineRequest): Promise<ProductionLine> => {
    const response = await client.put<APIResponse<ProductionLine>>(`/lines/${id}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update production line');
    }
    return response.data.data;
  },

  // Delete production line (soft delete)
  deleteLine: async (id: string): Promise<void> => {
    await client.delete(`/lines/${id}`);
  },

  // Set production line status
  setStatus: async (id: string, status: Status): Promise<ProductionLine> => {
    const request: SetStatusRequest = { status };
    const response = await client.post<APIResponse<ProductionLine>>(`/lines/${id}/status`, request);
    if (!response.data.data) {
      throw new Error('Failed to set status');
    }
    return response.data.data;
  },

  // Get status change history
  getHistory: async (id: string, limit = 100): Promise<StatusChange[]> => {
    const response = await client.get<APIResponse<StatusChange[]>>(
      `/lines/${id}/status/history?limit=${limit}`
    );
    return response.data.data || [];
  },
};
