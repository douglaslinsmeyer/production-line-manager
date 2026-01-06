package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/service"
)

// SignalProfileHandler handles HTTP requests for signal profiles
type SignalProfileHandler struct {
	service *service.SignalProfileService
	logger  *zap.Logger
}

// NewSignalProfileHandler creates a new SignalProfileHandler
func NewSignalProfileHandler(service *service.SignalProfileService, logger *zap.Logger) *SignalProfileHandler {
	return &SignalProfileHandler{
		service: service,
		logger:  logger,
	}
}

// ========== Profile CRUD Endpoints ==========

// List godoc
// @Summary List signal profiles
// @Description Get all active signal profiles
// @Tags profiles
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=[]domain.SignalProfile}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles [get]
func (h *SignalProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profiles, err := h.service.List(ctx)
	if err != nil {
		h.logger.Error("failed to list signal profiles", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to list signal profiles", nil)
		return
	}

	writeList(w, profiles, len(profiles))
}

// Get godoc
// @Summary Get signal profile
// @Description Get a signal profile by ID
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} Response{data=domain.SignalProfile}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id} [get]
func (h *SignalProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Get profile
	profile, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Signal profile not found", nil)
			return
		}
		h.logger.Error("failed to get signal profile", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get signal profile", nil)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// Create godoc
// @Summary Create signal profile
// @Description Create a new signal profile configuration
// @Tags profiles
// @Accept json
// @Produce json
// @Param profile body domain.CreateProfileRequest true "Profile details"
// @Success 201 {object} Response{data=domain.SignalProfile}
// @Failure 400 {object} Response{error=APIError}
// @Failure 409 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles [post]
func (h *SignalProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req domain.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// TODO: Get user from authentication context
	createdBy := "system" // Placeholder

	// Create profile
	profile, err := h.service.Create(ctx, req, createdBy)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNameExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Profile name already exists", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidButtonCycle) || errors.Is(err, domain.ErrInvalidDefaultState) ||
			errors.Is(err, domain.ErrInvalidLightMode) || errors.Is(err, domain.ErrInvalidBuzzerMode) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, err.Error(), nil)
			return
		}
		h.logger.Error("failed to create signal profile", zap.String("name", req.Name), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to create signal profile", nil)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, http.StatusOK, profile)
}

// Update godoc
// @Summary Update signal profile
// @Description Update a signal profile (triggers version increment if config changes)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param profile body domain.UpdateProfileRequest true "Profile updates"
// @Success 200 {object} Response{data=domain.SignalProfile}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 409 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id} [put]
func (h *SignalProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Parse request body
	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// TODO: Get user from authentication context
	updatedBy := "system" // Placeholder

	// Update profile
	profile, err := h.service.Update(ctx, id, req, updatedBy)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Signal profile not found", nil)
			return
		}
		if errors.Is(err, domain.ErrProfileNameExists) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Profile name already exists", nil)
			return
		}
		if errors.Is(err, domain.ErrInvalidButtonCycle) || errors.Is(err, domain.ErrInvalidDefaultState) ||
			errors.Is(err, domain.ErrInvalidLightMode) || errors.Is(err, domain.ErrInvalidBuzzerMode) {
			writeError(w, http.StatusBadRequest, ErrCodeValidation, err.Error(), nil)
			return
		}
		h.logger.Error("failed to update signal profile", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to update signal profile", nil)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// Delete godoc
// @Summary Delete signal profile
// @Description Delete a signal profile (soft delete, only if not in use)
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 409 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id} [delete]
func (h *SignalProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Delete profile
	err = h.service.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Signal profile not found", nil)
			return
		}
		if errors.Is(err, domain.ErrProfileInUse) {
			writeError(w, http.StatusConflict, ErrCodeConflict, "Profile is assigned to one or more lines", nil)
			return
		}
		h.logger.Error("failed to delete signal profile", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to delete signal profile", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Version Management Endpoints ==========

// GetVersions godoc
// @Summary Get profile versions
// @Description Get version history for a signal profile
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} Response{data=[]domain.ProfileVersion}
// @Failure 400 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id}/versions [get]
func (h *SignalProfileHandler) GetVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Get versions
	versions, err := h.service.GetVersions(ctx, id)
	if err != nil {
		h.logger.Error("failed to get profile versions", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get profile versions", nil)
		return
	}

	writeList(w, versions, len(versions))
}

