package handler

import (
	"encoding/json"
	"net/http"
)

// Response represents the standard API response envelope
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
	Meta  *Meta       `json:"meta,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta represents metadata for list responses
type Meta struct {
	Total int `json:"total"`
}

// Error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeInvalidID    = "INVALID_ID"
)

// writeJSON writes a success response with data
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{Data: data})
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// writeList writes a list response with metadata
func writeList(w http.ResponseWriter, data interface{}, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Data: data,
		Meta: &Meta{Total: total},
	})
}
