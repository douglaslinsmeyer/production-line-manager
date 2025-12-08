import client from './client';
import type {
  APIResponse,
  Schedule,
  ScheduleSummary,
  ScheduleDay,
  ScheduleBreak,
  ScheduleHoliday,
  SuggestedHolidaysResponse,
  ScheduleException,
  LineScheduleException,
  EffectiveSchedule,
  ProductionLine,
  CreateScheduleRequest,
  UpdateScheduleRequest,
  UpdateScheduleDayRequest,
  SetScheduleBreaksRequest,
  CreateHolidayRequest,
  UpdateHolidayRequest,
  CreateScheduleExceptionRequest,
  UpdateScheduleExceptionRequest,
  CreateLineScheduleExceptionRequest,
  UpdateLineScheduleExceptionRequest,
  AssignScheduleToLineRequest,
} from './types';

// Schedules API
export const schedulesApi = {
  // Get all schedules
  getSchedules: async (): Promise<ScheduleSummary[]> => {
    const response = await client.get<APIResponse<ScheduleSummary[]>>('/schedules');
    return response.data.data || [];
  },

  // Get single schedule by ID (full details)
  getSchedule: async (id: string): Promise<Schedule> => {
    const response = await client.get<APIResponse<Schedule>>(`/schedules/${id}`);
    if (!response.data.data) {
      throw new Error('Schedule not found');
    }
    return response.data.data;
  },

  // Create new schedule
  createSchedule: async (data: CreateScheduleRequest): Promise<Schedule> => {
    const response = await client.post<APIResponse<Schedule>>('/schedules', data);
    if (!response.data.data) {
      throw new Error('Failed to create schedule');
    }
    return response.data.data;
  },

  // Update schedule
  updateSchedule: async (id: string, data: UpdateScheduleRequest): Promise<Schedule> => {
    const response = await client.put<APIResponse<Schedule>>(`/schedules/${id}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update schedule');
    }
    return response.data.data;
  },

  // Delete schedule
  deleteSchedule: async (id: string): Promise<void> => {
    await client.delete(`/schedules/${id}`);
  },

  // Get schedule day
  getDay: async (scheduleId: string, dayId: string): Promise<ScheduleDay> => {
    const response = await client.get<APIResponse<ScheduleDay>>(`/schedules/${scheduleId}/days/${dayId}`);
    if (!response.data.data) {
      throw new Error('Schedule day not found');
    }
    return response.data.data;
  },

  // Update schedule day
  updateDay: async (scheduleId: string, dayId: string, data: UpdateScheduleDayRequest): Promise<ScheduleDay> => {
    const response = await client.put<APIResponse<ScheduleDay>>(`/schedules/${scheduleId}/days/${dayId}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update schedule day');
    }
    return response.data.data;
  },

  // Set day breaks
  setDayBreaks: async (scheduleId: string, dayId: string, data: SetScheduleBreaksRequest): Promise<ScheduleBreak[]> => {
    const response = await client.put<APIResponse<ScheduleBreak[]>>(`/schedules/${scheduleId}/days/${dayId}/breaks`, data);
    return response.data.data || [];
  },

  // Get lines using this schedule
  getLinesForSchedule: async (scheduleId: string): Promise<ProductionLine[]> => {
    const response = await client.get<APIResponse<ProductionLine[]>>(`/schedules/${scheduleId}/lines`);
    return response.data.data || [];
  },
};

