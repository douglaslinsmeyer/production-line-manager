package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ping/production-line-api/internal/domain"
)

// ScheduleRepository handles database operations for schedules
type ScheduleRepository struct {
	db *pgxpool.Pool
}

// NewScheduleRepository creates a new ScheduleRepository
func NewScheduleRepository(db *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// ========== Schedule CRUD ==========

// Create creates a new schedule with all 7 days
func (r *ScheduleRepository) Create(ctx context.Context, req domain.CreateScheduleRequest) (*domain.Schedule, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert schedule
	scheduleQuery := `
		INSERT INTO schedules (name, description, timezone)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, timezone, created_at, updated_at
	`

	var schedule domain.Schedule
	err = tx.QueryRow(ctx, scheduleQuery, req.Name, req.Description, req.Timezone).Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.Description,
		&schedule.Timezone,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrScheduleNameExists
		}
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	// Insert days
	schedule.Days = make([]domain.ScheduleDay, 0, 7)
	for _, dayInput := range req.Days {
		day, err := r.insertScheduleDay(ctx, tx, schedule.ID, dayInput)
		if err != nil {
			return nil, err
		}
		schedule.Days = append(schedule.Days, *day)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &schedule, nil
}

func (r *ScheduleRepository) insertScheduleDay(ctx context.Context, tx pgx.Tx, scheduleID uuid.UUID, input domain.CreateScheduleDayInput) (*domain.ScheduleDay, error) {
	dayQuery := `
		INSERT INTO schedule_days (schedule_id, day_of_week, is_working_day, shift_start, shift_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, schedule_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
	`

	var day domain.ScheduleDay
	err := tx.QueryRow(ctx, dayQuery,
		scheduleID,
		input.DayOfWeek,
		input.IsWorkingDay,
		input.ShiftStart,
		input.ShiftEnd,
	).Scan(
		&day.ID,
		&day.ScheduleID,
		&day.DayOfWeek,
		&day.IsWorkingDay,
		&day.ShiftStart,
		&day.ShiftEnd,
		&day.CreatedAt,
		&day.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert schedule day: %w", err)
	}

	// Insert breaks for this day
	day.Breaks = make([]domain.ScheduleBreak, 0, len(input.Breaks))
	for _, breakInput := range input.Breaks {
		brk, err := r.insertScheduleBreak(ctx, tx, day.ID, breakInput)
		if err != nil {
			return nil, err
		}
		day.Breaks = append(day.Breaks, *brk)
	}

	return &day, nil
}

func (r *ScheduleRepository) insertScheduleBreak(ctx context.Context, tx pgx.Tx, dayID uuid.UUID, input domain.CreateScheduleBreakInput) (*domain.ScheduleBreak, error) {
	breakQuery := `
		INSERT INTO schedule_breaks (schedule_day_id, name, break_start, break_end)
		VALUES ($1, $2, $3, $4)
		RETURNING id, schedule_day_id, name, break_start, break_end, created_at
	`

	var brk domain.ScheduleBreak
	err := tx.QueryRow(ctx, breakQuery, dayID, input.Name, input.BreakStart, input.BreakEnd).Scan(
		&brk.ID,
		&brk.ScheduleDayID,
		&brk.Name,
		&brk.BreakStart,
		&brk.BreakEnd,
		&brk.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert schedule break: %w", err)
	}

	return &brk, nil
}

// GetByID retrieves a schedule by ID with all its days and breaks
func (r *ScheduleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Schedule, error) {
	query := `
		SELECT id, name, description, timezone, created_at, updated_at, deleted_at
		FROM schedules
		WHERE id = $1 AND deleted_at IS NULL
	`

	var schedule domain.Schedule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.Description,
		&schedule.Timezone,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
		&schedule.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	// Load days with breaks
	days, err := r.getDaysForSchedule(ctx, schedule.ID)
	if err != nil {
		return nil, err
	}
	schedule.Days = days

	// Load holidays (all, no year filter)
	holidays, err := r.GetHolidays(ctx, schedule.ID, nil)
	if err != nil {
		return nil, err
	}
	schedule.Holidays = holidays

	return &schedule, nil
}

func (r *ScheduleRepository) getDaysForSchedule(ctx context.Context, scheduleID uuid.UUID) ([]domain.ScheduleDay, error) {
	query := `
		SELECT id, schedule_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
		FROM schedule_days
		WHERE schedule_id = $1
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule days: %w", err)
	}
	defer rows.Close()

	var days []domain.ScheduleDay
	dayIDs := make([]uuid.UUID, 0, 7)
	for rows.Next() {
		var day domain.ScheduleDay
		err := rows.Scan(
			&day.ID,
			&day.ScheduleID,
			&day.DayOfWeek,
			&day.IsWorkingDay,
			&day.ShiftStart,
			&day.ShiftEnd,
			&day.CreatedAt,
			&day.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule day: %w", err)
		}
		days = append(days, day)
		dayIDs = append(dayIDs, day.ID)
	}

	if days == nil {
		days = []domain.ScheduleDay{}
	}

	// Load breaks for all days
	if len(dayIDs) > 0 {
		breaksMap, err := r.getBreaksForDays(ctx, dayIDs)
		if err != nil {
			return nil, err
		}
		for i := range days {
			if breaks, ok := breaksMap[days[i].ID]; ok {
				days[i].Breaks = breaks
			} else {
				days[i].Breaks = []domain.ScheduleBreak{}
			}
		}
	}

	return days, nil
}

func (r *ScheduleRepository) getBreaksForDays(ctx context.Context, dayIDs []uuid.UUID) (map[uuid.UUID][]domain.ScheduleBreak, error) {
	query := `
		SELECT id, schedule_day_id, name, break_start, break_end, created_at
		FROM schedule_breaks
		WHERE schedule_day_id = ANY($1)
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule breaks: %w", err)
	}
	defer rows.Close()

	breaksMap := make(map[uuid.UUID][]domain.ScheduleBreak)
	for rows.Next() {
		var brk domain.ScheduleBreak
		err := rows.Scan(
			&brk.ID,
			&brk.ScheduleDayID,
			&brk.Name,
			&brk.BreakStart,
			&brk.BreakEnd,
			&brk.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule break: %w", err)
		}
		breaksMap[brk.ScheduleDayID] = append(breaksMap[brk.ScheduleDayID], brk)
	}

	return breaksMap, nil
}

