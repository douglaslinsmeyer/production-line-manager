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
)
