package domain

import (
	"time"

	"github.com/google/uuid"
)

// Label represents a tag/category for production lines
type Label struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Color       *string    `json:"color,omitempty" db:"color"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateLabelRequest represents the request to create a new label
type CreateLabelRequest struct {
	Name        string  `json:"name" validate:"required,max=100"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Description *string `json:"description,omitempty"`
}

// UpdateLabelRequest represents the request to update a label
type UpdateLabelRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Description *string `json:"description,omitempty"`
}

// ProductionLineLabel represents the relationship between a line and label
type ProductionLineLabel struct {
	LineID     uuid.UUID `json:"line_id" db:"line_id"`
	LabelID    uuid.UUID `json:"label_id" db:"label_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}

// AssignLabelsRequest represents the request to assign labels to a line
type AssignLabelsRequest struct {
	LabelIDs []uuid.UUID `json:"label_ids" validate:"required,min=1,dive,uuid"`
}