// List retrieves all active schedules with line counts
func (r *ScheduleRepository) List(ctx context.Context) ([]domain.ScheduleSummary, error) {
	query := `
		SELECT s.id, s.name, s.description, s.timezone, s.created_at, s.updated_at,
		       COUNT(pl.id) as line_count
		FROM schedules s
		LEFT JOIN production_lines pl ON pl.schedule_id = s.id AND pl.deleted_at IS NULL
		WHERE s.deleted_at IS NULL
		GROUP BY s.id, s.name, s.description, s.timezone, s.created_at, s.updated_at
		ORDER BY s.name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []domain.ScheduleSummary
	for rows.Next() {
		var s domain.ScheduleSummary
		err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Description,
			&s.Timezone,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.LineCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", err)
		}
		schedules = append(schedules, s)
	}

	if schedules == nil {
		schedules = []domain.ScheduleSummary{}
	}

	return schedules, nil
}

// Update updates a schedule's metadata
func (r *ScheduleRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateScheduleRequest) (*domain.Schedule, error) {
	query := `
		UPDATE schedules
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    timezone = COALESCE($4, timezone),
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, description, timezone, created_at, updated_at
	`

	var schedule domain.Schedule
	err := r.db.QueryRow(ctx, query, id, req.Name, req.Description, req.Timezone).Scan(
		&schedule.ID,
		&schedule.Name,
		&schedule.Description,
		&schedule.Timezone,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrScheduleNameExists
		}
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	// Load days
	days, err := r.getDaysForSchedule(ctx, schedule.ID)
	if err != nil {
		return nil, err
	}
	schedule.Days = days

	return &schedule, nil
}

// Delete soft deletes a schedule
func (r *ScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE schedules
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrScheduleNotFound
	}

	return nil
}

// ========== Schedule Day Operations ==========

// GetDay retrieves a schedule day by ID
func (r *ScheduleRepository) GetDay(ctx context.Context, dayID uuid.UUID) (*domain.ScheduleDay, error) {
	query := `
		SELECT id, schedule_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
		FROM schedule_days
		WHERE id = $1
	`

	var day domain.ScheduleDay
	err := r.db.QueryRow(ctx, query, dayID).Scan(
		&day.ID,
		&day.ScheduleID,
		&day.DayOfWeek,
		&day.IsWorkingDay,
		&day.ShiftStart,
		&day.ShiftEnd,
		&day.CreatedAt,
		&day.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleDayNotFound
		}
		return nil, fmt.Errorf("failed to get schedule day: %w", err)
	}

	// Load breaks
	breaks, err := r.getBreaksForDay(ctx, day.ID)
	if err != nil {
		return nil, err
	}
	day.Breaks = breaks

	return &day, nil
}

func (r *ScheduleRepository) getBreaksForDay(ctx context.Context, dayID uuid.UUID) ([]domain.ScheduleBreak, error) {
	query := `
		SELECT id, schedule_day_id, name, break_start, break_end, created_at
		FROM schedule_breaks
		WHERE schedule_day_id = $1
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get breaks for day: %w", err)
	}
	defer rows.Close()

	var breaks []domain.ScheduleBreak
	for rows.Next() {
		var brk domain.ScheduleBreak
		err := rows.Scan(
			&brk.ID,
			&brk.ScheduleDayID,
			&brk.Name,
			&brk.BreakStart,
			&brk.BreakEnd,
			&brk.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan break: %w", err)
		}
		breaks = append(breaks, brk)
	}

	if breaks == nil {
		breaks = []domain.ScheduleBreak{}
	}

	return breaks, nil
}

// UpdateDay updates a schedule day
func (r *ScheduleRepository) UpdateDay(ctx context.Context, dayID uuid.UUID, req domain.UpdateScheduleDayRequest) (*domain.ScheduleDay, error) {
	query := `
		UPDATE schedule_days
		SET is_working_day = COALESCE($2, is_working_day),
		    shift_start = COALESCE($3, shift_start),
		    shift_end = COALESCE($4, shift_end),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, schedule_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
	`

	var day domain.ScheduleDay
	err := r.db.QueryRow(ctx, query, dayID, req.IsWorkingDay, req.ShiftStart, req.ShiftEnd).Scan(
		&day.ID,
		&day.ScheduleID,
		&day.DayOfWeek,
		&day.IsWorkingDay,
		&day.ShiftStart,
		&day.ShiftEnd,
		&day.CreatedAt,
		&day.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleDayNotFound
		}
		return nil, fmt.Errorf("failed to update schedule day: %w", err)
	}

	// Load breaks
	breaks, err := r.getBreaksForDay(ctx, day.ID)
	if err != nil {
		return nil, err
	}
	day.Breaks = breaks

	return &day, nil
}

