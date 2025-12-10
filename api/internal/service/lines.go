package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/repository"
)

// Publisher defines the interface for publishing events to MQTT
type Publisher interface {
	PublishCreated(line *domain.ProductionLine) error
	PublishUpdated(line *domain.ProductionLine) error
	PublishDeleted(id uuid.UUID, code string) error
	PublishStatus(line *domain.ProductionLine) error
	PublishRaw(topic string, payload []byte) error
}

// LineService handles business logic for production lines
type LineService struct {
	repo       *repository.LineRepository
	logRepo    *repository.StatusLogRepository
	deviceRepo domain.DeviceRepository
	publisher  Publisher
	validator  *validator.Validate
	logger     *zap.Logger
}

// NewLineService creates a new LineService
func NewLineService(
	repo *repository.LineRepository,
	logRepo *repository.StatusLogRepository,
	deviceRepo domain.DeviceRepository,
	publisher Publisher,
	logger *zap.Logger,
) *LineService {
	return &LineService{
		repo:       repo,
		logRepo:    logRepo,
		deviceRepo: deviceRepo,
		publisher:  publisher,
		validator:  validator.New(),
		logger:     logger,
	}
}

// Create creates a new production line
func (s *LineService) Create(ctx context.Context, req domain.CreateLineRequest) (*domain.ProductionLine, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create production line
	line, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create production line",
			zap.String("code", req.Code),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("production line created",
		zap.String("id", line.ID.String()),
		zap.String("code", line.Code),
		zap.String("name", line.Name))

	// Publish created event (best-effort, don't fail request on error)
	if err := s.publisher.PublishCreated(line); err != nil {
		s.logger.Error("failed to publish created event",
			zap.String("line_code", line.Code),
			zap.Error(err))
	}

	return line, nil
}

// GetByID retrieves a production line by its ID
func (s *LineService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductionLine, error) {
	line, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return line, nil
}

// GetByCode retrieves a production line by its code
func (s *LineService) GetByCode(ctx context.Context, code string) (*domain.ProductionLine, error) {
	line, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return line, nil
}

// List retrieves all active production lines
func (s *LineService) List(ctx context.Context) ([]domain.ProductionLine, error) {
	lines, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list production lines", zap.Error(err))
		return nil, err
	}
	return lines, nil
}

