package handler

import (
	"encoding/json"
	"net/http"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/mqtt"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	deviceRepo domain.DeviceRepository
	publisher  *mqtt.Publisher
	logger     *zap.Logger
}

func NewDeviceHandler(
	deviceRepo domain.DeviceRepository,
	publisher *mqtt.Publisher,
	logger *zap.Logger,
) *DeviceHandler {
	return &DeviceHandler{
		deviceRepo: deviceRepo,
		publisher:  publisher,
		logger:     logger,
	}
}

// ListDevices returns all discovered devices
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.deviceRepo.ListDevices()
	if err != nil {
		h.logger.Error("Failed to list devices", zap.Error(err))
		http.Error(w, "Failed to retrieve devices", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

// GetDevice returns a specific device by MAC address
func (h *DeviceHandler) GetDevice(w http.ResponseWriter, r *http.Request) {
	macAddress := chi.URLParam(r, "mac")

	device, err := h.deviceRepo.GetDeviceByMAC(macAddress)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err), zap.String("mac", macAddress))
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(device)
}

// AssignDevice assigns a device to a production line
func (h *DeviceHandler) AssignDevice(w http.ResponseWriter, r *http.Request) {
	macAddress := chi.URLParam(r, "mac")

	var req struct {
		LineID     uuid.UUID `json:"line_id"`
		AssignedBy *string   `json:"assigned_by,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if device exists
	device, err := h.deviceRepo.GetDeviceByMAC(macAddress)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err))
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Assign device to line
	if err := h.deviceRepo.AssignDeviceToLine(macAddress, req.LineID, req.AssignedBy); err != nil {
		h.logger.Error("Failed to assign device to line",
			zap.Error(err),
			zap.String("mac", macAddress),
			zap.String("line_id", req.LineID.String()))
		http.Error(w, "Failed to assign device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Device assigned to line",
		zap.String("mac", macAddress),
		zap.String("line_id", req.LineID.String()))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Device assigned successfully",
	})
}

// UnassignDevice removes a device's line assignment
func (h *DeviceHandler) UnassignDevice(w http.ResponseWriter, r *http.Request) {
	macAddress := chi.URLParam(r, "mac")

	if err := h.deviceRepo.UnassignDevice(macAddress); err != nil {
		h.logger.Error("Failed to unassign device",
			zap.Error(err),
			zap.String("mac", macAddress))
		http.Error(w, "Failed to unassign device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Device unassigned", zap.String("mac", macAddress))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Device unassigned successfully",
	})
}

// IdentifyDevice sends a flash command to the device
func (h *DeviceHandler) IdentifyDevice(w http.ResponseWriter, r *http.Request) {
	macAddress := chi.URLParam(r, "mac")

	var req struct {
		Duration int `json:"duration"`
	}

	// Default to 10 seconds
	req.Duration = 10

	if r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Ignore parse errors, use default
		}
	}

	// Check if device exists
	device, err := h.deviceRepo.GetDeviceByMAC(macAddress)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err))
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Send flash command to device
	command := domain.DeviceCommand{
		Command:  "flash_identify",
		Duration: &req.Duration,
	}

	topic := "devices/" + macAddress + "/command"
	payload, err := json.Marshal(command)
	if err != nil {
		h.logger.Error("Failed to marshal command", zap.Error(err))
		http.Error(w, "Failed to create command", http.StatusInternalServerError)
		return
	}

	if err := h.publisher.PublishRaw(topic, payload); err != nil {
		h.logger.Error("Failed to publish flash command",
			zap.Error(err),
			zap.String("mac", macAddress))
		http.Error(w, "Failed to send command to device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Flash identify command sent",
		zap.String("mac", macAddress),
		zap.Int("duration", req.Duration))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Flash command sent to device",
	})
}

// SendCommand sends a custom command to a device
func (h *DeviceHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	macAddress := chi.URLParam(r, "mac")

	var command domain.DeviceCommand
	if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate device exists
	device, err := h.deviceRepo.GetDeviceByMAC(macAddress)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err))
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}

	if device == nil {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Send command to device
	topic := "devices/" + macAddress + "/command"
	payload, err := json.Marshal(command)
	if err != nil {
		h.logger.Error("Failed to marshal command", zap.Error(err))
		http.Error(w, "Failed to create command", http.StatusInternalServerError)
		return
	}

	if err := h.publisher.PublishRaw(topic, payload); err != nil {
		h.logger.Error("Failed to publish command",
			zap.Error(err),
			zap.String("mac", macAddress),
			zap.String("command", command.Command))
		http.Error(w, "Failed to send command to device", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Command sent to device",
		zap.String("mac", macAddress),
		zap.String("command", command.Command))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Command sent to device",
	})
}

// GetLineDevice returns the device assigned to a specific production line
func (h *DeviceHandler) GetLineDevice(w http.ResponseWriter, r *http.Request) {
	lineIDStr := chi.URLParam(r, "id")

	lineID, err := uuid.Parse(lineIDStr)
	if err != nil {
		http.Error(w, "Invalid line ID", http.StatusBadRequest)
		return
	}

	assignment, err := h.deviceRepo.GetLineAssignment(lineID)
	if err != nil {
		h.logger.Error("Failed to get line assignment", zap.Error(err))
		http.Error(w, "Failed to retrieve assignment", http.StatusInternalServerError)
		return
	}

	if assignment == nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"assigned": false,
			"device":   nil,
		})
		return
	}

	// Get full device details
	device, err := h.deviceRepo.GetDeviceByMAC(assignment.DeviceMAC)
	if err != nil {
		h.logger.Error("Failed to get device", zap.Error(err))
		http.Error(w, "Failed to retrieve device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"assigned":   true,
		"device":     device,
		"assignment": assignment,
	})
}
