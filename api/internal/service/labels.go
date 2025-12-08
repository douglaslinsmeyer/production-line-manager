package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/repository"
)

// LabelService handles business logic for labels
type LabelService struct {
	repo      *repository.LabelRepository
	lineRepo  *repository.LineRepository
	publisher Publisher
	validator *validator.Validate
	logger    *zap.Logger
}

// NewLabelService creates a new LabelService
func NewLabelService(
	repo *repository.LabelRepository,
	lineRepo *repository.LineRepository,
	publisher Publisher,
	logger *zap.Logger,
) *LabelService {
	return &LabelService{
		repo:      repo,
		lineRepo:  lineRepo,
		publisher: publisher,
		validator: validator.New(),
		logger:    logger,
	}
}

// Create creates a new label
func (s *LabelService) Create(ctx context.Context, req domain.CreateLabelRequest) (*domain.Label, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	label, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create label", zap.String("name", req.Name), zap.Error(err))
		return nil, err
	}

	s.logger.Info("label created", zap.String("id", label.ID.String()), zap.String("name", label.Name))
	return label, nil
}

// GetByID retrieves a label by ID
func (s *LabelService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Label, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves all labels
func (s *LabelService) List(ctx context.Context) ([]domain.Label, error) {
	return s.repo.List(ctx)
}

// Update updates a label
func (s *LabelService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateLabelRequest) (*domain.Label, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	label, err := s.repo.Update(ctx, id, req)
	if err != nil {
		s.logger.Error("failed to update label", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	s.logger.Info("label updated", zap.String("id", label.ID.String()), zap.String("name", label.Name))
	return label, nil
}

// Delete deletes a label
func (s *LabelService) Delete(ctx context.Context, id uuid.UUID) error {
	label, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete label", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	s.logger.Info("label deleted", zap.String("id", id.String()), zap.String("name", label.Name))
	return nil
}

// AssignToLine assigns labels to a production line
func (s *LabelService) AssignToLine(ctx context.Context, lineID uuid.UUID, req domain.AssignLabelsRequest) error {
	if err := s.validator.Struct(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify line exists
	line, err := s.lineRepo.GetByID(ctx, lineID)
	if err != nil {
		return err
	}

	// Verify all labels exist
	for _, labelID := range req.LabelIDs {
		if _, err := s.repo.GetByID(ctx, labelID); err != nil {
			return fmt.Errorf("label %s not found: %w", labelID, err)
		}
	}

	if err := s.repo.AssignToLine(ctx, lineID, req.LabelIDs); err != nil {
		s.logger.Error("failed to assign labels", zap.String("line_id", lineID.String()), zap.Error(err))
		return err
	}

	s.logger.Info("labels assigned to line",
		zap.String("line_id", lineID.String()),
		zap.String("line_code", line.Code),
		zap.Int("label_count", len(req.LabelIDs)))

	// Publish label change event (best-effort)
	if err := s.publisher.PublishUpdated(line); err != nil {
		s.logger.Error("failed to publish label change event",
			zap.String("line_code", line.Code),
			zap.Error(err))
	}

	return nil
}

// GetLabelsForLine retrieves all labels for a production line
func (s *LabelService) GetLabelsForLine(ctx context.Context, lineID uuid.UUID) ([]domain.Label, error) {
	// Verify line exists
	if _, err := s.lineRepo.GetByID(ctx, lineID); err != nil {
		return nil, err
	}

	return s.repo.GetLabelsForLine(ctx, lineID)
}
