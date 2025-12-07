// Production Line domain types matching the Go backend

export type Status = 'on' | 'off' | 'maintenance' | 'error';

export interface ProductionLine {
  id: string;
  code: string;
  name: string;
  description?: string;
  status: Status;
  created_at: string;
  updated_at: string;
}

export interface CreateLineRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateLineRequest {
  name?: string;
  description?: string;
}

export interface SetStatusRequest {
  status: Status;
}

export interface StatusChange {
  time: string;
  line_id: string;
  line_code: string;
  old_status?: Status;
  new_status: Status;
  source: string;
  source_detail?: any;
}

// API Response envelope
export interface APIResponse<T> {
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  meta?: {
    total: number;
  };
}
