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
  schedule_id?: string;
  labels?: Label[];
  schedule?: ScheduleSummary;
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

// Schedule types
export type DayOfWeek = 0 | 1 | 2 | 3 | 4 | 5 | 6; // 0=Sunday, 6=Saturday

export type ScheduleSource = 'base' | 'holiday' | 'schedule_exception' | 'line_exception';

export interface ScheduleBreak {
  id: string;
  schedule_day_id: string;
  name: string;
  break_start: string; // HH:MM:SS format
  break_end: string;
  created_at: string;
  updated_at: string;
}

export interface ScheduleDay {
  id: string;
  schedule_id: string;
  day_of_week: DayOfWeek;
  is_working_day: boolean;
  shift_start?: string; // HH:MM:SS format
  shift_end?: string;
  breaks: ScheduleBreak[];
  created_at: string;
  updated_at: string;
}

export interface Schedule {
  id: string;
  name: string;
  description?: string;
  timezone: string;
  days: ScheduleDay[];
  created_at: string;
  updated_at: string;
}

export interface ScheduleSummary {
  id: string;
  name: string;
  description?: string;
  timezone: string;
  line_count: number;
  created_at: string;
  updated_at: string;
}

export interface ScheduleHoliday {
  id: string;
  schedule_id: string;
  holiday_date: string; // YYYY-MM-DD
  name: string;
  created_at: string;
  updated_at: string;
}

export interface SuggestedHoliday {
  date: string; // YYYY-MM-DD
  name: string;
  type: string; // "Public", "Bank", etc.
  nationwide: boolean;
}

export interface SuggestedHolidaysResponse {
  holidays: SuggestedHoliday[];
  country_code: string;
  year: number;
  cached: boolean;
  error?: string;
}

export interface ScheduleExceptionDay {
  id: string;
  exception_id: string;
  day_of_week: DayOfWeek;
  is_working_day: boolean;
  shift_start?: string;
  shift_end?: string;
  breaks: ScheduleBreak[];
  created_at: string;
  updated_at: string;
}

export interface ScheduleException {
  id: string;
  schedule_id: string;
  name: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;
  days: ScheduleExceptionDay[];
  created_at: string;
  updated_at: string;
}

export interface LineScheduleException {
  id: string;
  schedule_id: string;
  name: string;
  start_date: string;
  end_date: string;
  line_ids: string[];
  days: ScheduleExceptionDay[];
  created_at: string;
  updated_at: string;
}

export interface EffectiveBreak {
  name: string;
  break_start: string;
  break_end: string;
}

export interface EffectiveSchedule {
  line_id: string;
  date: string; // YYYY-MM-DD
  source: ScheduleSource;
  source_id?: string;
  source_name?: string;
  is_working_day: boolean;
  shift_start?: string;
  shift_end?: string;
  breaks: EffectiveBreak[];
}

// Schedule request types
export interface CreateScheduleDayInput {
  day_of_week: DayOfWeek;
  is_working_day: boolean;
  shift_start?: string;
  shift_end?: string;
  breaks?: CreateBreakInput[];
}

export interface CreateBreakInput {
  name: string;
  break_start: string;
  break_end: string;
}

export interface CreateScheduleRequest {
  name: string;
  description?: string;
  timezone?: string;
  days: CreateScheduleDayInput[];
}

export interface UpdateScheduleRequest {
  name?: string;
  description?: string;
  timezone?: string;
}

export interface UpdateScheduleDayRequest {
  is_working_day?: boolean;
  shift_start?: string;
  shift_end?: string;
}

export interface SetScheduleBreaksRequest {
  breaks: CreateBreakInput[];
}

export interface CreateHolidayRequest {
  holiday_date: string;
  name: string;
}

export interface UpdateHolidayRequest {
  holiday_date?: string;
  name?: string;
}

export interface CreateScheduleExceptionRequest {
  name: string;
  start_date: string;
  end_date: string;
  days: CreateScheduleDayInput[];
}

export interface UpdateScheduleExceptionRequest {
  name?: string;
  start_date?: string;
  end_date?: string;
}

export interface CreateLineScheduleExceptionRequest {
  name: string;
  start_date: string;
  end_date: string;
  line_ids: string[];
  days: CreateScheduleDayInput[];
}

export interface UpdateLineScheduleExceptionRequest {
  name?: string;
  start_date?: string;
  end_date?: string;
  line_ids?: string[];
}

export interface AssignScheduleToLineRequest {
  schedule_id: string | null;
}

// ========== Compliance Types ==========

export interface ComplianceQuery {
  start_date: string;
  end_date: string;
  line_ids?: string[];
  label_ids?: string[];
}

export interface LineComplianceMetrics {
  line_id: string;
  line_code: string;
  line_name: string;
  schedule_id?: string;
  schedule_name?: string;
  scheduled_uptime_hours: number;
  actual_uptime_hours: number;
  compliance_percentage: number;
  unplanned_downtime_hours: number;
  overtime_hours: number;
  scheduled_days: number;
  working_days: number;
}

export interface AggregateComplianceMetrics {
  total_lines: number;
  lines_with_schedule: number;
  total_scheduled_hours: number;
  total_actual_hours: number;
  average_compliance_percentage: number;
  total_unplanned_downtime_hours: number;
  total_overtime_hours: number;
  date_range: DateRange;
  line_metrics?: LineComplianceMetrics[];
}

export interface DateRange {
  start_date: string;
  end_date: string;
}

export interface DailyComplianceKPI {
  date: string;
  scheduled_uptime_hours: number;
  actual_uptime_hours: number;
  compliance_percentage: number;
  unplanned_downtime_hours: number;
  overtime_hours: number;
  is_working_day: boolean;
  source: string;
}
