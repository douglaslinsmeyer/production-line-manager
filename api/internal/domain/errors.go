package domain

import "errors"

var (
	// ErrNotFound is returned when a production line is not found
	ErrNotFound = errors.New("production line not found")

	// ErrCodeExists is returned when a production line code already exists
	ErrCodeExists = errors.New("production line code already exists")

	// ErrInvalidStatus is returned when an invalid status value is provided
	ErrInvalidStatus = errors.New("invalid status value")

	// ErrInvalidID is returned when an invalid UUID is provided
	ErrInvalidID = errors.New("invalid id format")

	// Schedule errors
	ErrScheduleNotFound          = errors.New("schedule not found")
	ErrScheduleNameExists        = errors.New("schedule name already exists")
	ErrScheduleDayNotFound       = errors.New("schedule day not found")
	ErrScheduleBreakNotFound     = errors.New("schedule break not found")
	ErrHolidayNotFound           = errors.New("holiday not found")
	ErrHolidayDateExists         = errors.New("holiday date already exists for this schedule")
	ErrExceptionNotFound         = errors.New("schedule exception not found")
	ErrExceptionDatesOverlap     = errors.New("exception dates overlap with existing exception")
	ErrLineExceptionNotFound     = errors.New("line schedule exception not found")
	ErrLineExceptionLinesOverlap = errors.New("line exception overlaps for one or more lines")

	// Schedule validation errors
	ErrInvalidDayOfWeek       = errors.New("invalid day of week value")
	ErrInvalidTimeFormat      = errors.New("invalid time format")
	ErrInvalidDateFormat      = errors.New("invalid date format")
	ErrInvalidTimezone        = errors.New("invalid timezone")
	ErrBreakOutsideShift      = errors.New("break times must fall within shift times")
	ErrBreaksOverlap          = errors.New("breaks cannot overlap")
	ErrMissingShiftTimes      = errors.New("working days must have shift times")
	ErrUnexpectedShiftTimes   = errors.New("non-working days cannot have shift times")
	ErrInvalidDateRange       = errors.New("start date must be before or equal to end date")
	ErrMissingLinesToException = errors.New("line exception must specify at least one line")
	ErrLineNotAssignedSchedule = errors.New("line is not assigned to this schedule")
)