// Rollback godoc
// @Summary Rollback profile
// @Description Rollback a profile to a previous version
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param request body domain.RollbackProfileRequest true "Rollback request"
// @Success 200 {object} Response{data=domain.SignalProfile}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id}/rollback [post]
func (h *SignalProfileHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Parse request body
	var req domain.RollbackProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// TODO: Get user from authentication context
	rolledBackBy := "system" // Placeholder

	// Rollback profile
	profile, err := h.service.RollbackToVersion(ctx, id, req, rolledBackBy)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Signal profile not found", nil)
			return
		}
		if errors.Is(err, domain.ErrProfileVersionNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Profile version not found", nil)
			return
		}
		h.logger.Error("failed to rollback profile", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to rollback profile", nil)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// GetDeviceStatus godoc
// @Summary Get device sync status
// @Description Get device synchronization status for a profile
// @Tags profiles
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} Response{data=domain.ProfileDeviceStatusResponse}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /profiles/{id}/device-status [get]
func (h *SignalProfileHandler) GetDeviceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse ID from URL
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidID, "Invalid profile ID", nil)
		return
	}

	// Get device status
	status, err := h.service.GetDeviceStatus(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProfileNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Signal profile not found", nil)
			return
		}
		h.logger.Error("failed to get device status", zap.String("id", id.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get device status", nil)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

// ========== Device Endpoints ==========

// DeviceHeartbeat godoc
// @Summary Device heartbeat
// @Description Process device heartbeat and check for profile updates
// @Tags devices
// @Accept json
// @Produce json
// @Param mac path string true "Device MAC Address"
// @Param heartbeat body domain.DeviceHeartbeatRequest true "Heartbeat data"
// @Success 200 {object} Response{data=domain.DeviceHeartbeatResponse}
// @Failure 400 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/{mac}/heartbeat [post]
func (h *SignalProfileHandler) DeviceHeartbeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get MAC address from URL
	mac := chi.URLParam(r, "mac")

	// Parse request body
	var req domain.DeviceHeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// Handle heartbeat
	response, err := h.service.HandleDeviceHeartbeat(ctx, mac, req)
	if err != nil {
		h.logger.Error("failed to handle device heartbeat", zap.String("mac", mac), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to process heartbeat", nil)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

// DeviceProfileUpdated godoc
// @Summary Confirm profile update
// @Description Device confirms successful profile update
// @Tags devices
// @Accept json
// @Produce json
// @Param mac path string true "Device MAC Address"
// @Param confirmation body domain.ProfileUpdatedRequest true "Update confirmation"
// @Success 200 {object} Response{data=map[string]bool}
// @Failure 400 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/{mac}/profile-updated [post]
func (h *SignalProfileHandler) DeviceProfileUpdated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get MAC address from URL
	mac := chi.URLParam(r, "mac")

	// Parse request body
	var req domain.ProfileUpdatedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// Confirm update
	err := h.service.ConfirmProfileUpdate(ctx, mac, req)
	if err != nil {
		h.logger.Error("failed to confirm profile update", zap.String("mac", mac), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to confirm update", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetDeviceState godoc
// @Summary Get device state
// @Description Get current state for a device
// @Tags devices
// @Accept json
// @Produce json
// @Param mac path string true "Device MAC Address"
// @Success 200 {object} Response{data=domain.DeviceSignalState}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/{mac}/signal-state [get]
func (h *SignalProfileHandler) GetDeviceState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get MAC address from URL
	mac := chi.URLParam(r, "mac")

	// Get device state
	state, err := h.service.GetDeviceState(ctx, mac)
	if err != nil {
		if errors.Is(err, domain.ErrDeviceStateNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Device state not found", nil)
			return
		}
		h.logger.Error("failed to get device state", zap.String("mac", mac), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get device state", nil)
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// SetDeviceState godoc
// @Summary Set device state
// @Description Manually set a device's state (creates override)
// @Tags devices
// @Accept json
// @Produce json
// @Param mac path string true "Device MAC Address"
// @Param state body domain.SetDeviceStateRequest true "State to set"
// @Success 200 {object} Response{data=map[string]bool}
// @Failure 400 {object} Response{error=APIError}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/{mac}/signal-state [put]
func (h *SignalProfileHandler) SetDeviceState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get MAC address from URL
	mac := chi.URLParam(r, "mac")

	// Parse request body
	var req domain.SetDeviceStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// Set device state
	err := h.service.SetDeviceState(ctx, mac, req)
	if err != nil {
		if errors.Is(err, domain.ErrDeviceStateNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Device state not found", nil)
			return
		}
		h.logger.Error("failed to set device state", zap.String("mac", mac), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to set device state", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// ResetDeviceOverride godoc
// @Summary Reset device override
// @Description Clear device override and return to default state
// @Tags devices
// @Accept json
// @Produce json
// @Param mac path string true "Device MAC Address"
// @Success 200 {object} Response{data=map[string]bool}
// @Failure 404 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/{mac}/reset-override [post]
func (h *SignalProfileHandler) ResetDeviceOverride(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get MAC address from URL
	mac := chi.URLParam(r, "mac")

	// Reset override
	err := h.service.ResetDeviceOverride(ctx, mac)
	if err != nil {
		if errors.Is(err, domain.ErrDeviceStateNotFound) {
			writeError(w, http.StatusNotFound, ErrCodeNotFound, "Device state not found", nil)
			return
		}
		h.logger.Error("failed to reset device override", zap.String("mac", mac), zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to reset override", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetOverriddenDevices godoc
// @Summary List overridden devices
// @Description Get all devices with active overrides
// @Tags devices
// @Accept json
// @Produce json
// @Success 200 {object} Response{data=[]domain.DeviceSignalState}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/overridden [get]
func (h *SignalProfileHandler) GetOverriddenDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get overridden devices
	devices, err := h.service.GetOverriddenDevices(ctx)
	if err != nil {
		h.logger.Error("failed to get overridden devices", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to get overridden devices", nil)
		return
	}

	writeList(w, devices, len(devices))
}

// BulkResetOverrides godoc
// @Summary Bulk reset overrides
// @Description Reset overrides for multiple devices
// @Tags devices
// @Accept json
// @Produce json
// @Param request body domain.BulkResetOverrideRequest true "Device MACs"
// @Success 200 {object} Response{data=map[string]bool}
// @Failure 400 {object} Response{error=APIError}
// @Failure 500 {object} Response{error=APIError}
// @Router /devices/bulk-reset-override [post]
func (h *SignalProfileHandler) BulkResetOverrides(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req domain.BulkResetOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", nil)
		return
	}

	// Reset overrides
	err := h.service.BulkResetOverrides(ctx, req.DeviceMACs)
	if err != nil {
		h.logger.Error("failed to bulk reset overrides", zap.Error(err))
		writeError(w, http.StatusInternalServerError, ErrCodeInternal, "Failed to reset overrides", nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
