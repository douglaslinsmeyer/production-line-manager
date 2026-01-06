package domain

import (
	"time"

	"github.com/google/uuid"
)

// LightMode represents the mode for a tower light
type LightMode string

const (
	LightOff        LightMode = "off"
	LightOn         LightMode = "on"
	LightShortBlink LightMode = "shortBlink"
	LightLongBlink  LightMode = "longBlink"
)

// IsValid checks if the light mode value is valid
func (l LightMode) IsValid() bool {
	switch l {
	case LightOff, LightOn, LightShortBlink, LightLongBlink:
		return true
	default:
		return false
	}
}

// String returns the string representation of the light mode
func (l LightMode) String() string {
	return string(l)
}

// BuzzerMode represents the mode for the buzzer
type BuzzerMode string

const (
	BuzzerOff   BuzzerMode = "off"
	BuzzerOn    BuzzerMode = "on"
	BuzzerChirp BuzzerMode = "chirp"
)

// IsValid checks if the buzzer mode value is valid
func (b BuzzerMode) IsValid() bool {
	switch b {
	case BuzzerOff, BuzzerOn, BuzzerChirp:
		return true
	default:
		return false
	}
}

// String returns the string representation of the buzzer mode
func (b BuzzerMode) String() string {
	return string(b)
}

// ProfileStateOutputs defines the output configuration for a state
type ProfileStateOutputs struct {
	RedLight    LightMode  `json:"redLight" validate:"required,oneof=off on shortBlink longBlink"`
	YellowLight LightMode  `json:"yellowLight" validate:"required,oneof=off on shortBlink longBlink"`
	GreenLight  LightMode  `json:"greenLight" validate:"required,oneof=off on shortBlink longBlink"`
	Buzzer      BuzzerMode `json:"buzzer" validate:"required,oneof=off on chirp"`
}

// ProfileState represents a single state in a profile
type ProfileState struct {
	Name    string              `json:"name" validate:"required,min=1,max=50"`
	Outputs ProfileStateOutputs `json:"outputs" validate:"required"`
}

// ButtonBehavior defines how button presses cycle through states
type ButtonBehavior struct {
	ShortPressCycle []string `json:"shortPressCycle" validate:"required,min=1,dive,required"`
	LongPressCycle  []string `json:"longPressCycle" validate:"required,min=1,dive,required"`
}

// SignalProfile represents a signal profile configuration
type SignalProfile struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	Name           string         `json:"name" db:"name"`
	Description    *string        `json:"description,omitempty" db:"description"`
	Version        int            `json:"version" db:"version"`
	States         []ProfileState `json:"states" db:"states"`
	ButtonBehavior ButtonBehavior `json:"buttonBehavior" db:"button_behavior"`
	DefaultState   string         `json:"defaultState" db:"default_state"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	UpdatedBy      *string        `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ProfileVersion represents a version history entry
