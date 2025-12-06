package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of MQTT event
type EventType string

const (
	EventCreated EventType = "created"
	EventUpdated EventType = "updated"
	EventDeleted EventType = "deleted"
	EventStatus  EventType = "status"
)

// LineEvent represents a production line event (created/updated)
type LineEvent struct {
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      *ProductionLine `json:"data,omitempty"`
}

// LineDeletedEvent represents a production line deletion event
type LineDeletedEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
}

// StatusEvent represents a status change event
type StatusEvent struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Status    Status    `json:"status"`
}

// StatusCommand represents a command to change production line status
// This is received via MQTT from shop floor controllers
type StatusCommand struct {
	Code   string `json:"code"`
	Status Status `json:"status"`
}
