package mqtt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"ping/production-line-api/internal/domain"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// DeviceDiscoveryHandler handles device announcement messages
type DeviceDiscoveryHandler struct {
	deviceRepo  domain.DeviceRepository
	publisher   *Publisher
	lineService LineService
	logger      *zap.Logger
}

// LineService defines the interface for line operations
type LineService interface {
	SetStatus(ctx context.Context, id uuid.UUID, status domain.Status, source string, sourceDetail interface{}) (*domain.ProductionLine, error)
}

// NewDeviceDiscoveryHandler creates a new device discovery handler
func NewDeviceDiscoveryHandler(
	deviceRepo domain.DeviceRepository,
	publisher *Publisher,
	lineService LineService,
	logger *zap.Logger,
) *DeviceDiscoveryHandler {
	return &DeviceDiscoveryHandler{
		deviceRepo:  deviceRepo,
		publisher:   publisher,
		lineService: lineService,
		logger:      logger,
	}
}

// HandleAnnouncement processes device announcement messages
func (h *DeviceDiscoveryHandler) HandleAnnouncement(client mqtt.Client, msg mqtt.Message) {
	h.logger.Info("Received device announcement",
		zap.String("topic", msg.Topic()),
		zap.Int("payload_size", len(msg.Payload())))

	// Parse announcement
	var announcement domain.DeviceAnnouncement
	if err := json.Unmarshal(msg.Payload(), &announcement); err != nil {
		h.logger.Error("Failed to parse device announcement",
			zap.Error(err),
			zap.String("payload", string(msg.Payload())))
		return
	}

	// Validate MAC address
	if announcement.MACAddress == "" {
		h.logger.Warn("Device announcement missing MAC address",
			zap.String("device_id", announcement.DeviceID))
		return
	}

	// Marshal capabilities to JSON
	capabilitiesJSON, err := json.Marshal(announcement.Capabilities)
	if err != nil {
		h.logger.Error("Failed to marshal capabilities", zap.Error(err))
		capabilitiesJSON = []byte("{}")
	}

	// Marshal status metadata to JSON
	metadataJSON, err := json.Marshal(announcement.Status)
	if err != nil {
		h.logger.Error("Failed to marshal status metadata", zap.Error(err))
		metadataJSON = []byte("{}")
	}

	// Create or update device record
	device := &domain.DiscoveredDevice{
		MACAddress:      announcement.MACAddress,
		DeviceType:      announcement.DeviceType,
		FirmwareVersion: &announcement.FirmwareVersion,
		IPAddress:       &announcement.IPAddress,
		Capabilities:    capabilitiesJSON,
		FirstSeen:       time.Now(), // Will be ignored if device exists
		LastSeen:        time.Now(),
		Status:          domain.DeviceStatusOnline,
		Metadata:        metadataJSON,
	}

	if err := h.deviceRepo.UpsertDevice(device); err != nil {
		h.logger.Error("Failed to upsert device",
			zap.Error(err),
			zap.String("mac", announcement.MACAddress))
		return
	}

	h.logger.Info("Device discovered/updated",
		zap.String("mac", announcement.MACAddress),
		zap.String("type", announcement.DeviceType),
		zap.String("firmware", announcement.FirmwareVersion),
		zap.String("ip", announcement.IPAddress))

	// TODO: Emit websocket event for real-time dashboard updates
}

