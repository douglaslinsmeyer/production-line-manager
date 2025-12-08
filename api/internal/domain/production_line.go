package domain

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the operational status of a production line
type Status string

const (
	StatusOn          Status = "on"
	StatusOff         Status = "off"
	StatusMaintenance Status = "maintenance"
	StatusError       Status = "error"
)

// IsValid checks if the status value is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusOn, StatusOff, StatusMaintenance, StatusError:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// ProductionLine represents a production line in the facility
type ProductionLine struct {
	ID          uuid.UUID         `json:"id" db:"id"`
	Code        string            `json:"code" db:"code"`
	Name        string            `json:"name" db:"name"`
	Description *string           `json:"description,omitempty" db:"description"`
	Status      Status            `json:"status" db:"status"`
	ScheduleID  *uuid.UUID        `json:"schedule_id,omitempty" db:"schedule_id"`
	Labels      []Label           `json:"labels,omitempty" db:"-"`           // Loaded via JOIN, not from production_lines table
	Schedule    *ScheduleSummary  `json:"schedule,omitempty" db:"-"`         // Loaded via JOIN when needed
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateLineRequest represents the request to create a new production line
type CreateLineRequest struct {
	Code        string  `json:"code" validate:"required,max=50"`
	Name        string  `json:"name" validate:"required,max=255"`
	Description *string `json:"description,omitempty"`
}

// UpdateLineRequest represents the request to update a production line
type UpdateLineRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty"`
}

// SetStatusRequest represents the request to set a production line's status
type SetStatusRequest struct {
	Status Status `json:"status" validate:"required,oneof=on off maintenance error"`
}

// StatusChange represents a status change event in the audit log
type StatusChange struct {
	Time         time.Time   `json:"time" db:"time"`
	LineID       uuid.UUID   `json:"line_id" db:"line_id"`
	LineCode     string      `json:"line_code" db:"line_code"`
	OldStatus    *Status     `json:"old_status,omitempty" db:"old_status"`
	NewStatus    Status      `json:"new_status" db:"new_status"`
	Source       string      `json:"source" db:"source"`
	SourceDetail interface{} `json:"source_detail,omitempty" db:"source_detail"`
}