// SetDayBreaks replaces all breaks for a day
func (r *ScheduleRepository) SetDayBreaks(ctx context.Context, dayID uuid.UUID, breaks []domain.CreateScheduleBreakInput) ([]domain.ScheduleBreak, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing breaks
	_, err = tx.Exec(ctx, `DELETE FROM schedule_breaks WHERE schedule_day_id = $1`, dayID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing breaks: %w", err)
	}

	// Insert new breaks
	result := make([]domain.ScheduleBreak, 0, len(breaks))
	for _, input := range breaks {
		brk, err := r.insertScheduleBreak(ctx, tx, dayID, input)
		if err != nil {
			return nil, err
		}
		result = append(result, *brk)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// ========== Holiday Operations ==========

// GetHolidays retrieves all holidays for a schedule, optionally filtered by year
func (r *ScheduleRepository) GetHolidays(ctx context.Context, scheduleID uuid.UUID, year *int) ([]domain.ScheduleHoliday, error) {
	var query string
	var args []interface{}

	if year != nil {
		query = `
			SELECT id, schedule_id, holiday_date::text, name, created_at
			FROM schedule_holidays
			WHERE schedule_id = $1 AND EXTRACT(YEAR FROM holiday_date) = $2
			ORDER BY holiday_date ASC
		`
		args = []interface{}{scheduleID, *year}
	} else {
		query = `
			SELECT id, schedule_id, holiday_date::text, name, created_at
			FROM schedule_holidays
			WHERE schedule_id = $1
			ORDER BY holiday_date ASC
		`
		args = []interface{}{scheduleID}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get holidays: %w", err)
	}
	defer rows.Close()

	var holidays []domain.ScheduleHoliday
	for rows.Next() {
		var h domain.ScheduleHoliday
		err := rows.Scan(&h.ID, &h.ScheduleID, &h.HolidayDate, &h.Name, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan holiday: %w", err)
		}
		holidays = append(holidays, h)
	}

	if holidays == nil {
		holidays = []domain.ScheduleHoliday{}
	}

	return holidays, nil
}

// CreateHoliday creates a new holiday
func (r *ScheduleRepository) CreateHoliday(ctx context.Context, scheduleID uuid.UUID, req domain.CreateHolidayRequest) (*domain.ScheduleHoliday, error) {
	query := `
		INSERT INTO schedule_holidays (schedule_id, holiday_date, name)
		VALUES ($1, $2, $3)
		RETURNING id, schedule_id, holiday_date::text, name, created_at
	`

	var h domain.ScheduleHoliday
	err := r.db.QueryRow(ctx, query, scheduleID, req.HolidayDate, req.Name).Scan(
		&h.ID, &h.ScheduleID, &h.HolidayDate, &h.Name, &h.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrHolidayDateExists
		}
		return nil, fmt.Errorf("failed to create holiday: %w", err)
	}

	return &h, nil
}

// GetHoliday retrieves a holiday by ID
func (r *ScheduleRepository) GetHoliday(ctx context.Context, holidayID uuid.UUID) (*domain.ScheduleHoliday, error) {
	query := `
		SELECT id, schedule_id, holiday_date::text, name, created_at
		FROM schedule_holidays
		WHERE id = $1
	`

	var h domain.ScheduleHoliday
	err := r.db.QueryRow(ctx, query, holidayID).Scan(
		&h.ID, &h.ScheduleID, &h.HolidayDate, &h.Name, &h.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHolidayNotFound
		}
		return nil, fmt.Errorf("failed to get holiday: %w", err)
	}

	return &h, nil
}

// UpdateHoliday updates a holiday
func (r *ScheduleRepository) UpdateHoliday(ctx context.Context, holidayID uuid.UUID, req domain.UpdateHolidayRequest) (*domain.ScheduleHoliday, error) {
	query := `
		UPDATE schedule_holidays
		SET holiday_date = COALESCE($2, holiday_date),
		    name = COALESCE($3, name)
		WHERE id = $1
		RETURNING id, schedule_id, holiday_date::text, name, created_at
	`

	var h domain.ScheduleHoliday
	err := r.db.QueryRow(ctx, query, holidayID, req.HolidayDate, req.Name).Scan(
		&h.ID, &h.ScheduleID, &h.HolidayDate, &h.Name, &h.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHolidayNotFound
		}
		return nil, fmt.Errorf("failed to update holiday: %w", err)
	}

	return &h, nil
}

