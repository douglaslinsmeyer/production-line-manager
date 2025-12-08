import client from './client';
import type {
  APIResponse,
  ProductionLine,
  CreateLineRequest,
  UpdateLineRequest,
  SetStatusRequest,
  StatusChange,
  Status,
  Label,
  AssignLabelsRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
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

  // Get labels for a production line
  getLabelsForLine: async (id: string): Promise<Label[]> => {
    const response = await client.get<APIResponse<Label[]>>(`/lines/${id}/labels`);
    return response.data.data || [];
  },

  // Assign labels to a production line
  assignLabels: async (id: string, labelIds: string[]): Promise<Label[]> => {
    const request: AssignLabelsRequest = { label_ids: labelIds };
    const response = await client.put<APIResponse<Label[]>>(`/lines/${id}/labels`, request);
    return response.data.data || [];
  },
};

// Labels API
export const labelsApi = {
  // Get all labels
  getLabels: async (): Promise<Label[]> => {
    const response = await client.get<APIResponse<Label[]>>('/labels');
    return response.data.data || [];
  },

  // Get single label by ID
  getLabel: async (id: string): Promise<Label> => {
    const response = await client.get<APIResponse<Label>>(`/labels/${id}`);
    if (!response.data.data) {
      throw new Error('Label not found');
    }
    return response.data.data;
  },

  // Create new label
  createLabel: async (data: CreateLabelRequest): Promise<Label> => {
    const response = await client.post<APIResponse<Label>>('/labels', data);
    if (!response.data.data) {
      throw new Error('Failed to create label');
    }
    return response.data.data;
  },

  // Update label
  updateLabel: async (id: string, data: UpdateLabelRequest): Promise<Label> => {
    const response = await client.put<APIResponse<Label>>(`/labels/${id}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update label');
    }
    return response.data.data;
  },

  // Delete label
  deleteLabel: async (id: string): Promise<void> => {
    await client.delete(`/labels/${id}`);
  },
};
