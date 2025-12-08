package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/repository"
)

// ComplianceService handles business logic for schedule compliance
type ComplianceService struct {
	complianceRepo *repository.ComplianceRepository
	lineRepo       *repository.LineRepository
	logger         *zap.Logger
}

// NewComplianceService creates a new ComplianceService
func NewComplianceService(
	complianceRepo *repository.ComplianceRepository,
	lineRepo *repository.LineRepository,
	logger *zap.Logger,
) *ComplianceService {
	return &ComplianceService{
		complianceRepo: complianceRepo,
		lineRepo:       lineRepo,
		logger:         logger,
	}
}

// GetAggregateCompliance returns aggregate compliance metrics
func (s *ComplianceService) GetAggregateCompliance(
	ctx context.Context,
	query domain.ComplianceQuery,
) (*domain.AggregateComplianceMetrics, error) {
	// Validate date format
	if err := s.validateDateRange(query.StartDate, query.EndDate); err != nil {
		return nil, err
	}

	s.logger.Info("calculating aggregate compliance",
		zap.String("start_date", query.StartDate),
		zap.String("end_date", query.EndDate),
		zap.Int("line_count", len(query.LineIDs)),
		zap.Int("label_count", len(query.LabelIDs)),
	)

	metrics, err := s.complianceRepo.GetAggregateComplianceMetrics(ctx, query)
	if err != nil {
		s.logger.Error("failed to get aggregate compliance", zap.Error(err))
		return nil, err
	}

	return metrics, nil
}

// GetLineCompliance returns compliance metrics for specific lines
func (s *ComplianceService) GetLineCompliance(
	ctx context.Context,
	query domain.ComplianceQuery,
) ([]domain.LineComplianceMetrics, error) {
	// Validate date format
	if err := s.validateDateRange(query.StartDate, query.EndDate); err != nil {
		return nil, err
	}

	s.logger.Info("calculating line compliance",
		zap.String("start_date", query.StartDate),
		zap.String("end_date", query.EndDate),
	)

	metrics, err := s.complianceRepo.GetLineComplianceMetrics(ctx, query)
	if err != nil {
		s.logger.Error("failed to get line compliance", zap.Error(err))
		return nil, err
	}

	return metrics, nil
}

// GetDailyComplianceKPIs returns daily compliance data for a specific line
func (s *ComplianceService) GetDailyComplianceKPIs(
	ctx context.Context,
	lineID uuid.UUID,
	startDate, endDate string,
) ([]domain.DailyComplianceKPI, error) {
	// Validate date format
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Verify line exists
	_, err := s.lineRepo.GetByID(ctx, lineID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("calculating daily compliance KPIs",
		zap.String("line_id", lineID.String()),
		zap.String("start_date", startDate),
		zap.String("end_date", endDate),
	)

	kpis, err := s.complianceRepo.GetDailyComplianceKPIs(ctx, lineID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get daily compliance KPIs", zap.Error(err))
		return nil, err
	}

	return kpis, nil
}

// validateDateRange validates that dates are in correct format and range is valid
func (s *ComplianceService) validateDateRange(startDate, endDate string) error {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format, expected YYYY-MM-DD: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return fmt.Errorf("invalid end_date format, expected YYYY-MM-DD: %w", err)
	}

	if end.Before(start) {
		return fmt.Errorf("end_date must be on or after start_date")
	}

	// Limit range to 90 days to prevent excessive computation
	if end.Sub(start) > 90*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 90 days")
	}

	return nil
}