// Update updates a production line's details
func (s *LineService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateLineRequest) (*domain.ProductionLine, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update production line
	line, err := s.repo.Update(ctx, id, req)
	if err != nil {
		s.logger.Error("failed to update production line",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("production line updated",
		zap.String("id", line.ID.String()),
		zap.String("code", line.Code))

	// Publish updated event (best-effort)
	if err := s.publisher.PublishUpdated(line); err != nil {
		s.logger.Error("failed to publish updated event",
			zap.String("line_code", line.Code),
			zap.Error(err))
	}

	return line, nil
}

// Delete soft deletes a production line
func (s *LineService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get line first to publish event
	line, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete production line
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete production line",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("production line deleted",
		zap.String("id", id.String()),
		zap.String("code", line.Code))

	// Publish deleted event (best-effort)
	if err := s.publisher.PublishDeleted(id, line.Code); err != nil {
		s.logger.Error("failed to publish deleted event",
			zap.String("line_code", line.Code),
			zap.Error(err))
	}

	return nil
}

// SetStatus sets the status of a production line (idempotent)
func (s *LineService) SetStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.Status,
	source string,
	sourceDetail interface{},
) (*domain.ProductionLine, error) {
	// Validate status
	if !status.IsValid() {
		return nil, domain.ErrInvalidStatus
	}

	// Get current line state
	line, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := line.Status

	// Idempotent: if already in desired state, return success without logging
	if line.Status == status {
		s.logger.Debug("production line already in desired status",
			zap.String("id", id.String()),
			zap.String("code", line.Code),
			zap.String("status", status.String()))
		return line, nil
	}

	// Update status
	line, err = s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("failed to update production line status",
			zap.String("id", id.String()),
			zap.String("old_status", oldStatus.String()),
			zap.String("new_status", status.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("production line status changed",
		zap.String("id", line.ID.String()),
		zap.String("code", line.Code),
		zap.String("old_status", oldStatus.String()),
		zap.String("new_status", status.String()),
		zap.String("source", source))

	// Log status change to audit trail
	change := &domain.StatusChange{
		Time:         time.Now(),
		LineID:       line.ID,
		LineCode:     line.Code,
		OldStatus:    &oldStatus,
		NewStatus:    status,
		Source:       source,
		SourceDetail: sourceDetail,
	}

	if err := s.logRepo.Insert(ctx, change); err != nil {
		s.logger.Error("failed to log status change",
			zap.String("line_code", line.Code),
			zap.Error(err))
		// Don't fail the request, status was already updated
	}

	// Publish status event (best-effort)
	if err := s.publisher.PublishStatus(line); err != nil {
		s.logger.Error("failed to publish status event",
			zap.String("line_code", line.Code),
			zap.Error(err))
	}

	// Send tower light control commands to assigned device
	if err := s.updateDeviceTowerLights(ctx, line); err != nil {
		s.logger.Error("failed to update device tower lights",
			zap.String("line_code", line.Code),
			zap.String("line_id", line.ID.String()),
			zap.Error(err))
		// Don't fail the request, status change was successful
	}

	return line, nil
}

// updateDeviceTowerLights sends output control commands to device assigned to this line
// Tower light mapping: CH0=Green (on), CH1=Yellow (maintenance), CH2=Red (off/error)
func (s *LineService) updateDeviceTowerLights(ctx context.Context, line *domain.ProductionLine) error {
	// Get device assigned to this line
	assignment, err := s.deviceRepo.GetLineAssignment(line.ID)
	if err != nil {
		return fmt.Errorf("failed to get line assignment: %w", err)
	}

	if assignment == nil {
		// No device assigned - nothing to do
		s.logger.Debug("no device assigned to line",
			zap.String("line_code", line.Code))
		return nil
	}

	s.logger.Info("updating tower lights for assigned device",
		zap.String("line_code", line.Code),
		zap.String("device_mac", assignment.DeviceMAC),
		zap.String("status", line.Status.String()))

	// Determine which outputs to turn on based on status
	var greenOn, yellowOn, redOn bool

	switch line.Status {
	case domain.StatusOn:
		greenOn = true
		yellowOn = false
		redOn = false
	case domain.StatusMaintenance:
		greenOn = false
		yellowOn = true
		redOn = false
	case domain.StatusOff:
		greenOn = false
		yellowOn = false
		redOn = true
	case domain.StatusError:
		greenOn = false
		yellowOn = false
		redOn = true
	}

	// Send commands to device
	topic := fmt.Sprintf("devices/%s/command", assignment.DeviceMAC)

	// Command to set green light (CH0)
	if err := s.sendOutputCommand(topic, 0, greenOn); err != nil {
		return err
	}

	// Command to set yellow light (CH1)
	if err := s.sendOutputCommand(topic, 1, yellowOn); err != nil {
		return err
	}

	// Command to set red light (CH2)
	if err := s.sendOutputCommand(topic, 2, redOn); err != nil {
		return err
	}

	s.logger.Info("tower lights updated",
		zap.String("line_code", line.Code),
		zap.Bool("green", greenOn),
		zap.Bool("yellow", yellowOn),
		zap.Bool("red", redOn))

	return nil
}

// sendOutputCommand sends a set_output command to a device
func (s *LineService) sendOutputCommand(topic string, channel int, state bool) error {
	command := map[string]interface{}{
		"command": "set_output",
		"channel": channel,
		"state":   state,
	}

	payload, err := json.Marshal(command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	return s.publisher.PublishRaw(topic, payload)
}

// GetStatusHistory retrieves the status change history for a production line
func (s *LineService) GetStatusHistory(ctx context.Context, id uuid.UUID, limit int) ([]domain.StatusChange, error) {
	// Verify line exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return nil, err
	}

	// Get history
	history, err := s.logRepo.GetHistory(ctx, id, limit)
	if err != nil {
		s.logger.Error("failed to get status history",
			zap.String("line_id", id.String()),
			zap.Error(err))
		return nil, err
	}

	return history, nil
}
