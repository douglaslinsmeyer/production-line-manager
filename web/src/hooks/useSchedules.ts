import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  schedulesApi,
  holidaysApi,
  exceptionsApi,
  lineExceptionsApi,
  effectiveScheduleApi,
  lineScheduleApi,
} from '@/api/schedules';
import type {
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
} from '@/api/types';
import { lineKeys } from './useLines';

// Query keys for cache management
export const scheduleKeys = {
  all: ['schedules'] as const,
  lists: () => [...scheduleKeys.all, 'list'] as const,
  list: () => [...scheduleKeys.lists()] as const,
  details: () => [...scheduleKeys.all, 'detail'] as const,
  detail: (id: string) => [...scheduleKeys.details(), id] as const,
  days: (scheduleId: string) => [...scheduleKeys.detail(scheduleId), 'days'] as const,
  day: (scheduleId: string, dayId: string) => [...scheduleKeys.days(scheduleId), dayId] as const,
  holidays: (scheduleId: string) => [...scheduleKeys.detail(scheduleId), 'holidays'] as const,
  holiday: (scheduleId: string, holidayId: string) => [...scheduleKeys.holidays(scheduleId), holidayId] as const,
  exceptions: (scheduleId: string) => [...scheduleKeys.detail(scheduleId), 'exceptions'] as const,
  exception: (scheduleId: string, exceptionId: string) => [...scheduleKeys.exceptions(scheduleId), exceptionId] as const,
  lineExceptions: (scheduleId: string) => [...scheduleKeys.detail(scheduleId), 'lineExceptions'] as const,
  lineException: (scheduleId: string, exceptionId: string) => [...scheduleKeys.lineExceptions(scheduleId), exceptionId] as const,
  scheduleLines: (scheduleId: string) => [...scheduleKeys.detail(scheduleId), 'lines'] as const,
  effective: () => ['effective-schedule'] as const,
  effectiveSingle: (lineId: string, date: string) => [...scheduleKeys.effective(), lineId, date] as const,
  effectiveRange: (lineId: string, startDate: string, endDate: string) => [...scheduleKeys.effective(), 'range', lineId, startDate, endDate] as const,
};

// ========== Schedule CRUD ==========

// Get all schedules
export function useSchedules() {
  return useQuery({
    queryKey: scheduleKeys.list(),
    queryFn: schedulesApi.getSchedules,
    staleTime: 30000, // Schedules don't change frequently
  });
}

// Get single schedule (full details)
export function useSchedule(id: string) {
  return useQuery({
    queryKey: scheduleKeys.detail(id),
    queryFn: () => schedulesApi.getSchedule(id),
    enabled: !!id,
  });
}

// Create schedule mutation
export function useCreateSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateScheduleRequest) => schedulesApi.createSchedule(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.list() });
    },
  });
}

// Update schedule mutation
export function useUpdateSchedule(id: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateScheduleRequest) => schedulesApi.updateSchedule(id, data),
    onSuccess: (updatedSchedule) => {
      queryClient.setQueryData(scheduleKeys.detail(id), updatedSchedule);
      queryClient.invalidateQueries({ queryKey: scheduleKeys.list() });
    },
  });
}

// Delete schedule mutation
export function useDeleteSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => schedulesApi.deleteSchedule(id),
    onSuccess: (_data, id) => {
      queryClient.removeQueries({ queryKey: scheduleKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.list() });
      // Also invalidate lines since they may have had this schedule
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
    },
  });
}

// ========== Schedule Day Operations ==========

// Get schedule day
export function useScheduleDay(scheduleId: string, dayId: string) {
  return useQuery({
    queryKey: scheduleKeys.day(scheduleId, dayId),
    queryFn: () => schedulesApi.getDay(scheduleId, dayId),
    enabled: !!scheduleId && !!dayId,
  });
}

// Update schedule day mutation
export function useUpdateScheduleDay(scheduleId: string, dayId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateScheduleDayRequest) => schedulesApi.updateDay(scheduleId, dayId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.detail(scheduleId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.day(scheduleId, dayId) });
    },
  });
}

// Set day breaks mutation
export function useSetDayBreaks(scheduleId: string, dayId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: SetScheduleBreaksRequest) => schedulesApi.setDayBreaks(scheduleId, dayId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.detail(scheduleId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.day(scheduleId, dayId) });
    },
  });
}

// ========== Holiday Operations ==========

// Get all holidays for a schedule
export function useHolidays(scheduleId: string) {
  return useQuery({
    queryKey: scheduleKeys.holidays(scheduleId),
    queryFn: () => holidaysApi.getHolidays(scheduleId),
    enabled: !!scheduleId,
  });
}

// Create holiday mutation
export function useCreateHoliday(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateHolidayRequest) => holidaysApi.createHoliday(scheduleId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.holidays(scheduleId) });
    },
  });
}

// Update holiday mutation
export function useUpdateHoliday(scheduleId: string, holidayId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateHolidayRequest) => holidaysApi.updateHoliday(scheduleId, holidayId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.holidays(scheduleId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.holiday(scheduleId, holidayId) });
    },
  });
}