// HandleDeviceStatus processes device status messages
func (h *DeviceDiscoveryHandler) HandleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	h.logger.Debug("Received device status",
		zap.String("topic", msg.Topic()))

	// Parse message
	var status struct {
		DeviceID          string  `json:"device_id"`
		LineState         string  `json:"line_state"`
		DigitalInputs     int     `json:"digital_inputs"`
		DigitalOutputs    int     `json:"digital_outputs"`
		EthernetConnected bool    `json:"ethernet_connected"`
		AssignedLine      *string `json:"assigned_line"`
		Timestamp         int64   `json:"timestamp"`
	}

	if err := json.Unmarshal(msg.Payload(), &status); err != nil {
		h.logger.Error("Failed to parse device status", zap.Error(err))
		return
	}

	// Update last_seen timestamp
	_, err := h.deviceRepo.GetDeviceByMAC(status.DeviceID)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err))
		return
	}

	// Check if device is assigned to a line
	assignment, err := h.deviceRepo.GetDeviceAssignment(status.DeviceID)
	if err != nil {
		h.logger.Error("Failed to get device assignment", zap.Error(err))
		return
	}

	if assignment == nil {
		h.logger.Debug("Device not assigned to a line - skipping status translation",
			zap.String("device_mac", status.DeviceID))
		return
	}

	// Device is assigned - translate line_state to production line status
	h.logger.Debug("Translating device status to line event",
		zap.String("device_mac", status.DeviceID),
		zap.String("line_id", assignment.LineID.String()),
		zap.String("line_state", status.LineState))

	// Map firmware line_state to domain.Status
	var lineStatus domain.Status
	switch status.LineState {
	case "ON":
		lineStatus = domain.StatusOn
	case "OFF":
		lineStatus = domain.StatusOff
	case "MAINTENANCE":
		lineStatus = domain.StatusMaintenance
	case "ERROR":
		lineStatus = domain.StatusError
	case "UNKNOWN":
		// Skip update - device not synchronized yet
		h.logger.Debug("Device line_state is UNKNOWN - skipping status update",
			zap.String("device_mac", status.DeviceID))
		return
	default:
		h.logger.Warn("Unknown line_state value from device",
			zap.String("device_mac", status.DeviceID),
			zap.String("line_state", status.LineState))
		return
	}

	// Prepare source detail for audit trail
	sourceDetail := map[string]interface{}{
		"device_mac":       status.DeviceID,
		"digital_inputs":   status.DigitalInputs,
		"digital_outputs":  status.DigitalOutputs,
		"device_timestamp": status.Timestamp,
	}

	// Update production line status via LineService
	// This will also publish to MQTT and trigger SSE broadcast
	ctx := context.Background()
	updatedLine, err := h.lineService.SetStatus(ctx, assignment.LineID, lineStatus, "device", sourceDetail)
	if err != nil {
		h.logger.Error("Failed to update line status from device",
			zap.String("device_mac", status.DeviceID),
			zap.String("line_id", assignment.LineID.String()),
			zap.String("status", lineStatus.String()),
			zap.Error(err))
		return
	}

	h.logger.Info("Line status updated from device",
		zap.String("device_mac", status.DeviceID),
		zap.String("line_id", updatedLine.ID.String()),
		zap.String("line_code", updatedLine.Code),
		zap.String("old_status", updatedLine.Status.String()),
		zap.String("new_status", lineStatus.String()))
}

// HandleInputChange processes device input change events
func (h *DeviceDiscoveryHandler) HandleInputChange(client mqtt.Client, msg mqtt.Message) {
	h.logger.Debug("Received device input change",
		zap.String("topic", msg.Topic()))

	// Parse message
	var inputChange struct {
		DeviceID  string `json:"device_id"`
		Channel   int    `json:"channel"`
		State     bool   `json:"state"`
		AllInputs int    `json:"all_inputs"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(msg.Payload(), &inputChange); err != nil {
		h.logger.Error("Failed to parse input change", zap.Error(err))
		return
	}

	// Check if device is assigned to a line
	assignment, err := h.deviceRepo.GetDeviceAssignment(inputChange.DeviceID)
	if err != nil {
		h.logger.Error("Failed to get device assignment", zap.Error(err))
		return
	}

	if assignment == nil {
		h.logger.Debug("Input change from unassigned device - ignoring",
			zap.String("device_mac", inputChange.DeviceID),
			zap.Int("channel", inputChange.Channel))
		return
	}

	h.logger.Info("Input change from assigned device",
		zap.String("device_mac", inputChange.DeviceID),
		zap.String("line_id", assignment.LineID.String()),
		zap.Int("channel", inputChange.Channel),
		zap.Bool("state", inputChange.State))

	// TODO: Translate input state to line status
	// This will depend on your business logic for interpreting inputs
	// For example: all inputs HIGH = line running, any LOW = line stopped
}

// StartStaleDeviceMonitor starts a background task to mark stale devices offline
func (h *DeviceDiscoveryHandler) StartStaleDeviceMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			threshold := 2 * time.Minute
			if err := h.deviceRepo.MarkStaleDevicesOffline(threshold); err != nil {
				h.logger.Error("Failed to mark stale devices offline", zap.Error(err))
			}
		}
	}()

	h.logger.Info("Started stale device monitor", zap.Duration("threshold", 2*time.Minute))
}
