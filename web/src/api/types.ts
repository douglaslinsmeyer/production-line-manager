// Production Line domain types matching the Go backend

export type Status = 'on' | 'off' | 'maintenance' | 'error';

export interface Label {
  id: string;
  name: string;
  color?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface ProductionLine {
  id: string;
  code: string;
  name: string;
  description?: string;
  status: Status;
  labels?: Label[];
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

// Label requests
export interface CreateLabelRequest {
  name: string;
  color?: string;
  description?: string;
}

export interface UpdateLabelRequest {
  name?: string;
  color?: string;
  description?: string;
}

export interface AssignLabelsRequest {
  label_ids: string[];
}

// Analytics types
export type TimeRange = '24h' | '7d' | '30d' | 'all' | 'custom';

export interface AnalyticsFilters {
  label_ids: string[];
  timeframe: TimeRange;
  start_time?: string; // ISO datetime
  end_time?: string; // ISO datetime
}

export interface AggregateMetrics {
  total_lines: number;
  total_uptime_hours: number;
  average_uptime_percentage: number;
  total_downtime_hours: number;
  total_maintenance_hours: number;
  mttr_hours: number;
  total_interruptions: number;
  status_distribution: Record<Status, number>;
  time_range: {
    start: string;
    end: string;
  };
}

export interface LineMetrics {
  line_id: string;
  line_code: string;
  line_name: string;
  labels: Label[];
  uptime_hours: number;
  uptime_percentage: number;
  downtime_hours: number;
  maintenance_hours: number;
  error_hours: number;
  mttr_hours: number;
  interruption_count: number;
  current_status: Status;
  status_distribution: Record<Status, number>;
}

export interface LabelMetrics {
  label: Label;
  line_count: number;
  average_uptime_percentage: number;
  total_uptime_hours: number;
  total_interruptions: number;
  mttr_hours: number;
  status_distribution?: Record<Status, number>;
}

export interface DailyKPI {
  date: string; // YYYY-MM-DD
  uptime_hours: number;
  uptime_percentage: number;
  maintenance_hours: number;
  interruption_count: number;
  mttr_hours: number;
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