// Delete holiday mutation
export function useDeleteHoliday(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (holidayId: string) => holidaysApi.deleteHoliday(scheduleId, holidayId),
    onSuccess: (_data, holidayId) => {
      queryClient.removeQueries({ queryKey: scheduleKeys.holiday(scheduleId, holidayId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.holidays(scheduleId) });
    },
  });
}

// ========== Schedule Exception Operations ==========

// Get all exceptions for a schedule
export function useExceptions(scheduleId: string) {
  return useQuery({
    queryKey: scheduleKeys.exceptions(scheduleId),
    queryFn: () => exceptionsApi.getExceptions(scheduleId),
    enabled: !!scheduleId,
  });
}

// Get single exception
export function useException(scheduleId: string, exceptionId: string) {
  return useQuery({
    queryKey: scheduleKeys.exception(scheduleId, exceptionId),
    queryFn: () => exceptionsApi.getException(scheduleId, exceptionId),
    enabled: !!scheduleId && !!exceptionId,
  });
}

// Create exception mutation
export function useCreateException(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateScheduleExceptionRequest) => exceptionsApi.createException(scheduleId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.exceptions(scheduleId) });
    },
  });
}

// Update exception mutation
export function useUpdateException(scheduleId: string, exceptionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateScheduleExceptionRequest) => exceptionsApi.updateException(scheduleId, exceptionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.exceptions(scheduleId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.exception(scheduleId, exceptionId) });
    },
  });
}

// Delete exception mutation
export function useDeleteException(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (exceptionId: string) => exceptionsApi.deleteException(scheduleId, exceptionId),
    onSuccess: (_data, exceptionId) => {
      queryClient.removeQueries({ queryKey: scheduleKeys.exception(scheduleId, exceptionId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.exceptions(scheduleId) });
    },
  });
}

// ========== Line-Specific Exception Operations ==========

// Get all line exceptions for a schedule
export function useLineExceptions(scheduleId: string) {
  return useQuery({
    queryKey: scheduleKeys.lineExceptions(scheduleId),
    queryFn: () => lineExceptionsApi.getLineExceptions(scheduleId),
    enabled: !!scheduleId,
  });
}

// Get single line exception
export function useLineException(scheduleId: string, exceptionId: string) {
  return useQuery({
    queryKey: scheduleKeys.lineException(scheduleId, exceptionId),
    queryFn: () => lineExceptionsApi.getLineException(scheduleId, exceptionId),
    enabled: !!scheduleId && !!exceptionId,
  });
}

// Create line exception mutation
export function useCreateLineException(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateLineScheduleExceptionRequest) => lineExceptionsApi.createLineException(scheduleId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.lineExceptions(scheduleId) });
    },
  });
}

// Update line exception mutation
export function useUpdateLineException(scheduleId: string, exceptionId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: UpdateLineScheduleExceptionRequest) => lineExceptionsApi.updateLineException(scheduleId, exceptionId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: scheduleKeys.lineExceptions(scheduleId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.lineException(scheduleId, exceptionId) });
    },
  });
}

// Delete line exception mutation
export function useDeleteLineException(scheduleId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (exceptionId: string) => lineExceptionsApi.deleteLineException(scheduleId, exceptionId),
    onSuccess: (_data, exceptionId) => {
      queryClient.removeQueries({ queryKey: scheduleKeys.lineException(scheduleId, exceptionId) });
      queryClient.invalidateQueries({ queryKey: scheduleKeys.lineExceptions(scheduleId) });
    },
  });
}

// ========== Schedule Lines ==========

// Get lines using a schedule
export function useScheduleLines(scheduleId: string) {
  return useQuery({
    queryKey: scheduleKeys.scheduleLines(scheduleId),
    queryFn: () => schedulesApi.getLinesForSchedule(scheduleId),
    enabled: !!scheduleId,
  });
}

// ========== Effective Schedule ==========

// Get effective schedule for a line on a date
export function useEffectiveSchedule(lineId: string, date: string) {
  return useQuery({
    queryKey: scheduleKeys.effectiveSingle(lineId, date),
    queryFn: () => effectiveScheduleApi.getEffectiveSchedule(lineId, date),
    enabled: !!lineId && !!date,
  });
}

// Get effective schedule range
export function useEffectiveScheduleRange(lineId: string, startDate: string, endDate: string) {
  return useQuery({
    queryKey: scheduleKeys.effectiveRange(lineId, startDate, endDate),
    queryFn: () => effectiveScheduleApi.getEffectiveScheduleRange(lineId, startDate, endDate),
    enabled: !!lineId && !!startDate && !!endDate,
  });
}

// ========== Line Schedule Assignment ==========

// Assign schedule to line mutation
export function useAssignScheduleToLine() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ lineId, scheduleId }: { lineId: string; scheduleId: string | null }) =>
      lineScheduleApi.assignScheduleToLine(lineId, scheduleId),
    onSuccess: (_data, { lineId }) => {
      // Invalidate line details
      queryClient.invalidateQueries({ queryKey: lineKeys.detail(lineId) });
      // Invalidate lines list
      queryClient.invalidateQueries({ queryKey: lineKeys.list() });
      // Invalidate schedule lines (since assignment changed)
      queryClient.invalidateQueries({ queryKey: scheduleKeys.all });
    },
  });
}
