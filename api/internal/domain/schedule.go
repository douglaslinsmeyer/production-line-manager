package domain

import (
	"time"

	"github.com/google/uuid"
)

// DayOfWeek represents days of the week (0=Sunday, 6=Saturday)
type DayOfWeek int

const (
	Sunday    DayOfWeek = 0
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
)

// IsValid checks if the day of week value is valid
func (d DayOfWeek) IsValid() bool {
	return d >= Sunday && d <= Saturday
}

// String returns the string representation of the day
func (d DayOfWeek) String() string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if d.IsValid() {
		return days[d]
	}
	return "Unknown"
}

// TimeOnly represents a time without date (HH:MM:SS format)
type TimeOnly string

// Hour returns the hour component of the time
func (t TimeOnly) Hour() int {
	parsed, err := time.Parse("15:04:05", string(t))
	if err != nil {
		return 0
	}
	return parsed.Hour()
}

// Minute returns the minute component of the time
func (t TimeOnly) Minute() int {
	parsed, err := time.Parse("15:04:05", string(t))
	if err != nil {
		return 0
	}
	return parsed.Minute()
}

// Second returns the second component of the time
func (t TimeOnly) Second() int {
	parsed, err := time.Parse("15:04:05", string(t))
	if err != nil {
		return 0
	}
	return parsed.Second()
}

// ToTime converts TimeOnly to a time.Time on a reference date
func (t TimeOnly) ToTime(referenceDate time.Time) time.Time {
	return time.Date(
		referenceDate.Year(), referenceDate.Month(), referenceDate.Day(),
		t.Hour(), t.Minute(), t.Second(), 0, referenceDate.Location(),
	)
}

// Schedule represents a production schedule
type Schedule struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	Description *string        `json:"description,omitempty" db:"description"`
	Timezone    string         `json:"timezone" db:"timezone"`
	Days        []ScheduleDay  `json:"days,omitempty" db:"-"`
	Holidays    []ScheduleHoliday `json:"holidays,omitempty" db:"-"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ScheduleDay represents a single day's shift in a schedule
type ScheduleDay struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	ScheduleID   uuid.UUID       `json:"schedule_id" db:"schedule_id"`
	DayOfWeek    DayOfWeek       `json:"day_of_week" db:"day_of_week"`
	IsWorkingDay bool            `json:"is_working_day" db:"is_working_day"`
	ShiftStart   *TimeOnly       `json:"shift_start,omitempty" db:"shift_start"`
	ShiftEnd     *TimeOnly       `json:"shift_end,omitempty" db:"shift_end"`
	Breaks       []ScheduleBreak `json:"breaks,omitempty" db:"-"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// ScheduleBreak represents a break within a shift
type ScheduleBreak struct {
	ID            uuid.UUID `json:"id" db:"id"`
	ScheduleDayID uuid.UUID `json:"schedule_day_id" db:"schedule_day_id"`
	Name          *string   `json:"name,omitempty" db:"name"`
	BreakStart    TimeOnly  `json:"break_start" db:"break_start"`
	BreakEnd      TimeOnly  `json:"break_end" db:"break_end"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ScheduleHoliday represents a date when no production occurs
type ScheduleHoliday struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ScheduleID  uuid.UUID `json:"schedule_id" db:"schedule_id"`
	HolidayDate string    `json:"holiday_date" db:"holiday_date"` // YYYY-MM-DD format
	Name        *string   `json:"name,omitempty" db:"name"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ScheduleException represents a date range with different shift times for ALL lines
type ScheduleException struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	ScheduleID  uuid.UUID              `json:"schedule_id" db:"schedule_id"`
	Name        string                 `json:"name" db:"name"`
	Description *string                `json:"description,omitempty" db:"description"`
	StartDate   string                 `json:"start_date" db:"start_date"` // YYYY-MM-DD
	EndDate     string                 `json:"end_date" db:"end_date"`     // YYYY-MM-DD
	Days        []ScheduleExceptionDay `json:"days,omitempty" db:"-"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ScheduleExceptionDay represents a single day's shift within an exception period
type ScheduleExceptionDay struct {
	ID           uuid.UUID                `json:"id" db:"id"`
	ExceptionID  uuid.UUID                `json:"exception_id" db:"exception_id"`
	DayOfWeek    DayOfWeek                `json:"day_of_week" db:"day_of_week"`
	IsWorkingDay bool                     `json:"is_working_day" db:"is_working_day"`
	ShiftStart   *TimeOnly                `json:"shift_start,omitempty" db:"shift_start"`
	ShiftEnd     *TimeOnly                `json:"shift_end,omitempty" db:"shift_end"`
	Breaks       []ScheduleExceptionBreak `json:"breaks,omitempty" db:"-"`
	CreatedAt    time.Time                `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" db:"updated_at"`
}

