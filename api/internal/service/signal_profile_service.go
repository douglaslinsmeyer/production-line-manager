package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/repository"
)

// ProfilePublisher defines the interface for publishing profile events
type ProfilePublisher interface {
	PublishProfileCreated(profile *domain.SignalProfile) error
	PublishProfileUpdated(profile *domain.SignalProfile, affectedDevices int) error
	PublishProfileDeleted(id uuid.UUID, name string) error
	PublishProfileAssigned(lineID uuid.UUID, profile *domain.SignalProfile) error
	PublishDeviceProfileUpdate(deviceMAC string, profile *domain.SignalProfile) error
}

// SignalProfileService handles business logic for signal profiles
type SignalProfileService struct {
	repo       *repository.SignalProfileRepository
	lineRepo   *repository.LineRepository
	deviceRepo domain.DeviceRepository
	publisher  ProfilePublisher
	validator  *validator.Validate
	logger     *zap.Logger
}

// NewSignalProfileService creates a new SignalProfileService
func NewSignalProfileService(
	repo *repository.SignalProfileRepository,
	lineRepo *repository.LineRepository,
	deviceRepo domain.DeviceRepository,
	publisher ProfilePublisher,
	logger *zap.Logger,
) *SignalProfileService {
	return &SignalProfileService{
		repo:       repo,
		lineRepo:   lineRepo,
		deviceRepo: deviceRepo,
		publisher:  publisher,
		validator:  validator.New(),
		logger:     logger,
	}
}

// ========== Core CRUD Operations ==========

