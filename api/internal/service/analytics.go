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

// AnalyticsService handles business logic for analytics queries
type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
	lineRepo      *repository.LineRepository
	logger        *zap.Logger
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(
	analyticsRepo *repository.AnalyticsRepository,
	lineRepo      *repository.LineRepository,
	logger        *zap.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		lineRepo:      lineRepo,
		logger:        logger,
	}
}

// GetAggregateMetrics retrieves aggregate metrics across all or filtered lines
func (s *AnalyticsService) GetAggregateMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) (*domain.AggregateMetrics, error) {
	// Parse timeframe and set defaults if needed
	if err := s.parseTimeframe(&query); err != nil {
		return nil, err
	}

	metrics, err := s.analyticsRepo.GetAggregateMetrics(ctx, query)
	if err != nil {
		s.logger.Error("failed to get aggregate metrics", zap.Error(err))
		return nil, err
	}

	return metrics, nil
}

// GetLineMetrics retrieves metrics for each production line
func (s *AnalyticsService) GetLineMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) ([]domain.LineMetrics, error) {
	if err := s.parseTimeframe(&query); err != nil {
		return nil, err
	}

	metrics, err := s.analyticsRepo.GetLineMetrics(ctx, query)
	if err != nil {
		s.logger.Error("failed to get line metrics", zap.Error(err))
		return nil, err
	}

	return metrics, nil
}

// GetLabelMetrics retrieves metrics grouped by label
func (s *AnalyticsService) GetLabelMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) ([]domain.LabelMetrics, error) {
	if err := s.parseTimeframe(&query); err != nil {
		return nil, err
	}

	metrics, err := s.analyticsRepo.GetLabelMetrics(ctx, query)
	if err != nil {
		s.logger.Error("failed to get label metrics", zap.Error(err))
		return nil, err
	}

	return metrics, nil
}

// GetDailyKPIs retrieves KPIs for each 24-hour period
func (s *AnalyticsService) GetDailyKPIs(
	ctx context.Context,
	lineID uuid.UUID,
	query domain.AnalyticsQuery,
) ([]domain.DailyKPI, error) {
	if err := s.parseTimeframe(&query); err != nil {
		return nil, err
	}

	// Verify line exists
	if _, err := s.lineRepo.GetByID(ctx, lineID); err != nil {
		return nil, err
	}

	kpis, err := s.analyticsRepo.GetDailyKPIs(ctx, lineID, *query.StartTime, *query.EndTime)
	if err != nil {
		s.logger.Error("failed to get daily KPIs",
			zap.String("line_id", lineID.String()),
			zap.Error(err))
		return nil, err
	}

	return kpis, nil
}

// parseTimeframe converts timeframe string to actual start/end times
func (s *AnalyticsService) parseTimeframe(query *domain.AnalyticsQuery) error {
	now := time.Now()

	// If custom timeframe with explicit times, validate them
	if query.Timeframe == "custom" {
		if query.StartTime == nil || query.EndTime == nil {
			return fmt.Errorf("custom timeframe requires start_time and end_time")
		}
		if query.StartTime.After(*query.EndTime) {
			return fmt.Errorf("start_time must be before end_time")
		}
		return nil
	}

	// Parse predefined timeframes
	switch query.Timeframe {
	case "24h", "":
		start := now.Add(-24 * time.Hour)
		query.StartTime = &start
		query.EndTime = &now
	case "7d":
		start := now.Add(-7 * 24 * time.Hour)
		query.StartTime = &start
		query.EndTime = &now
	case "30d":
		start := now.Add(-30 * 24 * time.Hour)
		query.StartTime = &start
		query.EndTime = &now
	case "all":
		// Start from beginning of time
		start := time.Unix(0, 0)
		query.StartTime = &start
		query.EndTime = &now
	default:
		return fmt.Errorf("invalid timeframe: %s (valid: 24h, 7d, 30d, all, custom)", query.Timeframe)
	}

	return nil
}