// ScheduleExceptionBreak represents a break within an exception day's shift
type ScheduleExceptionBreak struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ExceptionDayID uuid.UUID `json:"exception_day_id" db:"exception_day_id"`
	Name           *string   `json:"name,omitempty" db:"name"`
	BreakStart     TimeOnly  `json:"break_start" db:"break_start"`
	BreakEnd       TimeOnly  `json:"break_end" db:"break_end"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// LineScheduleException represents a date range with different shift times for SPECIFIC lines
type LineScheduleException struct {
	ID          uuid.UUID                  `json:"id" db:"id"`
	ScheduleID  uuid.UUID                  `json:"schedule_id" db:"schedule_id"`
	Name        string                     `json:"name" db:"name"`
	Description *string                    `json:"description,omitempty" db:"description"`
	StartDate   string                     `json:"start_date" db:"start_date"` // YYYY-MM-DD
	EndDate     string                     `json:"end_date" db:"end_date"`     // YYYY-MM-DD
	LineIDs     []uuid.UUID                `json:"line_ids" db:"-"`            // Lines this exception applies to
	Days        []LineScheduleExceptionDay `json:"days,omitempty" db:"-"`
	CreatedAt   time.Time                  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at" db:"updated_at"`
}

// LineScheduleExceptionDay represents a single day's shift for a line-specific exception
type LineScheduleExceptionDay struct {
	ID           uuid.UUID                    `json:"id" db:"id"`
	ExceptionID  uuid.UUID                    `json:"exception_id" db:"exception_id"`
	DayOfWeek    DayOfWeek                    `json:"day_of_week" db:"day_of_week"`
	IsWorkingDay bool                         `json:"is_working_day" db:"is_working_day"`
	ShiftStart   *TimeOnly                    `json:"shift_start,omitempty" db:"shift_start"`
	ShiftEnd     *TimeOnly                    `json:"shift_end,omitempty" db:"shift_end"`
	Breaks       []LineScheduleExceptionBreak `json:"breaks,omitempty" db:"-"`
	CreatedAt    time.Time                    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time                    `json:"updated_at" db:"updated_at"`
}

// LineScheduleExceptionBreak represents a break within a line-specific exception day's shift
type LineScheduleExceptionBreak struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ExceptionDayID uuid.UUID `json:"exception_day_id" db:"exception_day_id"`
	Name           *string   `json:"name,omitempty" db:"name"`
	BreakStart     TimeOnly  `json:"break_start" db:"break_start"`
	BreakEnd       TimeOnly  `json:"break_end" db:"break_end"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ScheduleSource indicates where the effective schedule came from
type ScheduleSource string

const (
	SourceNoSchedule        ScheduleSource = "no_schedule"
	SourceBase              ScheduleSource = "base"
	SourceHoliday           ScheduleSource = "holiday"
	SourceScheduleException ScheduleSource = "schedule_exception"
	SourceLineException     ScheduleSource = "line_exception"
)

// EffectiveBreak represents a break in the effective schedule
type EffectiveBreak struct {
	Name       *string  `json:"name,omitempty"`
	BreakStart TimeOnly `json:"break_start"`
	BreakEnd   TimeOnly `json:"break_end"`
}

// EffectiveSchedule represents the computed schedule for a line on a specific date
type EffectiveSchedule struct {
	LineID       uuid.UUID        `json:"line_id"`
	LineCode     string           `json:"line_code"`
	Date         string           `json:"date"` // YYYY-MM-DD
	ScheduleID   *uuid.UUID       `json:"schedule_id,omitempty"`
	ScheduleName *string          `json:"schedule_name,omitempty"`
	Source       ScheduleSource   `json:"source"`
	SourceID     *uuid.UUID       `json:"source_id,omitempty"`
	SourceName   *string          `json:"source_name,omitempty"`
	IsWorkingDay bool             `json:"is_working_day"`
	ShiftStart   *TimeOnly        `json:"shift_start,omitempty"`
	ShiftEnd     *TimeOnly        `json:"shift_end,omitempty"`
	Breaks       []EffectiveBreak `json:"breaks,omitempty"`
}

// ScheduleSummary provides a summary view of a schedule for list endpoints
type ScheduleSummary struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Timezone    string    `json:"timezone" db:"timezone"`
	LineCount   int       `json:"line_count" db:"line_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