// Create creates a new signal profile
func (s *SignalProfileService) Create(ctx context.Context, req domain.CreateProfileRequest, createdBy string) (*domain.SignalProfile, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate profile configuration
	if err := s.validateStateReferences(req.States, req.ButtonBehavior, req.DefaultState); err != nil {
		return nil, err
	}

	// Validate output modes
	for _, state := range req.States {
		if !state.Outputs.RedLight.IsValid() || !state.Outputs.YellowLight.IsValid() || !state.Outputs.GreenLight.IsValid() {
			return nil, domain.ErrInvalidLightMode
		}
		if !state.Outputs.Buzzer.IsValid() {
			return nil, domain.ErrInvalidBuzzerMode
		}
	}

	// Create profile
	profile := &domain.SignalProfile{
		Name:           req.Name,
		Description:    req.Description,
		Version:        1,
		States:         req.States,
		ButtonBehavior: req.ButtonBehavior,
		DefaultState:   req.DefaultState,
		UpdatedBy:      &createdBy,
	}

	if err := s.repo.Create(ctx, profile); err != nil {
		s.logger.Error("failed to create signal profile",
			zap.String("name", req.Name),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("signal profile created",
		zap.String("id", profile.ID.String()),
		zap.String("name", profile.Name),
		zap.Int("states", len(profile.States)))

	// Create initial version history
	if err := s.createVersionSnapshot(profile, createdBy, []string{"Initial profile creation"}, nil); err != nil {
		s.logger.Error("failed to create version snapshot",
			zap.String("profile_id", profile.ID.String()),
			zap.Error(err))
		// Don't fail the request
	}

	// Publish created event (best-effort)
	if err := s.publisher.PublishProfileCreated(profile); err != nil {
		s.logger.Error("failed to publish profile created event",
			zap.String("profile_id", profile.ID.String()),
			zap.Error(err))
	}

	return profile, nil
}

// GetByID retrieves a signal profile by ID
func (s *SignalProfileService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SignalProfile, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all signal profiles
func (s *SignalProfileService) List(ctx context.Context) ([]*domain.SignalProfile, error) {
	return s.repo.List(ctx)
}

// Update updates a signal profile
func (s *SignalProfileService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateProfileRequest, updatedBy string) (*domain.SignalProfile, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing profile
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = req.Description
	}

	// Track if configuration changed (not just metadata)
	configChanged := false

	if req.States != nil {
		configChanged = true
		existing.States = *req.States
	}
	if req.ButtonBehavior != nil {
		configChanged = true
		existing.ButtonBehavior = *req.ButtonBehavior
	}
	if req.DefaultState != nil {
		configChanged = true
		existing.DefaultState = *req.DefaultState
	}

	// Validate updated profile configuration
	if req.States != nil || req.ButtonBehavior != nil || req.DefaultState != nil {
		if err := s.validateStateReferences(existing.States, existing.ButtonBehavior, existing.DefaultState); err != nil {
			return nil, err
		}
	}

	// Validate output modes if states changed
	if req.States != nil {
		for _, state := range existing.States {
			if !state.Outputs.RedLight.IsValid() || !state.Outputs.YellowLight.IsValid() || !state.Outputs.GreenLight.IsValid() {
				return nil, domain.ErrInvalidLightMode
			}
			if !state.Outputs.Buzzer.IsValid() {
				return nil, domain.ErrInvalidBuzzerMode
			}
		}
	}

	// Increment version if configuration changed
	if configChanged {
		// Get old version for comparison
		oldProfile := &domain.SignalProfile{
			States:         existing.States,
			ButtonBehavior: existing.ButtonBehavior,
			DefaultState:   existing.DefaultState,
		}

		existing.Version++
		existing.UpdatedBy = &updatedBy

		// Detect changes
		changes := s.detectChanges(existing, oldProfile)

		// Create version snapshot
		if err := s.createVersionSnapshot(existing, updatedBy, changes, nil); err != nil {
			s.logger.Error("failed to create version snapshot",
				zap.String("profile_id", id.String()),
				zap.Error(err))
			// Don't fail the request
		}

		// Mark all devices with this profile as update-pending
		devices, err := s.repo.GetDevicesWithProfile(ctx, id)
		if err != nil {
			s.logger.Error("failed to get devices for update notification",
				zap.String("profile_id", id.String()),
				zap.Error(err))
		} else {
			for _, device := range devices {
				if err := s.repo.UpdateDeviceVersionStatus(ctx, device.DeviceMAC, string(domain.VersionStatusUpdatePending)); err != nil {
					s.logger.Error("failed to mark device as update-pending",
						zap.String("device_mac", device.DeviceMAC),
						zap.Error(err))
				}
			}
		}

		s.logger.Info("signal profile updated (version incremented)",
			zap.String("id", id.String()),
			zap.Int("new_version", existing.Version),
			zap.Int("affected_devices", len(devices)))

		// Publish updated event
		if err := s.publisher.PublishProfileUpdated(existing, len(devices)); err != nil {
			s.logger.Error("failed to publish profile updated event",
				zap.String("profile_id", id.String()),
				zap.Error(err))
		}
	} else {
		existing.UpdatedBy = &updatedBy
		s.logger.Info("signal profile metadata updated (no version change)",
			zap.String("id", id.String()))
	}

	// Update in database
	if err := s.repo.Update(ctx, id, existing); err != nil {
		s.logger.Error("failed to update signal profile",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}

	return existing, nil
}

// Delete deletes a signal profile (soft delete)
func (s *SignalProfileService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if profile is in use
	inUse, err := s.repo.IsProfileInUse(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrProfileInUse
	}

	// Get profile for logging
	profile, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete profile
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete signal profile",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("signal profile deleted",
		zap.String("id", id.String()),
		zap.String("name", profile.Name))

	// Publish deleted event
	if err := s.publisher.PublishProfileDeleted(id, profile.Name); err != nil {
		s.logger.Error("failed to publish profile deleted event",
			zap.String("profile_id", id.String()),
			zap.Error(err))
	}

	return nil
}

// ========== Validation Logic ==========

// validateStateReferences validates that button cycles and default state reference existing states
func (s *SignalProfileService) validateStateReferences(states []domain.ProfileState, buttonBehavior domain.ButtonBehavior, defaultState string) error {
	// Build map of state names
	stateMap := make(map[string]bool)
	for _, state := range states {
		if stateMap[state.Name] {
			return fmt.Errorf("duplicate state name: %s", state.Name)
		}
		stateMap[state.Name] = true
	}

	// Validate short press cycle
	for _, stateName := range buttonBehavior.ShortPressCycle {
		if !stateMap[stateName] {
			return fmt.Errorf("%w: %s in short press cycle", domain.ErrInvalidButtonCycle, stateName)
		}
	}

	// Validate long press cycle
	for _, stateName := range buttonBehavior.LongPressCycle {
		if !stateMap[stateName] {
			return fmt.Errorf("%w: %s in long press cycle", domain.ErrInvalidButtonCycle, stateName)
		}
	}

	// Validate default state
	if !stateMap[defaultState] {
		return fmt.Errorf("%w: %s", domain.ErrInvalidDefaultState, defaultState)
	}

	return nil
}

// detectChanges compares two profiles and returns a list of changes
func (s *SignalProfileService) detectChanges(new, old *domain.SignalProfile) []string {
	var changes []string

	// Compare states
	if len(new.States) != len(old.States) {
		changes = append(changes, fmt.Sprintf("Number of states changed from %d to %d", len(old.States), len(new.States)))
	} else {
		// Check for modified states
		oldStateMap := make(map[string]domain.ProfileState)
		for _, state := range old.States {
			oldStateMap[state.Name] = state
		}

		for _, newState := range new.States {
			if oldState, exists := oldStateMap[newState.Name]; exists {
				// Compare outputs
				if newState.Outputs != oldState.Outputs {
					changes = append(changes, fmt.Sprintf("State '%s' output configuration changed", newState.Name))
				}
			} else {
				changes = append(changes, fmt.Sprintf("State '%s' added", newState.Name))
			}
		}

		// Check for removed states
		newStateMap := make(map[string]bool)
		for _, state := range new.States {
			newStateMap[state.Name] = true
		}
		for _, oldState := range old.States {
			if !newStateMap[oldState.Name] {
				changes = append(changes, fmt.Sprintf("State '%s' removed", oldState.Name))
			}
		}
	}

	// Compare button behavior
	if fmt.Sprint(new.ButtonBehavior.ShortPressCycle) != fmt.Sprint(old.ButtonBehavior.ShortPressCycle) {
		changes = append(changes, "Short press cycle modified")
	}
	if fmt.Sprint(new.ButtonBehavior.LongPressCycle) != fmt.Sprint(old.ButtonBehavior.LongPressCycle) {
		changes = append(changes, "Long press cycle modified")
	}

	// Compare default state
	if new.DefaultState != old.DefaultState {
		changes = append(changes, fmt.Sprintf("Default state changed from '%s' to '%s'", old.DefaultState, new.DefaultState))
	}

	return changes
}

// ========== Version Management ==========

// createVersionSnapshot creates a version history entry
func (s *SignalProfileService) createVersionSnapshot(profile *domain.SignalProfile, changedBy string, changes []string, description *string) error {
	version := &domain.ProfileVersion{
		ProfileID:         profile.ID,
		Version:           profile.Version,
		Config:            profile,
		ChangedBy:         changedBy,
		Changes:           changes,
		ChangeDescription: description,
	}

	return s.repo.CreateVersion(context.Background(), version)
}

// GetVersions retrieves version history for a profile
func (s *SignalProfileService) GetVersions(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileVersion, error) {
	return s.repo.GetVersions(ctx, profileID)
}

// RollbackToVersion rolls back a profile to a previous version
func (s *SignalProfileService) RollbackToVersion(ctx context.Context, profileID uuid.UUID, req domain.RollbackProfileRequest, rolledBackBy string) (*domain.SignalProfile, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get current profile
	current, err := s.repo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// Get target version
	targetVersion, err := s.repo.GetVersionByNumber(ctx, profileID, req.TargetVersion)
	if err != nil {
		return nil, err
	}

	// Apply rollback (create new version with old config)
	current.States = targetVersion.Config.States
	current.ButtonBehavior = targetVersion.Config.ButtonBehavior
	current.DefaultState = targetVersion.Config.DefaultState
	current.Version++
	current.UpdatedBy = &rolledBackBy

	// Update database
	if err := s.repo.Update(ctx, profileID, current); err != nil {
		return nil, err
	}

	// Create version snapshot
	changes := []string{fmt.Sprintf("Rolled back to version %d", req.TargetVersion)}
	if err := s.createVersionSnapshot(current, rolledBackBy, changes, &req.Reason); err != nil {
		s.logger.Error("failed to create version snapshot",
			zap.String("profile_id", profileID.String()),
			zap.Error(err))
	}

	// Mark devices as update-pending
	devices, err := s.repo.GetDevicesWithProfile(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to get devices for rollback notification",
			zap.String("profile_id", profileID.String()),
			zap.Error(err))
	} else {
		for _, device := range devices {
			if err := s.repo.UpdateDeviceVersionStatus(ctx, device.DeviceMAC, string(domain.VersionStatusUpdatePending)); err != nil {
				s.logger.Error("failed to mark device as update-pending",
					zap.String("device_mac", device.DeviceMAC),
					zap.Error(err))
			}
		}
	}

	s.logger.Info("signal profile rolled back",
		zap.String("id", profileID.String()),
		zap.Int("target_version", req.TargetVersion),
		zap.Int("new_version", current.Version))

	// Publish updated event
	if err := s.publisher.PublishProfileUpdated(current, len(devices)); err != nil {
		s.logger.Error("failed to publish profile updated event",
			zap.String("profile_id", profileID.String()),
			zap.Error(err))
	}

	return current, nil
}

// GetDeviceStatus retrieves device sync status for a profile
func (s *SignalProfileService) GetDeviceStatus(ctx context.Context, profileID uuid.UUID) (*domain.ProfileDeviceStatusResponse, error) {
	// Get profile
	profile, err := s.repo.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	// Get devices with this profile
	deviceStates, err := s.repo.GetDevicesWithProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}

	response := &domain.ProfileDeviceStatusResponse{
		ProfileID:      profileID,
		CurrentVersion: profile.Version,
		Devices:        make([]domain.DeviceVersionStatus, 0),
	}

	// Build device status list
	for _, state := range deviceStates {
		// Determine status
		status := state.VersionStatus
		if state.ProfileVersion == profile.Version {
			status = string(domain.VersionStatusUpToDate)
		} else if state.ProfileVersion < profile.Version {
			status = string(domain.VersionStatusUpdatePending)
		}

		// Get assigned line (if any)
		assignment, err := s.deviceRepo.GetDeviceAssignment(state.DeviceMAC)
		var assignedLine *uuid.UUID
		if err == nil && assignment != nil {
			assignedLine = &assignment.LineID
		}

		deviceStatus := domain.DeviceVersionStatus{
			DeviceMAC:    state.DeviceMAC,
			DeviceID:     state.DeviceMAC, // TODO: Get actual device ID if different
			Version:      state.ProfileVersion,
			Status:       status,
			LastCheck:    state.LastVersionCheck,
			AssignedLine: assignedLine,
		}
		response.Devices = append(response.Devices, deviceStatus)

		// Update summary
		switch domain.VersionStatus(status) {
		case domain.VersionStatusUpToDate:
			response.Summary.UpToDate++
		case domain.VersionStatusUpdatePending:
			response.Summary.UpdatePending++
		case domain.VersionStatusFailed:
			response.Summary.Failed++
		case domain.VersionStatusOffline:
			response.Summary.Offline++
		}
	}

	return response, nil
}

