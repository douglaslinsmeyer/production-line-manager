package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DeviceStatus represents the operational status of a device
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusUnknown DeviceStatus = "unknown"
)

// DiscoveredDevice represents an ESP32 device that has announced itself
type DiscoveredDevice struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	MACAddress      string          `json:"mac_address" db:"mac_address"`
	DeviceType      string          `json:"device_type" db:"device_type"`
	FirmwareVersion *string         `json:"firmware_version,omitempty" db:"firmware_version"`
	IPAddress       *string         `json:"ip_address,omitempty" db:"ip_address"`
	Capabilities    json.RawMessage `json:"capabilities,omitempty" db:"capabilities"`
	FirstSeen       time.Time       `json:"first_seen" db:"first_seen"`
	LastSeen        time.Time       `json:"last_seen" db:"last_seen"`
	Status          DeviceStatus    `json:"status" db:"status"`
	Metadata        json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// DeviceCapabilities represents what a device can do
type DeviceCapabilities struct {
	DigitalInputs  int  `json:"digital_inputs"`
	DigitalOutputs int  `json:"digital_outputs"`
	Ethernet       bool `json:"ethernet"`
	WiFi           bool `json:"wifi"`
}

// DeviceLineAssignment represents a device assigned to a production line
type DeviceLineAssignment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	DeviceMAC  string     `json:"device_mac" db:"device_mac"`
	LineID     uuid.UUID  `json:"line_id" db:"line_id"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	AssignedBy *string    `json:"assigned_by,omitempty" db:"assigned_by"`
	Active     bool       `json:"active" db:"active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// DeviceAnnouncement represents the MQTT message from devices
type DeviceAnnouncement struct {
	DeviceID        string                 `json:"device_id"`
	DeviceType      string                 `json:"device_type"`
	FirmwareVersion string                 `json:"firmware_version"`
	IPAddress       string                 `json:"ip_address"`
	MACAddress      string                 `json:"mac_address"`
	Capabilities    DeviceCapabilities     `json:"capabilities"`
	Status          map[string]interface{} `json:"status"`
	Timestamp       int64                  `json:"timestamp"`
}

// DeviceWithAssignment includes device info with its line assignment
type DeviceWithAssignment struct {
	DiscoveredDevice
	AssignedLineID   *uuid.UUID `json:"assigned_line_id,omitempty" db:"assigned_line_id"`
	AssignedLineCode *string    `json:"assigned_line_code,omitempty" db:"assigned_line_code"`
	AssignedLineName *string    `json:"assigned_line_name,omitempty" db:"assigned_line_name"`
	AssignedAt       *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`
}

// DeviceCommand represents a command to send to a device
type DeviceCommand struct {
	Command  string                 `json:"command"`
	Channel  *int                   `json:"channel,omitempty"`
	State    *bool                  `json:"state,omitempty"`
	Duration *int                   `json:"duration,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// DeviceRepository defines the interface for device data access
type DeviceRepository interface {
	// Device discovery and registration
	UpsertDevice(device *DiscoveredDevice) error
	GetDeviceByMAC(macAddress string) (*DiscoveredDevice, error)
	DeleteDevice(macAddress string) error
	ListDevices() ([]*DeviceWithAssignment, error)
	UpdateDeviceStatus(macAddress string, status DeviceStatus) error
	MarkStaleDevicesOffline(threshold time.Duration) error

	// Device-Line assignments
	AssignDeviceToLine(deviceMAC string, lineID uuid.UUID, assignedBy *string) error
	UnassignDevice(deviceMAC string) error
	GetDeviceAssignment(deviceMAC string) (*DeviceLineAssignment, error)
	GetLineAssignment(lineID uuid.UUID) (*DeviceLineAssignment, error)
	ListAssignments() ([]*DeviceLineAssignment, error)
}