type ProfileVersion struct {
	ID                uuid.UUID      `json:"id" db:"id"`
	ProfileID         uuid.UUID      `json:"profile_id" db:"profile_id"`
	Version           int            `json:"version" db:"version"`
	Config            *SignalProfile `json:"config" db:"config"`
	ChangedBy         string         `json:"changed_by" db:"changed_by"`
	ChangeDescription *string        `json:"change_description,omitempty" db:"change_description"`
	Changes           []string       `json:"changes" db:"changes"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
}

// DeviceSignalState tracks a device's current state and profile version
type DeviceSignalState struct {
	DeviceMAC        string     `json:"device_mac" db:"device_mac"`
	ProfileID        *uuid.UUID `json:"profile_id,omitempty" db:"profile_id"`
	ProfileVersion   int        `json:"profile_version" db:"profile_version"`
	CurrentState     string     `json:"current_state" db:"current_state"`
	IsOverridden     bool       `json:"is_overridden" db:"is_overridden"`
	ProfileHash      *string    `json:"profile_hash,omitempty" db:"profile_hash"`
	LastStateChange  time.Time  `json:"last_state_change" db:"last_state_change"`
	LastSync         time.Time  `json:"last_sync" db:"last_sync"`
	LastVersionCheck time.Time  `json:"last_version_check" db:"last_version_check"`
	VersionStatus    string     `json:"version_status" db:"version_status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// VersionStatus represents the sync status of a device
type VersionStatus string

const (
	VersionStatusUpToDate      VersionStatus = "up-to-date"
	VersionStatusUpdatePending VersionStatus = "update-pending"
	VersionStatusFailed        VersionStatus = "failed"
	VersionStatusOffline       VersionStatus = "offline"
	VersionStatusUnknown       VersionStatus = "unknown"
)

// Request/Response types

// CreateProfileRequest represents the request to create a new profile
type CreateProfileRequest struct {
	Name           string         `json:"name" validate:"required,max=100"`
	Description    *string        `json:"description,omitempty"`
	States         []ProfileState `json:"states" validate:"required,min=1,dive"`
	ButtonBehavior ButtonBehavior `json:"buttonBehavior" validate:"required"`
	DefaultState   string         `json:"defaultState" validate:"required"`
}

// UpdateProfileRequest represents the request to update a profile
type UpdateProfileRequest struct {
	Name           *string         `json:"name,omitempty" validate:"omitempty,max=100"`
	Description    *string         `json:"description,omitempty"`
	States         *[]ProfileState `json:"states,omitempty" validate:"omitempty,min=1,dive"`
	ButtonBehavior *ButtonBehavior `json:"buttonBehavior,omitempty"`
	DefaultState   *string         `json:"defaultState,omitempty"`
}

// AssignProfileToLineRequest represents the request to assign a profile to a line
type AssignProfileToLineRequest struct {
	ProfileID uuid.UUID `json:"profile_id" validate:"required"`
}

// DeviceHeartbeatRequest represents a device heartbeat with profile sync info
type DeviceHeartbeatRequest struct {
	ProfileID      *uuid.UUID `json:"profileId,omitempty"`
	ProfileVersion int        `json:"profileVersion"`
	CurrentState   string     `json:"currentState" validate:"required"`
	IsOverridden   bool       `json:"isOverridden"`
}

// DeviceHeartbeatResponse represents the response to a device heartbeat
type DeviceHeartbeatResponse struct {
	LatestVersion   int            `json:"latestVersion"`
	UpdateAvailable bool           `json:"updateAvailable"`
	Profile         *SignalProfile `json:"profile,omitempty"`
}

// ProfileUpdatedRequest represents confirmation that device updated profile
type ProfileUpdatedRequest struct {
	ProfileID       uuid.UUID `json:"profileId" validate:"required"`
	NewVersion      int       `json:"newVersion" validate:"required"`
	PreviousVersion int       `json:"previousVersion" validate:"required"`
	CurrentState    string    `json:"currentState" validate:"required"`
	StateChanged    bool      `json:"stateChanged"`
}

// SetDeviceStateRequest represents a manual state change request
type SetDeviceStateRequest struct {
	State string `json:"state" validate:"required"`
}

// RollbackProfileRequest represents a request to rollback to a previous version
type RollbackProfileRequest struct {
	TargetVersion int    `json:"targetVersion" validate:"required,gt=0"`
	Reason        string `json:"reason" validate:"required"`
}

// DeviceVersionStatus represents the version status of a single device
type DeviceVersionStatus struct {
	DeviceMAC     string     `json:"device_mac"`
	DeviceID      string     `json:"device_id"`
	Version       int        `json:"version"`
	Status        string     `json:"status"`
	LastCheck     time.Time  `json:"last_check"`
	AssignedLine  *uuid.UUID `json:"assigned_line,omitempty"`
}

// ProfileDeviceStatusResponse represents device sync status for a profile
type ProfileDeviceStatusResponse struct {
	ProfileID      uuid.UUID             `json:"profile_id"`
	CurrentVersion int                   `json:"current_version"`
	Devices        []DeviceVersionStatus `json:"devices"`
	Summary        struct {
		UpToDate      int `json:"up_to_date"`
		UpdatePending int `json:"update_pending"`
		Failed        int `json:"failed"`
		Offline       int `json:"offline"`
	} `json:"summary"`
}

// BulkResetOverrideRequest represents a request to reset overrides for multiple devices
type BulkResetOverrideRequest struct {
	DeviceMACs []string `json:"device_macs" validate:"required,min=1,dive,required"`
}