// Schedule Holidays API
export const holidaysApi = {
  // Get all holidays for a schedule, optionally filtered by year
  getHolidays: async (scheduleId: string, year?: number): Promise<ScheduleHoliday[]> => {
    const params = year ? `?year=${year}` : '';
    const response = await client.get<APIResponse<ScheduleHoliday[]>>(`/schedules/${scheduleId}/holidays${params}`);
    return response.data.data || [];
  },

  // Get single holiday
  getHoliday: async (scheduleId: string, holidayId: string): Promise<ScheduleHoliday> => {
    const response = await client.get<APIResponse<ScheduleHoliday>>(`/schedules/${scheduleId}/holidays/${holidayId}`);
    if (!response.data.data) {
      throw new Error('Holiday not found');
    }
    return response.data.data;
  },

  // Create holiday
  createHoliday: async (scheduleId: string, data: CreateHolidayRequest): Promise<ScheduleHoliday> => {
    const response = await client.post<APIResponse<ScheduleHoliday>>(`/schedules/${scheduleId}/holidays`, data);
    if (!response.data.data) {
      throw new Error('Failed to create holiday');
    }
    return response.data.data;
  },

  // Update holiday
  updateHoliday: async (scheduleId: string, holidayId: string, data: UpdateHolidayRequest): Promise<ScheduleHoliday> => {
    const response = await client.put<APIResponse<ScheduleHoliday>>(`/schedules/${scheduleId}/holidays/${holidayId}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update holiday');
    }
    return response.data.data;
  },

  // Delete holiday
  deleteHoliday: async (scheduleId: string, holidayId: string): Promise<void> => {
    await client.delete(`/schedules/${scheduleId}/holidays/${holidayId}`);
  },

  // Get suggested public holidays from external API
  getSuggestedHolidays: async (year: number): Promise<SuggestedHolidaysResponse> => {
    const response = await client.get<APIResponse<SuggestedHolidaysResponse>>(`/suggested-holidays?year=${year}`);
    if (!response.data.data) {
      throw new Error('Failed to fetch suggested holidays');
    }
    return response.data.data;
  },
};

// Schedule Exceptions API
export const exceptionsApi = {
  // Get all exceptions for a schedule
  getExceptions: async (scheduleId: string): Promise<ScheduleException[]> => {
    const response = await client.get<APIResponse<ScheduleException[]>>(`/schedules/${scheduleId}/exceptions`);
    return response.data.data || [];
  },

  // Get single exception
  getException: async (scheduleId: string, exceptionId: string): Promise<ScheduleException> => {
    const response = await client.get<APIResponse<ScheduleException>>(`/schedules/${scheduleId}/exceptions/${exceptionId}`);
    if (!response.data.data) {
      throw new Error('Schedule exception not found');
    }
    return response.data.data;
  },

  // Create exception
  createException: async (scheduleId: string, data: CreateScheduleExceptionRequest): Promise<ScheduleException> => {
    const response = await client.post<APIResponse<ScheduleException>>(`/schedules/${scheduleId}/exceptions`, data);
    if (!response.data.data) {
      throw new Error('Failed to create schedule exception');
    }
    return response.data.data;
  },

  // Update exception
  updateException: async (scheduleId: string, exceptionId: string, data: UpdateScheduleExceptionRequest): Promise<ScheduleException> => {
    const response = await client.put<APIResponse<ScheduleException>>(`/schedules/${scheduleId}/exceptions/${exceptionId}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update schedule exception');
    }
    return response.data.data;
  },

  // Delete exception
  deleteException: async (scheduleId: string, exceptionId: string): Promise<void> => {
    await client.delete(`/schedules/${scheduleId}/exceptions/${exceptionId}`);
  },
};

// Line-specific Exceptions API
export const lineExceptionsApi = {
  // Get all line-specific exceptions for a schedule
  getLineExceptions: async (scheduleId: string): Promise<LineScheduleException[]> => {
    const response = await client.get<APIResponse<LineScheduleException[]>>(`/schedules/${scheduleId}/line-exceptions`);
    return response.data.data || [];
  },

  // Get single line exception
  getLineException: async (scheduleId: string, exceptionId: string): Promise<LineScheduleException> => {
    const response = await client.get<APIResponse<LineScheduleException>>(`/schedules/${scheduleId}/line-exceptions/${exceptionId}`);
    if (!response.data.data) {
      throw new Error('Line schedule exception not found');
    }
    return response.data.data;
  },

  // Create line exception
  createLineException: async (scheduleId: string, data: CreateLineScheduleExceptionRequest): Promise<LineScheduleException> => {
    const response = await client.post<APIResponse<LineScheduleException>>(`/schedules/${scheduleId}/line-exceptions`, data);
    if (!response.data.data) {
      throw new Error('Failed to create line schedule exception');
    }
    return response.data.data;
  },

  // Update line exception
  updateLineException: async (scheduleId: string, exceptionId: string, data: UpdateLineScheduleExceptionRequest): Promise<LineScheduleException> => {
    const response = await client.put<APIResponse<LineScheduleException>>(`/schedules/${scheduleId}/line-exceptions/${exceptionId}`, data);
    if (!response.data.data) {
      throw new Error('Failed to update line schedule exception');
    }
    return response.data.data;
  },

  // Delete line exception
  deleteLineException: async (scheduleId: string, exceptionId: string): Promise<void> => {
    await client.delete(`/schedules/${scheduleId}/line-exceptions/${exceptionId}`);
  },
};

// Effective Schedule API
export const effectiveScheduleApi = {
  // Get effective schedule for a line on a specific date
  getEffectiveSchedule: async (lineId: string, date: string): Promise<EffectiveSchedule> => {
    const response = await client.get<APIResponse<EffectiveSchedule>>(`/effective-schedule?line_id=${lineId}&date=${date}`);
    if (!response.data.data) {
      throw new Error('Effective schedule not found');
    }
    return response.data.data;
  },

  // Get effective schedule range for a line
  getEffectiveScheduleRange: async (lineId: string, startDate: string, endDate: string): Promise<EffectiveSchedule[]> => {
    const response = await client.get<APIResponse<EffectiveSchedule[]>>(
      `/effective-schedule/range?line_id=${lineId}&start_date=${startDate}&end_date=${endDate}`
    );
    return response.data.data || [];
  },
};

// Line Schedule Assignment API
export const lineScheduleApi = {
  // Assign schedule to a line
  assignScheduleToLine: async (lineId: string, scheduleId: string | null): Promise<void> => {
    const request: AssignScheduleToLineRequest = { schedule_id: scheduleId };
    await client.put(`/lines/${lineId}/schedule`, request);
  },
};