// DeleteHoliday deletes a holiday
func (r *ScheduleRepository) DeleteHoliday(ctx context.Context, holidayID uuid.UUID) error {
	query := `DELETE FROM schedule_holidays WHERE id = $1`

	result, err := r.db.Exec(ctx, query, holidayID)
	if err != nil {
		return fmt.Errorf("failed to delete holiday: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrHolidayNotFound
	}

	return nil
}

// ========== Schedule Exception Operations ==========

// GetExceptions retrieves all exceptions for a schedule
func (r *ScheduleRepository) GetExceptions(ctx context.Context, scheduleID uuid.UUID) ([]domain.ScheduleException, error) {
	query := `
		SELECT id, schedule_id, name, description, start_date, end_date, created_at, updated_at
		FROM schedule_exceptions
		WHERE schedule_id = $1
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []domain.ScheduleException
	for rows.Next() {
		var e domain.ScheduleException
		err := rows.Scan(&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exception: %w", err)
		}
		exceptions = append(exceptions, e)
	}

	if exceptions == nil {
		exceptions = []domain.ScheduleException{}
	}

	return exceptions, nil
}

// CreateException creates a schedule exception with its days
func (r *ScheduleRepository) CreateException(ctx context.Context, scheduleID uuid.UUID, req domain.CreateScheduleExceptionRequest) (*domain.ScheduleException, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check for overlapping exceptions
	overlapQuery := `
		SELECT EXISTS(
			SELECT 1 FROM schedule_exceptions
			WHERE schedule_id = $1
			  AND start_date <= $3
			  AND end_date >= $2
		)
	`
	var hasOverlap bool
	err = tx.QueryRow(ctx, overlapQuery, scheduleID, req.StartDate, req.EndDate).Scan(&hasOverlap)
	if err != nil {
		return nil, fmt.Errorf("failed to check overlap: %w", err)
	}
	if hasOverlap {
		return nil, domain.ErrExceptionDatesOverlap
	}

	// Insert exception
	exceptionQuery := `
		INSERT INTO schedule_exceptions (schedule_id, name, description, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, schedule_id, name, description, start_date, end_date, created_at, updated_at
	`

	var e domain.ScheduleException
	err = tx.QueryRow(ctx, exceptionQuery, scheduleID, req.Name, req.Description, req.StartDate, req.EndDate).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create exception: %w", err)
	}

	// Insert days
	e.Days = make([]domain.ScheduleExceptionDay, 0, 7)
	for _, dayInput := range req.Days {
		day, err := r.insertExceptionDay(ctx, tx, e.ID, dayInput)
		if err != nil {
			return nil, err
		}
		e.Days = append(e.Days, *day)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &e, nil
}

func (r *ScheduleRepository) insertExceptionDay(ctx context.Context, tx pgx.Tx, exceptionID uuid.UUID, input domain.CreateScheduleExceptionDayInput) (*domain.ScheduleExceptionDay, error) {
	dayQuery := `
		INSERT INTO schedule_exception_days (exception_id, day_of_week, is_working_day, shift_start, shift_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, exception_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
	`

	var day domain.ScheduleExceptionDay
	err := tx.QueryRow(ctx, dayQuery, exceptionID, input.DayOfWeek, input.IsWorkingDay, input.ShiftStart, input.ShiftEnd).Scan(
		&day.ID, &day.ExceptionID, &day.DayOfWeek, &day.IsWorkingDay, &day.ShiftStart, &day.ShiftEnd, &day.CreatedAt, &day.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert exception day: %w", err)
	}

	// Insert breaks
	day.Breaks = make([]domain.ScheduleExceptionBreak, 0, len(input.Breaks))
	for _, breakInput := range input.Breaks {
		brk, err := r.insertExceptionBreak(ctx, tx, day.ID, breakInput)
		if err != nil {
			return nil, err
		}
		day.Breaks = append(day.Breaks, *brk)
	}

	return &day, nil
}

func (r *ScheduleRepository) insertExceptionBreak(ctx context.Context, tx pgx.Tx, dayID uuid.UUID, input domain.CreateScheduleExceptionBreakInput) (*domain.ScheduleExceptionBreak, error) {
	breakQuery := `
		INSERT INTO schedule_exception_breaks (exception_day_id, name, break_start, break_end)
		VALUES ($1, $2, $3, $4)
		RETURNING id, exception_day_id, name, break_start, break_end, created_at
	`

	var brk domain.ScheduleExceptionBreak
	err := tx.QueryRow(ctx, breakQuery, dayID, input.Name, input.BreakStart, input.BreakEnd).Scan(
		&brk.ID, &brk.ExceptionDayID, &brk.Name, &brk.BreakStart, &brk.BreakEnd, &brk.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert exception break: %w", err)
	}

	return &brk, nil
}

// GetException retrieves a schedule exception by ID with its days
func (r *ScheduleRepository) GetException(ctx context.Context, exceptionID uuid.UUID) (*domain.ScheduleException, error) {
	query := `
		SELECT id, schedule_id, name, description, start_date, end_date, created_at, updated_at
		FROM schedule_exceptions
		WHERE id = $1
	`

	var e domain.ScheduleException
	err := r.db.QueryRow(ctx, query, exceptionID).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrExceptionNotFound
		}
		return nil, fmt.Errorf("failed to get exception: %w", err)
	}

	// Load days
	days, err := r.getExceptionDays(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Days = days

	return &e, nil
}

func (r *ScheduleRepository) getExceptionDays(ctx context.Context, exceptionID uuid.UUID) ([]domain.ScheduleExceptionDay, error) {
	query := `
		SELECT id, exception_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
		FROM schedule_exception_days
		WHERE exception_id = $1
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.Query(ctx, query, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exception days: %w", err)
	}
	defer rows.Close()

	var days []domain.ScheduleExceptionDay
	dayIDs := make([]uuid.UUID, 0, 7)
	for rows.Next() {
		var day domain.ScheduleExceptionDay
		err := rows.Scan(
			&day.ID, &day.ExceptionID, &day.DayOfWeek, &day.IsWorkingDay,
			&day.ShiftStart, &day.ShiftEnd, &day.CreatedAt, &day.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exception day: %w", err)
		}
		days = append(days, day)
		dayIDs = append(dayIDs, day.ID)
	}

	if days == nil {
		days = []domain.ScheduleExceptionDay{}
	}

	// Load breaks
	if len(dayIDs) > 0 {
		breaksMap, err := r.getExceptionBreaksForDays(ctx, dayIDs)
		if err != nil {
			return nil, err
		}
		for i := range days {
			if breaks, ok := breaksMap[days[i].ID]; ok {
				days[i].Breaks = breaks
			} else {
				days[i].Breaks = []domain.ScheduleExceptionBreak{}
			}
		}
	}

	return days, nil
}

func (r *ScheduleRepository) getExceptionBreaksForDays(ctx context.Context, dayIDs []uuid.UUID) (map[uuid.UUID][]domain.ScheduleExceptionBreak, error) {
	query := `
		SELECT id, exception_day_id, name, break_start, break_end, created_at
		FROM schedule_exception_breaks
		WHERE exception_day_id = ANY($1)
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get exception breaks: %w", err)
	}
	defer rows.Close()

	breaksMap := make(map[uuid.UUID][]domain.ScheduleExceptionBreak)
	for rows.Next() {
		var brk domain.ScheduleExceptionBreak
		err := rows.Scan(&brk.ID, &brk.ExceptionDayID, &brk.Name, &brk.BreakStart, &brk.BreakEnd, &brk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exception break: %w", err)
		}
		breaksMap[brk.ExceptionDayID] = append(breaksMap[brk.ExceptionDayID], brk)
	}

	return breaksMap, nil
}

// UpdateException updates a schedule exception metadata
func (r *ScheduleRepository) UpdateException(ctx context.Context, exceptionID uuid.UUID, req domain.UpdateScheduleExceptionRequest) (*domain.ScheduleException, error) {
	query := `
		UPDATE schedule_exceptions
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    start_date = COALESCE($4, start_date),
		    end_date = COALESCE($5, end_date),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, schedule_id, name, description, start_date, end_date, created_at, updated_at
	`

	var e domain.ScheduleException
	err := r.db.QueryRow(ctx, query, exceptionID, req.Name, req.Description, req.StartDate, req.EndDate).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrExceptionNotFound
		}
		return nil, fmt.Errorf("failed to update exception: %w", err)
	}

	// Load days
	days, err := r.getExceptionDays(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Days = days

	return &e, nil
}

// DeleteException deletes a schedule exception
func (r *ScheduleRepository) DeleteException(ctx context.Context, exceptionID uuid.UUID) error {
	query := `DELETE FROM schedule_exceptions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, exceptionID)
	if err != nil {
		return fmt.Errorf("failed to delete exception: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrExceptionNotFound
	}

	return nil
}

// ========== Line Schedule Exception Operations ==========

// GetLineExceptions retrieves all line-specific exceptions for a schedule
func (r *ScheduleRepository) GetLineExceptions(ctx context.Context, scheduleID uuid.UUID) ([]domain.LineScheduleException, error) {
	query := `
		SELECT id, schedule_id, name, description, start_date, end_date, created_at, updated_at
		FROM line_schedule_exceptions
		WHERE schedule_id = $1
		ORDER BY start_date ASC
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line exceptions: %w", err)
	}
	defer rows.Close()

	var exceptions []domain.LineScheduleException
	exceptionIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var e domain.LineScheduleException
		err := rows.Scan(&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line exception: %w", err)
		}
		exceptions = append(exceptions, e)
		exceptionIDs = append(exceptionIDs, e.ID)
	}

	if exceptions == nil {
		exceptions = []domain.LineScheduleException{}
	}

	// Load line IDs for each exception
	if len(exceptionIDs) > 0 {
		linesMap, err := r.getLineIDsForExceptions(ctx, exceptionIDs)
		if err != nil {
			return nil, err
		}
		for i := range exceptions {
			if lineIDs, ok := linesMap[exceptions[i].ID]; ok {
				exceptions[i].LineIDs = lineIDs
			} else {
				exceptions[i].LineIDs = []uuid.UUID{}
			}
		}
	}

	return exceptions, nil
}

func (r *ScheduleRepository) getLineIDsForExceptions(ctx context.Context, exceptionIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	query := `
		SELECT exception_id, line_id
		FROM line_schedule_exception_lines
		WHERE exception_id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, exceptionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get line IDs for exceptions: %w", err)
	}
	defer rows.Close()

	linesMap := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var exceptionID, lineID uuid.UUID
		err := rows.Scan(&exceptionID, &lineID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line ID: %w", err)
		}
		linesMap[exceptionID] = append(linesMap[exceptionID], lineID)
	}

	return linesMap, nil
}

// CreateLineException creates a line-specific exception
func (r *ScheduleRepository) CreateLineException(ctx context.Context, scheduleID uuid.UUID, req domain.CreateLineScheduleExceptionRequest) (*domain.LineScheduleException, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert exception
	exceptionQuery := `
		INSERT INTO line_schedule_exceptions (schedule_id, name, description, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, schedule_id, name, description, start_date, end_date, created_at, updated_at
	`

	var e domain.LineScheduleException
	err = tx.QueryRow(ctx, exceptionQuery, scheduleID, req.Name, req.Description, req.StartDate, req.EndDate).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create line exception: %w", err)
	}

	// Insert line associations
	lineQuery := `INSERT INTO line_schedule_exception_lines (exception_id, line_id) VALUES ($1, $2)`
	for _, lineID := range req.LineIDs {
		_, err = tx.Exec(ctx, lineQuery, e.ID, lineID)
		if err != nil {
			return nil, fmt.Errorf("failed to associate line with exception: %w", err)
		}
	}
	e.LineIDs = req.LineIDs

	// Insert days
	e.Days = make([]domain.LineScheduleExceptionDay, 0, 7)
	for _, dayInput := range req.Days {
		day, err := r.insertLineExceptionDay(ctx, tx, e.ID, dayInput)
		if err != nil {
			return nil, err
		}
		e.Days = append(e.Days, *day)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &e, nil
}

func (r *ScheduleRepository) insertLineExceptionDay(ctx context.Context, tx pgx.Tx, exceptionID uuid.UUID, input domain.CreateLineScheduleExceptionDayInput) (*domain.LineScheduleExceptionDay, error) {
	dayQuery := `
		INSERT INTO line_schedule_exception_days (exception_id, day_of_week, is_working_day, shift_start, shift_end)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, exception_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
	`

	var day domain.LineScheduleExceptionDay
	err := tx.QueryRow(ctx, dayQuery, exceptionID, input.DayOfWeek, input.IsWorkingDay, input.ShiftStart, input.ShiftEnd).Scan(
		&day.ID, &day.ExceptionID, &day.DayOfWeek, &day.IsWorkingDay, &day.ShiftStart, &day.ShiftEnd, &day.CreatedAt, &day.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert line exception day: %w", err)
	}

	// Insert breaks
	day.Breaks = make([]domain.LineScheduleExceptionBreak, 0, len(input.Breaks))
	for _, breakInput := range input.Breaks {
		brk, err := r.insertLineExceptionBreak(ctx, tx, day.ID, breakInput)
		if err != nil {
			return nil, err
		}
		day.Breaks = append(day.Breaks, *brk)
	}

	return &day, nil
}

func (r *ScheduleRepository) insertLineExceptionBreak(ctx context.Context, tx pgx.Tx, dayID uuid.UUID, input domain.CreateLineScheduleExceptionBreakInput) (*domain.LineScheduleExceptionBreak, error) {
	breakQuery := `
		INSERT INTO line_schedule_exception_breaks (exception_day_id, name, break_start, break_end)
		VALUES ($1, $2, $3, $4)
		RETURNING id, exception_day_id, name, break_start, break_end, created_at
	`

	var brk domain.LineScheduleExceptionBreak
	err := tx.QueryRow(ctx, breakQuery, dayID, input.Name, input.BreakStart, input.BreakEnd).Scan(
		&brk.ID, &brk.ExceptionDayID, &brk.Name, &brk.BreakStart, &brk.BreakEnd, &brk.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert line exception break: %w", err)
	}

	return &brk, nil
}

// GetLineException retrieves a line-specific exception by ID
func (r *ScheduleRepository) GetLineException(ctx context.Context, exceptionID uuid.UUID) (*domain.LineScheduleException, error) {
	query := `
		SELECT id, schedule_id, name, description, start_date, end_date, created_at, updated_at
		FROM line_schedule_exceptions
		WHERE id = $1
	`

	var e domain.LineScheduleException
	err := r.db.QueryRow(ctx, query, exceptionID).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLineExceptionNotFound
		}
		return nil, fmt.Errorf("failed to get line exception: %w", err)
	}

	// Load line IDs
	lineIDs, err := r.getLineIDsForException(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.LineIDs = lineIDs

	// Load days
	days, err := r.getLineExceptionDays(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Days = days

	return &e, nil
}

func (r *ScheduleRepository) getLineIDsForException(ctx context.Context, exceptionID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT line_id FROM line_schedule_exception_lines WHERE exception_id = $1`

	rows, err := r.db.Query(ctx, query, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line IDs: %w", err)
	}
	defer rows.Close()

	var lineIDs []uuid.UUID
	for rows.Next() {
		var lineID uuid.UUID
		if err := rows.Scan(&lineID); err != nil {
			return nil, fmt.Errorf("failed to scan line ID: %w", err)
		}
		lineIDs = append(lineIDs, lineID)
	}

	if lineIDs == nil {
		lineIDs = []uuid.UUID{}
	}

	return lineIDs, nil
}

func (r *ScheduleRepository) getLineExceptionDays(ctx context.Context, exceptionID uuid.UUID) ([]domain.LineScheduleExceptionDay, error) {
	query := `
		SELECT id, exception_id, day_of_week, is_working_day, shift_start, shift_end, created_at, updated_at
		FROM line_schedule_exception_days
		WHERE exception_id = $1
		ORDER BY day_of_week ASC
	`

	rows, err := r.db.Query(ctx, query, exceptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line exception days: %w", err)
	}
	defer rows.Close()

	var days []domain.LineScheduleExceptionDay
	dayIDs := make([]uuid.UUID, 0, 7)
	for rows.Next() {
		var day domain.LineScheduleExceptionDay
		err := rows.Scan(
			&day.ID, &day.ExceptionID, &day.DayOfWeek, &day.IsWorkingDay,
			&day.ShiftStart, &day.ShiftEnd, &day.CreatedAt, &day.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line exception day: %w", err)
		}
		days = append(days, day)
		dayIDs = append(dayIDs, day.ID)
	}

	if days == nil {
		days = []domain.LineScheduleExceptionDay{}
	}

	// Load breaks
	if len(dayIDs) > 0 {
		breaksMap, err := r.getLineExceptionBreaksForDays(ctx, dayIDs)
		if err != nil {
			return nil, err
		}
		for i := range days {
			if breaks, ok := breaksMap[days[i].ID]; ok {
				days[i].Breaks = breaks
			} else {
				days[i].Breaks = []domain.LineScheduleExceptionBreak{}
			}
		}
	}

	return days, nil
}

func (r *ScheduleRepository) getLineExceptionBreaksForDays(ctx context.Context, dayIDs []uuid.UUID) (map[uuid.UUID][]domain.LineScheduleExceptionBreak, error) {
	query := `
		SELECT id, exception_day_id, name, break_start, break_end, created_at
		FROM line_schedule_exception_breaks
		WHERE exception_day_id = ANY($1)
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get line exception breaks: %w", err)
	}
	defer rows.Close()

	breaksMap := make(map[uuid.UUID][]domain.LineScheduleExceptionBreak)
	for rows.Next() {
		var brk domain.LineScheduleExceptionBreak
		err := rows.Scan(&brk.ID, &brk.ExceptionDayID, &brk.Name, &brk.BreakStart, &brk.BreakEnd, &brk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line exception break: %w", err)
		}
		breaksMap[brk.ExceptionDayID] = append(breaksMap[brk.ExceptionDayID], brk)
	}

	return breaksMap, nil
}

