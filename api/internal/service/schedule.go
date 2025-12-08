package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ping/production-line-api/internal/domain"
	"ping/production-line-api/internal/external/openholidays"
	"ping/production-line-api/internal/repository"
)

// ScheduleService handles business logic for schedules
type ScheduleService struct {
	repo            *repository.ScheduleRepository
	lineRepo        *repository.LineRepository
	holidaysClient  *openholidays.Client
	holidaysCountry string
	validator       *validator.Validate
	logger          *zap.Logger
}

// NewScheduleService creates a new ScheduleService
func NewScheduleService(
	repo *repository.ScheduleRepository,
	lineRepo *repository.LineRepository,
	holidaysClient *openholidays.Client,
	holidaysCountry string,
	logger *zap.Logger,
) *ScheduleService {
	return &ScheduleService{
		repo:            repo,
		lineRepo:        lineRepo,
		holidaysClient:  holidaysClient,
		holidaysCountry: holidaysCountry,
		validator:       validator.New(),
		logger:          logger,
	}
}

// ========== Schedule CRUD ==========

// Create creates a new schedule
func (s *ScheduleService) Create(ctx context.Context, req domain.CreateScheduleRequest) (*domain.Schedule, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate days
	if err := s.validateDays(req.Days); err != nil {
		return nil, err
	}

	schedule, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create schedule",
			zap.String("name", req.Name),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("schedule created",
		zap.String("id", schedule.ID.String()),
		zap.String("name", schedule.Name))

	return schedule, nil
}

func (s *ScheduleService) validateDays(days []domain.CreateScheduleDayInput) error {
	for _, day := range days {
		if !day.DayOfWeek.IsValid() {
			return domain.ErrInvalidDayOfWeek
		}

		if day.IsWorkingDay {
			if day.ShiftStart == nil || day.ShiftEnd == nil {
				return domain.ErrMissingShiftTimes
			}
		} else {
			if day.ShiftStart != nil || day.ShiftEnd != nil {
				return domain.ErrUnexpectedShiftTimes
			}
		}

		// Validate breaks
		for _, brk := range day.Breaks {
			if brk.BreakStart >= brk.BreakEnd {
				return fmt.Errorf("break start must be before break end")
			}
		}
	}
	return nil
}