// ========== Line Assignment ==========

// AssignToLine assigns a profile to a production line
func (s *SignalProfileService) AssignToLine(ctx context.Context, lineID uuid.UUID, profileID uuid.UUID) error {
	// Verify profile exists
	profile, err := s.repo.GetByID(ctx, profileID)
	if err != nil {
		return err
	}

	// Verify line exists
	line, err := s.lineRepo.GetByID(ctx, lineID)
	if err != nil {
		return err
	}

	// Assign profile to line
	if err := s.repo.AssignToLine(ctx, lineID, profileID); err != nil {
		s.logger.Error("failed to assign profile to line",
			zap.String("line_id", lineID.String()),
			zap.String("profile_id", profileID.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("profile assigned to line",
		zap.String("line_id", lineID.String()),
		zap.String("line_code", line.Code),
		zap.String("profile_id", profileID.String()),
		zap.String("profile_name", profile.Name))

	// Publish assignment event
	if err := s.publisher.PublishProfileAssigned(lineID, profile); err != nil {
		s.logger.Error("failed to publish profile assigned event",
			zap.String("line_id", lineID.String()),
			zap.Error(err))
	}

	return nil
}

// GetLineProfile retrieves the profile assigned to a line
func (s *SignalProfileService) GetLineProfile(ctx context.Context, lineID uuid.UUID) (*domain.SignalProfile, error) {
	return s.repo.GetLineProfile(ctx, lineID)
}

// ========== Device Operations ==========

// HandleDeviceHeartbeat processes a device heartbeat and checks for profile updates
func (s *SignalProfileService) HandleDeviceHeartbeat(ctx context.Context, deviceMAC string, req domain.DeviceHeartbeatRequest) (*domain.DeviceHeartbeatResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get device's line assignment
	assignment, err := s.deviceRepo.GetDeviceAssignment(deviceMAC)
	if err != nil || assignment == nil {
		// Device not assigned to a line
		return &domain.DeviceHeartbeatResponse{
			LatestVersion:   0,
			UpdateAvailable: false,
		}, nil
	}

	// Get line's profile
	profile, err := s.repo.GetLineProfile(ctx, assignment.LineID)
	if err != nil {
		if err == domain.ErrProfileNotAssignedToLine {
			// Line has no profile assigned
			return &domain.DeviceHeartbeatResponse{
				LatestVersion:   0,
				UpdateAvailable: false,
			}, nil
		}
		return nil, err
	}

	// Update device state
	deviceState := &domain.DeviceSignalState{
		DeviceMAC:        deviceMAC,
		ProfileID:        &profile.ID,
		ProfileVersion:   req.ProfileVersion,
		CurrentState:     req.CurrentState,
		IsOverridden:     req.IsOverridden,
		LastSync:         time.Now(),
		LastVersionCheck: time.Now(),
	}

	// Determine version status
	if req.ProfileVersion == profile.Version {
		deviceState.VersionStatus = string(domain.VersionStatusUpToDate)
	} else if req.ProfileVersion < profile.Version {
		deviceState.VersionStatus = string(domain.VersionStatusUpdatePending)
	} else {
		// Device has newer version than backend (shouldn't happen)
		deviceState.VersionStatus = string(domain.VersionStatusUnknown)
		s.logger.Warn("device has newer version than backend",
			zap.String("device_mac", deviceMAC),
			zap.Int("device_version", req.ProfileVersion),
			zap.Int("backend_version", profile.Version))
	}

	// Upsert device state
	if err := s.repo.UpsertDeviceState(ctx, deviceState); err != nil {
		s.logger.Error("failed to upsert device state",
			zap.String("device_mac", deviceMAC),
			zap.Error(err))
		// Don't fail the heartbeat
	}

	// Build response
	response := &domain.DeviceHeartbeatResponse{
		LatestVersion:   profile.Version,
		UpdateAvailable: req.ProfileVersion < profile.Version,
	}

	// Include profile if update available
	if response.UpdateAvailable {
		response.Profile = profile
		s.logger.Info("profile update available for device",
			zap.String("device_mac", deviceMAC),
			zap.Int("device_version", req.ProfileVersion),
			zap.Int("latest_version", profile.Version))
	}

	return response, nil
}

// ConfirmProfileUpdate confirms that a device successfully updated its profile
func (s *SignalProfileService) ConfirmProfileUpdate(ctx context.Context, deviceMAC string, req domain.ProfileUpdatedRequest) error {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get device state
	deviceState, err := s.repo.GetDeviceState(ctx, deviceMAC)
	if err != nil {
		if err == domain.ErrDeviceStateNotFound {
			// Create new state
			deviceState = &domain.DeviceSignalState{
				DeviceMAC: deviceMAC,
			}
		} else {
			return err
		}
	}

	// Update device state
	deviceState.ProfileID = &req.ProfileID
	deviceState.ProfileVersion = req.NewVersion
	deviceState.CurrentState = req.CurrentState
	deviceState.LastSync = time.Now()
	deviceState.LastVersionCheck = time.Now()
	deviceState.VersionStatus = string(domain.VersionStatusUpToDate)

	// If state changed during migration, log it
	if req.StateChanged {
		s.logger.Info("device state changed during profile update",
			zap.String("device_mac", deviceMAC),
			zap.String("new_state", req.CurrentState),
			zap.Int("new_version", req.NewVersion))
	}

	// Upsert device state
	if err := s.repo.UpsertDeviceState(ctx, deviceState); err != nil {
		s.logger.Error("failed to upsert device state",
			zap.String("device_mac", deviceMAC),
			zap.Error(err))
		return err
	}

	s.logger.Info("device profile update confirmed",
		zap.String("device_mac", deviceMAC),
		zap.Int("previous_version", req.PreviousVersion),
		zap.Int("new_version", req.NewVersion))

	return nil
}

// SetDeviceState manually sets a device's state (creates override)
func (s *SignalProfileService) SetDeviceState(ctx context.Context, deviceMAC string, req domain.SetDeviceStateRequest) error {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get device state
	deviceState, err := s.repo.GetDeviceState(ctx, deviceMAC)
	if err != nil {
		return err
	}

	// Verify state exists in profile
	profile, err := s.repo.GetByID(ctx, *deviceState.ProfileID)
	if err != nil {
		return err
	}

	stateExists := false
	for _, state := range profile.States {
		if state.Name == req.State {
			stateExists = true
			break
		}
	}
	if !stateExists {
		return fmt.Errorf("state '%s' does not exist in profile", req.State)
	}

	// Update device state
	deviceState.CurrentState = req.State
	deviceState.IsOverridden = true
	deviceState.LastStateChange = time.Now()

	if err := s.repo.UpsertDeviceState(ctx, deviceState); err != nil {
		return err
	}

	s.logger.Info("device state manually set",
		zap.String("device_mac", deviceMAC),
		zap.String("state", req.State))

	return nil
}

// ResetDeviceOverride clears the override flag and returns device to default state
func (s *SignalProfileService) ResetDeviceOverride(ctx context.Context, deviceMAC string) error {
	// Get device state
	deviceState, err := s.repo.GetDeviceState(ctx, deviceMAC)
	if err != nil {
		return err
	}

	// Get profile default state
	profile, err := s.repo.GetByID(ctx, *deviceState.ProfileID)
	if err != nil {
		return err
	}

	// Update device state
	deviceState.CurrentState = profile.DefaultState
	deviceState.IsOverridden = false
	deviceState.LastStateChange = time.Now()

	if err := s.repo.UpsertDeviceState(ctx, deviceState); err != nil {
		return err
	}

	s.logger.Info("device override reset",
		zap.String("device_mac", deviceMAC),
		zap.String("default_state", profile.DefaultState))

	return nil
}

// GetOverriddenDevices retrieves all devices with active overrides
func (s *SignalProfileService) GetOverriddenDevices(ctx context.Context) ([]*domain.DeviceSignalState, error) {
	return s.repo.GetOverriddenDevices(ctx)
}

// BulkResetOverrides resets overrides for multiple devices
func (s *SignalProfileService) BulkResetOverrides(ctx context.Context, deviceMACs []string) error {
	for _, deviceMAC := range deviceMACs {
		if err := s.ResetDeviceOverride(ctx, deviceMAC); err != nil {
			s.logger.Error("failed to reset override for device",
				zap.String("device_mac", deviceMAC),
				zap.Error(err))
			// Continue with other devices
		}
	}
	return nil
}

// GetDeviceState retrieves device signal state
func (s *SignalProfileService) GetDeviceState(ctx context.Context, deviceMAC string) (*domain.DeviceSignalState, error) {
	return s.repo.GetDeviceState(ctx, deviceMAC)
}

// ========== Helper Methods ==========

// calculateProfileHash calculates MD5 hash of profile for validation
func (s *SignalProfileService) calculateProfileHash(profile *domain.SignalProfile) string {
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return ""
	}
	hash := md5.Sum(profileJSON)
	return fmt.Sprintf("%x", hash)
}