// UpdateLineException updates a line-specific exception
func (r *ScheduleRepository) UpdateLineException(ctx context.Context, exceptionID uuid.UUID, req domain.UpdateLineScheduleExceptionRequest) (*domain.LineScheduleException, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update exception
	query := `
		UPDATE line_schedule_exceptions
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    start_date = COALESCE($4, start_date),
		    end_date = COALESCE($5, end_date),
		    updated_at = now()
		WHERE id = $1
		RETURNING id, schedule_id, name, description, start_date, end_date, created_at, updated_at
	`

	var e domain.LineScheduleException
	err = tx.QueryRow(ctx, query, exceptionID, req.Name, req.Description, req.StartDate, req.EndDate).Scan(
		&e.ID, &e.ScheduleID, &e.Name, &e.Description, &e.StartDate, &e.EndDate, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrLineExceptionNotFound
		}
		return nil, fmt.Errorf("failed to update line exception: %w", err)
	}

	// Update line associations if provided
	if len(req.LineIDs) > 0 {
		_, err = tx.Exec(ctx, `DELETE FROM line_schedule_exception_lines WHERE exception_id = $1`, exceptionID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing line associations: %w", err)
		}

		lineQuery := `INSERT INTO line_schedule_exception_lines (exception_id, line_id) VALUES ($1, $2)`
		for _, lineID := range req.LineIDs {
			_, err = tx.Exec(ctx, lineQuery, exceptionID, lineID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate line with exception: %w", err)
			}
		}
		e.LineIDs = req.LineIDs
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Load line IDs if not updated
	if len(e.LineIDs) == 0 {
		lineIDs, err := r.getLineIDsForException(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		e.LineIDs = lineIDs
	}

	// Load days
	days, err := r.getLineExceptionDays(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	e.Days = days

	return &e, nil
}

// DeleteLineException deletes a line-specific exception
func (r *ScheduleRepository) DeleteLineException(ctx context.Context, exceptionID uuid.UUID) error {
	query := `DELETE FROM line_schedule_exceptions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, exceptionID)
	if err != nil {
		return fmt.Errorf("failed to delete line exception: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrLineExceptionNotFound
	}

	return nil
}

// ========== Line Schedule Assignment ==========

// AssignScheduleToLine assigns a schedule to a production line
func (r *ScheduleRepository) AssignScheduleToLine(ctx context.Context, lineID uuid.UUID, scheduleID *uuid.UUID) error {
	query := `
		UPDATE production_lines
		SET schedule_id = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, lineID, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to assign schedule to line: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetLinesForSchedule retrieves all lines assigned to a schedule
func (r *ScheduleRepository) GetLinesForSchedule(ctx context.Context, scheduleID uuid.UUID) ([]domain.ProductionLine, error) {
	query := `
		SELECT id, code, name, description, status, schedule_id, created_at, updated_at
		FROM production_lines
		WHERE schedule_id = $1 AND deleted_at IS NULL
		ORDER BY code ASC
	`

	rows, err := r.db.Query(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lines for schedule: %w", err)
	}
	defer rows.Close()

	var lines []domain.ProductionLine
	for rows.Next() {
		var line domain.ProductionLine
		err := rows.Scan(
			&line.ID, &line.Code, &line.Name, &line.Description,
			&line.Status, &line.ScheduleID, &line.CreatedAt, &line.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line: %w", err)
		}
		lines = append(lines, line)
	}

	if lines == nil {
		lines = []domain.ProductionLine{}
	}

	return lines, nil
}

// ========== Effective Schedule Lookup ==========

// GetEffectiveSchedule computes the effective schedule for a line on a specific date
// Priority: line exception > schedule exception > holiday > base schedule
func (r *ScheduleRepository) GetEffectiveSchedule(ctx context.Context, lineID uuid.UUID, date string) (*domain.EffectiveSchedule, error) {
	// Get line info with schedule
	lineQuery := `
		SELECT pl.id, pl.code, pl.schedule_id, s.name as schedule_name, s.timezone
		FROM production_lines pl
		LEFT JOIN schedules s ON s.id = pl.schedule_id AND s.deleted_at IS NULL
		WHERE pl.id = $1 AND pl.deleted_at IS NULL
	`

	var lineCode string
	var scheduleID *uuid.UUID
	var scheduleName *string
	var timezone *string

	err := r.db.QueryRow(ctx, lineQuery, lineID).Scan(&lineID, &lineCode, &scheduleID, &scheduleName, &timezone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get line: %w", err)
	}

	result := &domain.EffectiveSchedule{
		LineID:       lineID,
		LineCode:     lineCode,
		Date:         date,
		ScheduleID:   scheduleID,
		ScheduleName: scheduleName,
	}

	// No schedule assigned
	if scheduleID == nil {
		result.Source = domain.SourceNoSchedule
		result.IsWorkingDay = false
		result.Breaks = []domain.EffectiveBreak{}
		return result, nil
	}

	// Check for line-specific exception first
	lineExceptionQuery := `
		SELECT lse.id, lse.name, lsed.is_working_day, lsed.shift_start, lsed.shift_end, lsed.id as day_id
		FROM line_schedule_exceptions lse
		JOIN line_schedule_exception_lines lsel ON lsel.exception_id = lse.id
		JOIN line_schedule_exception_days lsed ON lsed.exception_id = lse.id
		WHERE lse.schedule_id = $1
		  AND lsel.line_id = $2
		  AND $3::date BETWEEN lse.start_date AND lse.end_date
		  AND lsed.day_of_week = EXTRACT(DOW FROM $3::date)
		LIMIT 1
	`

	var exceptionID uuid.UUID
	var exceptionName string
	var isWorkingDay bool
	var shiftStart, shiftEnd *domain.TimeOnly
	var dayID uuid.UUID

	err = r.db.QueryRow(ctx, lineExceptionQuery, scheduleID, lineID, date).Scan(
		&exceptionID, &exceptionName, &isWorkingDay, &shiftStart, &shiftEnd, &dayID,
	)
	if err == nil {
		result.Source = domain.SourceLineException
		result.SourceID = &exceptionID
		result.SourceName = &exceptionName
		result.IsWorkingDay = isWorkingDay
		result.ShiftStart = shiftStart
		result.ShiftEnd = shiftEnd

		// Load breaks
		breaks, err := r.getLineExceptionBreaksForDay(ctx, dayID)
		if err != nil {
			return nil, err
		}
		result.Breaks = breaks
		return result, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check line exception: %w", err)
	}

	// Check for schedule-wide exception
	exceptionQuery := `
		SELECT se.id, se.name, sed.is_working_day, sed.shift_start, sed.shift_end, sed.id as day_id
		FROM schedule_exceptions se
		JOIN schedule_exception_days sed ON sed.exception_id = se.id
		WHERE se.schedule_id = $1
		  AND $2::date BETWEEN se.start_date AND se.end_date
		  AND sed.day_of_week = EXTRACT(DOW FROM $2::date)
		LIMIT 1
	`

	err = r.db.QueryRow(ctx, exceptionQuery, scheduleID, date).Scan(
		&exceptionID, &exceptionName, &isWorkingDay, &shiftStart, &shiftEnd, &dayID,
	)
	if err == nil {
		result.Source = domain.SourceScheduleException
		result.SourceID = &exceptionID
		result.SourceName = &exceptionName
		result.IsWorkingDay = isWorkingDay
		result.ShiftStart = shiftStart
		result.ShiftEnd = shiftEnd

		// Load breaks
		breaks, err := r.getScheduleExceptionBreaksForDay(ctx, dayID)
		if err != nil {
			return nil, err
		}
		result.Breaks = breaks
		return result, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check schedule exception: %w", err)
	}

	// Check for holiday
	holidayQuery := `
		SELECT id, name
		FROM schedule_holidays
		WHERE schedule_id = $1 AND holiday_date = $2::date
		LIMIT 1
	`

	var holidayID uuid.UUID
	var holidayName *string

	err = r.db.QueryRow(ctx, holidayQuery, scheduleID, date).Scan(&holidayID, &holidayName)
	if err == nil {
		result.Source = domain.SourceHoliday
		result.SourceID = &holidayID
		result.SourceName = holidayName
		result.IsWorkingDay = false
		result.Breaks = []domain.EffectiveBreak{}
		return result, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check holiday: %w", err)
	}

	// Use base schedule
	baseQuery := `
		SELECT sd.id, sd.is_working_day, sd.shift_start, sd.shift_end
		FROM schedule_days sd
		WHERE sd.schedule_id = $1
		  AND sd.day_of_week = EXTRACT(DOW FROM $2::date)
		LIMIT 1
	`

	err = r.db.QueryRow(ctx, baseQuery, scheduleID, date).Scan(&dayID, &isWorkingDay, &shiftStart, &shiftEnd)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("schedule day not found for date %s", date)
		}
		return nil, fmt.Errorf("failed to get base schedule: %w", err)
	}

	result.Source = domain.SourceBase
	result.SourceID = &dayID
	result.IsWorkingDay = isWorkingDay
	result.ShiftStart = shiftStart
	result.ShiftEnd = shiftEnd

	// Load breaks
	breaks, err := r.getBaseScheduleBreaksForDay(ctx, dayID)
	if err != nil {
		return nil, err
	}
	result.Breaks = breaks

	return result, nil
}

func (r *ScheduleRepository) getBaseScheduleBreaksForDay(ctx context.Context, dayID uuid.UUID) ([]domain.EffectiveBreak, error) {
	query := `
		SELECT name, break_start, break_end
		FROM schedule_breaks
		WHERE schedule_day_id = $1
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get breaks: %w", err)
	}
	defer rows.Close()

	var breaks []domain.EffectiveBreak
	for rows.Next() {
		var brk domain.EffectiveBreak
		err := rows.Scan(&brk.Name, &brk.BreakStart, &brk.BreakEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to scan break: %w", err)
		}
		breaks = append(breaks, brk)
	}

	if breaks == nil {
		breaks = []domain.EffectiveBreak{}
	}

	return breaks, nil
}

func (r *ScheduleRepository) getScheduleExceptionBreaksForDay(ctx context.Context, dayID uuid.UUID) ([]domain.EffectiveBreak, error) {
	query := `
		SELECT name, break_start, break_end
		FROM schedule_exception_breaks
		WHERE exception_day_id = $1
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get exception breaks: %w", err)
	}
	defer rows.Close()

	var breaks []domain.EffectiveBreak
	for rows.Next() {
		var brk domain.EffectiveBreak
		err := rows.Scan(&brk.Name, &brk.BreakStart, &brk.BreakEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to scan exception break: %w", err)
		}
		breaks = append(breaks, brk)
	}

	if breaks == nil {
		breaks = []domain.EffectiveBreak{}
	}

	return breaks, nil
}