// GetByID retrieves a schedule by ID
func (s *ScheduleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	schedule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

// List retrieves all schedules
func (s *ScheduleService) List(ctx context.Context) ([]domain.ScheduleSummary, error) {
	schedules, err := s.repo.List(ctx)
	if err != nil {
		s.logger.Error("failed to list schedules", zap.Error(err))
		return nil, err
	}
	return schedules, nil
}

// Update updates a schedule
func (s *ScheduleService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateScheduleRequest) (*domain.Schedule, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	schedule, err := s.repo.Update(ctx, id, req)
	if err != nil {
		s.logger.Error("failed to update schedule",
			zap.String("id", id.String()),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("schedule updated",
		zap.String("id", schedule.ID.String()),
		zap.String("name", schedule.Name))

	return schedule, nil
}

// Delete soft deletes a schedule
func (s *ScheduleService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete schedule",
			zap.String("id", id.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("schedule deleted", zap.String("id", id.String()))
	return nil
}

// ========== Schedule Day Operations ==========

// GetDay retrieves a schedule day
func (s *ScheduleService) GetDay(ctx context.Context, dayID uuid.UUID) (*domain.ScheduleDay, error) {
	return s.repo.GetDay(ctx, dayID)
}

// UpdateDay updates a schedule day
func (s *ScheduleService) UpdateDay(ctx context.Context, dayID uuid.UUID, req domain.UpdateScheduleDayRequest) (*domain.ScheduleDay, error) {
	day, err := s.repo.UpdateDay(ctx, dayID, req)
	if err != nil {
		s.logger.Error("failed to update schedule day",
			zap.String("day_id", dayID.String()),
			zap.Error(err))
		return nil, err
	}
	return day, nil
}

// SetDayBreaks replaces all breaks for a day
func (s *ScheduleService) SetDayBreaks(ctx context.Context, dayID uuid.UUID, req domain.SetScheduleBreaksRequest) ([]domain.ScheduleBreak, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	breaks, err := s.repo.SetDayBreaks(ctx, dayID, req.Breaks)
	if err != nil {
		s.logger.Error("failed to set day breaks",
			zap.String("day_id", dayID.String()),
			zap.Error(err))
		return nil, err
	}
	return breaks, nil
}

// ========== Holiday Operations ==========

// GetHolidays retrieves all holidays for a schedule, optionally filtered by year
func (s *ScheduleService) GetHolidays(ctx context.Context, scheduleID uuid.UUID, year *int) ([]domain.ScheduleHoliday, error) {
	return s.repo.GetHolidays(ctx, scheduleID, year)
}

// CreateHoliday creates a holiday
func (s *ScheduleService) CreateHoliday(ctx context.Context, scheduleID uuid.UUID, req domain.CreateHolidayRequest) (*domain.ScheduleHoliday, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	holiday, err := s.repo.CreateHoliday(ctx, scheduleID, req)
	if err != nil {
		s.logger.Error("failed to create holiday",
			zap.String("schedule_id", scheduleID.String()),
			zap.String("date", req.HolidayDate),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("holiday created",
		zap.String("schedule_id", scheduleID.String()),
		zap.String("date", holiday.HolidayDate))

	return holiday, nil
}

// GetHoliday retrieves a holiday by ID
func (s *ScheduleService) GetHoliday(ctx context.Context, holidayID uuid.UUID) (*domain.ScheduleHoliday, error) {
	return s.repo.GetHoliday(ctx, holidayID)
}

// UpdateHoliday updates a holiday
func (s *ScheduleService) UpdateHoliday(ctx context.Context, holidayID uuid.UUID, req domain.UpdateHolidayRequest) (*domain.ScheduleHoliday, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	holiday, err := s.repo.UpdateHoliday(ctx, holidayID, req)
	if err != nil {
		s.logger.Error("failed to update holiday",
			zap.String("holiday_id", holidayID.String()),
			zap.Error(err))
		return nil, err
	}
	return holiday, nil
}

// DeleteHoliday deletes a holiday
func (s *ScheduleService) DeleteHoliday(ctx context.Context, holidayID uuid.UUID) error {
	if err := s.repo.DeleteHoliday(ctx, holidayID); err != nil {
		s.logger.Error("failed to delete holiday",
			zap.String("holiday_id", holidayID.String()),
			zap.Error(err))
		return err
	}
	return nil
}

// GetSuggestedHolidays fetches public holidays from external API
func (s *ScheduleService) GetSuggestedHolidays(ctx context.Context, year int) (*domain.SuggestedHolidaysResponse, error) {
	if s.holidaysClient == nil {
		errMsg := "holidays API not configured"
		return &domain.SuggestedHolidaysResponse{
			Holidays:    []domain.SuggestedHoliday{},
			CountryCode: s.holidaysCountry,
			Year:        year,
			Cached:      false,
			Error:       &errMsg,
		}, nil
	}

	holidays, cached, err := s.holidaysClient.GetPublicHolidays(ctx, s.holidaysCountry, year)
	if err != nil {
		s.logger.Warn("failed to fetch suggested holidays",
			zap.String("country", s.holidaysCountry),
			zap.Int("year", year),
			zap.Error(err))
		errMsg := err.Error()
		return &domain.SuggestedHolidaysResponse{
			Holidays:    []domain.SuggestedHoliday{},
			CountryCode: s.holidaysCountry,
			Year:        year,
			Cached:      false,
			Error:       &errMsg,
		}, nil
	}

	// Convert to domain type
	suggested := make([]domain.SuggestedHoliday, 0, len(holidays))
	for _, h := range holidays {
		suggested = append(suggested, domain.SuggestedHoliday{
			Date:       h.StartDate,
			Name:       openholidays.GetName(h.Name),
			Type:       h.Type,
			Nationwide: h.Nationwide,
		})
	}

	return &domain.SuggestedHolidaysResponse{
		Holidays:    suggested,
		CountryCode: s.holidaysCountry,
		Year:        year,
		Cached:      cached,
	}, nil
}

// ========== Schedule Exception Operations ==========

// GetExceptions retrieves all exceptions for a schedule
func (s *ScheduleService) GetExceptions(ctx context.Context, scheduleID uuid.UUID) ([]domain.ScheduleException, error) {
	return s.repo.GetExceptions(ctx, scheduleID)
}

// CreateException creates a schedule exception
func (s *ScheduleService) CreateException(ctx context.Context, scheduleID uuid.UUID, req domain.CreateScheduleExceptionRequest) (*domain.ScheduleException, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate date range
	if req.StartDate > req.EndDate {
		return nil, domain.ErrInvalidDateRange
	}

	exception, err := s.repo.CreateException(ctx, scheduleID, req)
	if err != nil {
		s.logger.Error("failed to create schedule exception",
			zap.String("schedule_id", scheduleID.String()),
			zap.String("name", req.Name),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("schedule exception created",
		zap.String("schedule_id", scheduleID.String()),
		zap.String("exception_id", exception.ID.String()),
		zap.String("name", exception.Name))

	return exception, nil
}

// GetException retrieves a schedule exception by ID
func (s *ScheduleService) GetException(ctx context.Context, exceptionID uuid.UUID) (*domain.ScheduleException, error) {
	return s.repo.GetException(ctx, exceptionID)
}

// UpdateException updates a schedule exception
func (s *ScheduleService) UpdateException(ctx context.Context, exceptionID uuid.UUID, req domain.UpdateScheduleExceptionRequest) (*domain.ScheduleException, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	exception, err := s.repo.UpdateException(ctx, exceptionID, req)
	if err != nil {
		s.logger.Error("failed to update schedule exception",
			zap.String("exception_id", exceptionID.String()),
			zap.Error(err))
		return nil, err
	}
	return exception, nil
}

// DeleteException deletes a schedule exception
func (s *ScheduleService) DeleteException(ctx context.Context, exceptionID uuid.UUID) error {
	if err := s.repo.DeleteException(ctx, exceptionID); err != nil {
		s.logger.Error("failed to delete schedule exception",
			zap.String("exception_id", exceptionID.String()),
			zap.Error(err))
		return err
	}
	return nil
}

// ========== Line Schedule Exception Operations ==========

// GetLineExceptions retrieves all line-specific exceptions for a schedule
func (s *ScheduleService) GetLineExceptions(ctx context.Context, scheduleID uuid.UUID) ([]domain.LineScheduleException, error) {
	return s.repo.GetLineExceptions(ctx, scheduleID)
}

// CreateLineException creates a line-specific exception
func (s *ScheduleService) CreateLineException(ctx context.Context, scheduleID uuid.UUID, req domain.CreateLineScheduleExceptionRequest) (*domain.LineScheduleException, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Validate date range
	if req.StartDate > req.EndDate {
		return nil, domain.ErrInvalidDateRange
	}

	// Validate at least one line
	if len(req.LineIDs) == 0 {
		return nil, domain.ErrMissingLinesToException
	}

	exception, err := s.repo.CreateLineException(ctx, scheduleID, req)
	if err != nil {
		s.logger.Error("failed to create line schedule exception",
			zap.String("schedule_id", scheduleID.String()),
			zap.String("name", req.Name),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("line schedule exception created",
		zap.String("schedule_id", scheduleID.String()),
		zap.String("exception_id", exception.ID.String()),
		zap.String("name", exception.Name),
		zap.Int("line_count", len(exception.LineIDs)))

	return exception, nil
}

// GetLineException retrieves a line-specific exception by ID
func (s *ScheduleService) GetLineException(ctx context.Context, exceptionID uuid.UUID) (*domain.LineScheduleException, error) {
	return s.repo.GetLineException(ctx, exceptionID)
}

// UpdateLineException updates a line-specific exception
func (s *ScheduleService) UpdateLineException(ctx context.Context, exceptionID uuid.UUID, req domain.UpdateLineScheduleExceptionRequest) (*domain.LineScheduleException, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	exception, err := s.repo.UpdateLineException(ctx, exceptionID, req)
	if err != nil {
		s.logger.Error("failed to update line schedule exception",
			zap.String("exception_id", exceptionID.String()),
			zap.Error(err))
		return nil, err
	}
	return exception, nil
}

// DeleteLineException deletes a line-specific exception
func (s *ScheduleService) DeleteLineException(ctx context.Context, exceptionID uuid.UUID) error {
	if err := s.repo.DeleteLineException(ctx, exceptionID); err != nil {
		s.logger.Error("failed to delete line schedule exception",
			zap.String("exception_id", exceptionID.String()),
			zap.Error(err))
		return err
	}
	return nil
}

// ========== Line Assignment ==========

// AssignScheduleToLine assigns a schedule to a production line
func (s *ScheduleService) AssignScheduleToLine(ctx context.Context, lineID uuid.UUID, scheduleID *uuid.UUID) error {
	// Verify line exists
	if _, err := s.lineRepo.GetByID(ctx, lineID); err != nil {
		return err
	}

	// Verify schedule exists if not null
	if scheduleID != nil {
		if _, err := s.repo.GetByID(ctx, *scheduleID); err != nil {
			return err
		}
	}

	if err := s.repo.AssignScheduleToLine(ctx, lineID, scheduleID); err != nil {
		s.logger.Error("failed to assign schedule to line",
			zap.String("line_id", lineID.String()),
			zap.Error(err))
		return err
	}

	s.logger.Info("schedule assigned to line",
		zap.String("line_id", lineID.String()))

	return nil
}

// GetLinesForSchedule retrieves all lines assigned to a schedule
func (s *ScheduleService) GetLinesForSchedule(ctx context.Context, scheduleID uuid.UUID) ([]domain.ProductionLine, error) {
	return s.repo.GetLinesForSchedule(ctx, scheduleID)
}

// ========== Effective Schedule ==========

// GetEffectiveSchedule computes the effective schedule for a line on a date
func (s *ScheduleService) GetEffectiveSchedule(ctx context.Context, lineID uuid.UUID, date string) (*domain.EffectiveSchedule, error) {
	return s.repo.GetEffectiveSchedule(ctx, lineID, date)
}

// GetEffectiveScheduleRange computes effective schedules for a line over a date range
func (s *ScheduleService) GetEffectiveScheduleRange(ctx context.Context, lineID uuid.UUID, startDate, endDate string) ([]domain.EffectiveSchedule, error) {
	// Validate date range
	if startDate > endDate {
		return nil, domain.ErrInvalidDateRange
	}

	return s.repo.GetEffectiveScheduleRange(ctx, lineID, startDate, endDate)
}
