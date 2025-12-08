package domain

import "github.com/google/uuid"

// ========== Schedule Requests ==========

// CreateScheduleRequest represents the request to create a new schedule
type CreateScheduleRequest struct {
	Name        string                   `json:"name" validate:"required,max=100"`
	Description *string                  `json:"description,omitempty"`
	Timezone    string                   `json:"timezone" validate:"required"`
	Days        []CreateScheduleDayInput `json:"days" validate:"required,len=7,dive"`
}

// CreateScheduleDayInput represents input for a single day in schedule creation
type CreateScheduleDayInput struct {
	DayOfWeek    DayOfWeek                  `json:"day_of_week" validate:"min=0,max=6"`
	IsWorkingDay bool                       `json:"is_working_day"`
	ShiftStart   *TimeOnly                  `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly                  `json:"shift_end,omitempty"`
	Breaks       []CreateScheduleBreakInput `json:"breaks,omitempty" validate:"dive"`
}

// CreateScheduleBreakInput represents input for a break in schedule creation
type CreateScheduleBreakInput struct {
	Name       *string  `json:"name,omitempty" validate:"omitempty,max=100"`
	BreakStart TimeOnly `json:"break_start" validate:"required"`
	BreakEnd   TimeOnly `json:"break_end" validate:"required"`
}

// UpdateScheduleRequest represents the request to update a schedule
type UpdateScheduleRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Description *string `json:"description,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
}

// UpdateScheduleDayRequest represents the request to update a single day
type UpdateScheduleDayRequest struct {
	IsWorkingDay *bool     `json:"is_working_day,omitempty"`
	ShiftStart   *TimeOnly `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly `json:"shift_end,omitempty"`
}

// SetScheduleBreaksRequest replaces all breaks for a day
type SetScheduleBreaksRequest struct {
	Breaks []CreateScheduleBreakInput `json:"breaks" validate:"dive"`
}

// ========== Holiday Requests ==========

// CreateHolidayRequest represents the request to create a holiday
type CreateHolidayRequest struct {
	HolidayDate string  `json:"holiday_date" validate:"required"`
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
}

// UpdateHolidayRequest represents the request to update a holiday
type UpdateHolidayRequest struct {
	HolidayDate *string `json:"holiday_date,omitempty"`
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
}

// ========== Schedule Exception Requests ==========

// CreateScheduleExceptionRequest represents the request to create a schedule exception
type CreateScheduleExceptionRequest struct {
	Name        string                            `json:"name" validate:"required,max=100"`
	Description *string                           `json:"description,omitempty"`
	StartDate   string                            `json:"start_date" validate:"required"`
	EndDate     string                            `json:"end_date" validate:"required"`
	Days        []CreateScheduleExceptionDayInput `json:"days" validate:"required,len=7,dive"`
}

// CreateScheduleExceptionDayInput represents input for a single day in exception creation
type CreateScheduleExceptionDayInput struct {
	DayOfWeek    DayOfWeek                           `json:"day_of_week" validate:"min=0,max=6"`
	IsWorkingDay bool                                `json:"is_working_day"`
	ShiftStart   *TimeOnly                           `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly                           `json:"shift_end,omitempty"`
	Breaks       []CreateScheduleExceptionBreakInput `json:"breaks,omitempty" validate:"dive"`
}

// CreateScheduleExceptionBreakInput represents input for a break in exception creation
type CreateScheduleExceptionBreakInput struct {
	Name       *string  `json:"name,omitempty" validate:"omitempty,max=100"`
	BreakStart TimeOnly `json:"break_start" validate:"required"`
	BreakEnd   TimeOnly `json:"break_end" validate:"required"`
}

// UpdateScheduleExceptionRequest represents the request to update a schedule exception
type UpdateScheduleExceptionRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Description *string `json:"description,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

// UpdateScheduleExceptionDayRequest represents the request to update an exception day
type UpdateScheduleExceptionDayRequest struct {
	IsWorkingDay *bool     `json:"is_working_day,omitempty"`
	ShiftStart   *TimeOnly `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly `json:"shift_end,omitempty"`
}

// SetScheduleExceptionBreaksRequest replaces all breaks for an exception day
type SetScheduleExceptionBreaksRequest struct {
	Breaks []CreateScheduleExceptionBreakInput `json:"breaks" validate:"dive"`
}

// ========== Line Schedule Exception Requests ==========

// CreateLineScheduleExceptionRequest represents the request to create a line-specific exception
type CreateLineScheduleExceptionRequest struct {
	Name        string                                `json:"name" validate:"required,max=100"`
	Description *string                               `json:"description,omitempty"`
	StartDate   string                                `json:"start_date" validate:"required"`
	EndDate     string                                `json:"end_date" validate:"required"`
	LineIDs     []uuid.UUID                           `json:"line_ids" validate:"required,min=1,dive"`
	Days        []CreateLineScheduleExceptionDayInput `json:"days" validate:"required,len=7,dive"`
}

// CreateLineScheduleExceptionDayInput represents input for a single day
type CreateLineScheduleExceptionDayInput struct {
	DayOfWeek    DayOfWeek                               `json:"day_of_week" validate:"min=0,max=6"`
	IsWorkingDay bool                                    `json:"is_working_day"`
	ShiftStart   *TimeOnly                               `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly                               `json:"shift_end,omitempty"`
	Breaks       []CreateLineScheduleExceptionBreakInput `json:"breaks,omitempty" validate:"dive"`
}

// CreateLineScheduleExceptionBreakInput represents input for a break
type CreateLineScheduleExceptionBreakInput struct {
	Name       *string  `json:"name,omitempty" validate:"omitempty,max=100"`
	BreakStart TimeOnly `json:"break_start" validate:"required"`
	BreakEnd   TimeOnly `json:"break_end" validate:"required"`
}

// UpdateLineScheduleExceptionRequest represents the request to update a line-specific exception
type UpdateLineScheduleExceptionRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,max=100"`
	Description *string     `json:"description,omitempty"`
	StartDate   *string     `json:"start_date,omitempty"`
	EndDate     *string     `json:"end_date,omitempty"`
	LineIDs     []uuid.UUID `json:"line_ids,omitempty" validate:"omitempty,min=1,dive"`
}

// UpdateLineScheduleExceptionDayRequest represents the request to update a line exception day
type UpdateLineScheduleExceptionDayRequest struct {
	IsWorkingDay *bool     `json:"is_working_day,omitempty"`
	ShiftStart   *TimeOnly `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly `json:"shift_end,omitempty"`
}

// SetLineScheduleExceptionBreaksRequest replaces all breaks for a line exception day
type SetLineScheduleExceptionBreaksRequest struct {
	Breaks []CreateLineScheduleExceptionBreakInput `json:"breaks" validate:"dive"`
}

// ========== Line Assignment Requests ==========

// AssignScheduleToLineRequest represents the request to assign a schedule to a line
type AssignScheduleToLineRequest struct {
	ScheduleID *uuid.UUID `json:"schedule_id"` // NULL to unassign
}

// ========== Query Requests ==========

// EffectiveScheduleQuery represents query parameters for effective schedule lookup
type EffectiveScheduleQuery struct {
	LineID uuid.UUID `json:"line_id" validate:"required"`
	Date   string    `json:"date" validate:"required"` // YYYY-MM-DD
}

// EffectiveScheduleRangeQuery represents query for effective schedule over a date range
type EffectiveScheduleRangeQuery struct {
	LineID    uuid.UUID `json:"line_id" validate:"required"`
	StartDate string    `json:"start_date" validate:"required"`
	EndDate   string    `json:"end_date" validate:"required"`
}