func (r *ScheduleRepository) getLineExceptionBreaksForDay(ctx context.Context, dayID uuid.UUID) ([]domain.EffectiveBreak, error) {
	query := `
		SELECT name, break_start, break_end
		FROM line_schedule_exception_breaks
		WHERE exception_day_id = $1
		ORDER BY break_start ASC
	`

	rows, err := r.db.Query(ctx, query, dayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line exception breaks: %w", err)
	}
	defer rows.Close()

	var breaks []domain.EffectiveBreak
	for rows.Next() {
		var brk domain.EffectiveBreak
		err := rows.Scan(&brk.Name, &brk.BreakStart, &brk.BreakEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line exception break: %w", err)
		}
		breaks = append(breaks, brk)
	}

	if breaks == nil {
		breaks = []domain.EffectiveBreak{}
	}

	return breaks, nil
}

// GetEffectiveScheduleRange computes effective schedules for a line over a date range
func (r *ScheduleRepository) GetEffectiveScheduleRange(ctx context.Context, lineID uuid.UUID, startDate, endDate string) ([]domain.EffectiveSchedule, error) {
	// Generate all dates in range
	query := `
		SELECT d::date
		FROM generate_series($1::date, $2::date, '1 day'::interval) AS d
		ORDER BY d
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate date range: %w", err)
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, fmt.Errorf("failed to scan date: %w", err)
		}
		dates = append(dates, date)
	}

	// Get effective schedule for each date
	var results []domain.EffectiveSchedule
	for _, date := range dates {
		effective, err := r.GetEffectiveSchedule(ctx, lineID, date)
		if err != nil {
			return nil, err
		}
		results = append(results, *effective)
	}

	return results, nil
}
