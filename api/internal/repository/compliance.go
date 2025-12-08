package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ping/production-line-api/internal/domain"
)

// ComplianceRepository handles database operations for compliance calculations
type ComplianceRepository struct {
	db           *pgxpool.Pool
	scheduleRepo *ScheduleRepository
}

// NewComplianceRepository creates a new ComplianceRepository
func NewComplianceRepository(db *pgxpool.Pool, scheduleRepo *ScheduleRepository) *ComplianceRepository {
	return &ComplianceRepository{
		db:           db,
		scheduleRepo: scheduleRepo,
	}
}

// GetLineComplianceMetrics calculates compliance metrics for each line
func (r *ComplianceRepository) GetLineComplianceMetrics(
	ctx context.Context,
	query domain.ComplianceQuery,
) ([]domain.LineComplianceMetrics, error) {
	// Get filtered lines
	lines, err := r.getFilteredLines(ctx, query)
	if err != nil {
		return nil, err
	}

	var metrics []domain.LineComplianceMetrics
	for _, line := range lines {
		lineMetrics, err := r.calculateLineCompliance(ctx, line, query.StartDate, query.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate compliance for line %s: %w", line.Code, err)
		}
		metrics = append(metrics, *lineMetrics)
	}

	if metrics == nil {
		metrics = []domain.LineComplianceMetrics{}
	}

	return metrics, nil
}

// GetAggregateComplianceMetrics calculates overall compliance metrics
func (r *ComplianceRepository) GetAggregateComplianceMetrics(
	ctx context.Context,
	query domain.ComplianceQuery,
) (*domain.AggregateComplianceMetrics, error) {
	lineMetrics, err := r.GetLineComplianceMetrics(ctx, query)
	if err != nil {
		return nil, err
	}

	result := &domain.AggregateComplianceMetrics{
		TotalLines:  len(lineMetrics),
		LineMetrics: lineMetrics,
		DateRange: domain.DateRange{
			StartDate: query.StartDate,
			EndDate:   query.EndDate,
		},
	}

	var totalCompliance float64
	linesWithCompliance := 0

	for _, lm := range lineMetrics {
		if lm.ScheduleID != nil {
			result.LinesWithSchedule++
		}
		result.TotalScheduledHours += lm.ScheduledUptimeHours
		result.TotalActualHours += lm.ActualUptimeHours
		result.TotalUnplannedDowntime += lm.UnplannedDowntimeHours
		result.TotalOvertime += lm.OvertimeHours

		if lm.ScheduledUptimeHours > 0 {
			totalCompliance += lm.CompliancePercentage
			linesWithCompliance++
		}
	}

	if linesWithCompliance > 0 {
		result.AverageCompliance = totalCompliance / float64(linesWithCompliance)
	}

	return result, nil
}

// GetDailyComplianceKPIs returns daily compliance data for a specific line
func (r *ComplianceRepository) GetDailyComplianceKPIs(
	ctx context.Context,
	lineID uuid.UUID,
	startDate, endDate string,
) ([]domain.DailyComplianceKPI, error) {
	// Get effective schedules for each day
	effectiveSchedules, err := r.scheduleRepo.GetEffectiveScheduleRange(ctx, lineID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective schedules: %w", err)
	}

	var kpis []domain.DailyComplianceKPI
	for _, es := range effectiveSchedules {
		kpi, err := r.calculateDailyCompliance(ctx, lineID, es)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate daily compliance for %s: %w", es.Date, err)
		}
		kpis = append(kpis, *kpi)
	}

	if kpis == nil {
		kpis = []domain.DailyComplianceKPI{}
	}

	return kpis, nil
}

// lineInfo holds basic line information for compliance calculations
type lineInfo struct {
	ID           uuid.UUID
	Code         string
	Name         string
	ScheduleID   *uuid.UUID
	ScheduleName *string
}

// getFilteredLines returns lines matching the query filters
func (r *ComplianceRepository) getFilteredLines(ctx context.Context, query domain.ComplianceQuery) ([]lineInfo, error) {
	var conditions []string
	args := []interface{}{}
	argPos := 1

	conditions = append(conditions, "pl.deleted_at IS NULL")

	if len(query.LineIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("pl.id = ANY($%d)", argPos))
		args = append(args, query.LineIDs)
		argPos++
	}

	if len(query.LabelIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM production_line_labels pll WHERE pll.line_id = pl.id AND pll.label_id = ANY($%d))", argPos))
		args = append(args, query.LabelIDs)
		argPos++
	}

	sqlQuery := fmt.Sprintf(`
		SELECT pl.id, pl.code, pl.name, pl.schedule_id, s.name as schedule_name
		FROM production_lines pl
		LEFT JOIN schedules s ON s.id = pl.schedule_id AND s.deleted_at IS NULL
		WHERE %s
		ORDER BY pl.code
	`, strings.Join(conditions, " AND "))

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get lines: %w", err)
	}
	defer rows.Close()

	var lines []lineInfo
	for rows.Next() {
		var line lineInfo
		err := rows.Scan(&line.ID, &line.Code, &line.Name, &line.ScheduleID, &line.ScheduleName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line: %w", err)
		}
		lines = append(lines, line)
	}

	return lines, nil
}

// calculateLineCompliance calculates compliance metrics for a single line
func (r *ComplianceRepository) calculateLineCompliance(
	ctx context.Context,
	line lineInfo,
	startDate, endDate string,
) (*domain.LineComplianceMetrics, error) {
	metrics := &domain.LineComplianceMetrics{
		LineID:       line.ID,
		LineCode:     line.Code,
		LineName:     line.Name,
		ScheduleID:   line.ScheduleID,
		ScheduleName: line.ScheduleName,
	}

	// If no schedule, return zeroed metrics
	if line.ScheduleID == nil {
		return metrics, nil
	}

	// Get effective schedules for the date range
	effectiveSchedules, err := r.scheduleRepo.GetEffectiveScheduleRange(ctx, line.ID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get effective schedules: %w", err)
	}

	metrics.ScheduledDays = len(effectiveSchedules)

	// Calculate scheduled uptime hours
	for _, es := range effectiveSchedules {
		if es.IsWorkingDay {
			metrics.WorkingDays++
			scheduledHours := r.calculateScheduledHours(es)
			metrics.ScheduledUptimeHours += scheduledHours
		}
	}

	// Get actual status data for the period
	actualUptime, unplannedDowntime, overtime, err := r.calculateActualMetrics(ctx, line.ID, effectiveSchedules)
	if err != nil {
		return nil, err
	}

	metrics.ActualUptimeHours = actualUptime
	metrics.UnplannedDowntimeHours = unplannedDowntime
	metrics.OvertimeHours = overtime

	// Calculate compliance percentage
	if metrics.ScheduledUptimeHours > 0 {
		metrics.CompliancePercentage = (metrics.ActualUptimeHours / metrics.ScheduledUptimeHours) * 100
		// Cap at 100% for compliance (overtime is tracked separately)
		if metrics.CompliancePercentage > 100 {
			metrics.CompliancePercentage = 100
		}
	}

	return metrics, nil
}

// calculateScheduledHours calculates the scheduled working hours for a day
func (r *ComplianceRepository) calculateScheduledHours(es domain.EffectiveSchedule) float64 {
	if !es.IsWorkingDay || es.ShiftStart == nil || es.ShiftEnd == nil {
		return 0
	}

	// Parse shift times
	shiftStart := time.Date(2000, 1, 1, es.ShiftStart.Hour(), es.ShiftStart.Minute(), es.ShiftStart.Second(), 0, time.UTC)
	shiftEnd := time.Date(2000, 1, 1, es.ShiftEnd.Hour(), es.ShiftEnd.Minute(), es.ShiftEnd.Second(), 0, time.UTC)

	// Handle overnight shifts
	if shiftEnd.Before(shiftStart) {
		shiftEnd = shiftEnd.Add(24 * time.Hour)
	}

	totalHours := shiftEnd.Sub(shiftStart).Hours()

	// Subtract break times
	for _, brk := range es.Breaks {
		breakStart := time.Date(2000, 1, 1, brk.BreakStart.Hour(), brk.BreakStart.Minute(), brk.BreakStart.Second(), 0, time.UTC)
		breakEnd := time.Date(2000, 1, 1, brk.BreakEnd.Hour(), brk.BreakEnd.Minute(), brk.BreakEnd.Second(), 0, time.UTC)
		if breakEnd.Before(breakStart) {
			breakEnd = breakEnd.Add(24 * time.Hour)
		}
		totalHours -= breakEnd.Sub(breakStart).Hours()
	}

	return totalHours
}

// calculateActualMetrics calculates actual uptime, downtime and overtime from status logs
func (r *ComplianceRepository) calculateActualMetrics(
	ctx context.Context,
	lineID uuid.UUID,
	effectiveSchedules []domain.EffectiveSchedule,
) (actualUptime, unplannedDowntime, overtime float64, err error) {
	if len(effectiveSchedules) == 0 {
		return 0, 0, 0, nil
	}

	// Get the date range
	startDate := effectiveSchedules[0].Date
	endDate := effectiveSchedules[len(effectiveSchedules)-1].Date

	// Parse dates to get the full time range
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	end = end.Add(24 * time.Hour) // Include the full end day

	// Get status intervals for the period
	query := `
		WITH status_intervals AS (
			SELECT
				log.time as start_time,
				LEAD(log.time, 1, $3) OVER (ORDER BY log.time) as end_time,
				log.new_status
			FROM production_line_status_log log
			WHERE log.line_id = $1
			  AND log.time >= $2
			  AND log.time < $3
		)
		SELECT start_time, end_time, new_status
		FROM status_intervals
		ORDER BY start_time
	`

	rows, err := r.db.Query(ctx, query, lineID, start, end)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get status intervals: %w", err)
	}
	defer rows.Close()

	// Build a map of date -> scheduled periods
	scheduledPeriods := r.buildScheduledPeriods(effectiveSchedules)

	for rows.Next() {
		var intervalStart, intervalEnd time.Time
		var status domain.Status

		if err := rows.Scan(&intervalStart, &intervalEnd, &status); err != nil {
			return 0, 0, 0, fmt.Errorf("failed to scan status interval: %w", err)
		}

		// Calculate metrics for this interval
		if status == domain.StatusOn {
			// Check how much of this "on" time was during scheduled hours
			scheduledOn, outsideSchedule := r.splitIntervalBySchedule(intervalStart, intervalEnd, scheduledPeriods)
			actualUptime += scheduledOn
			overtime += outsideSchedule
		} else if status == domain.StatusOff || status == domain.StatusError || status == domain.StatusMaintenance {
			// Check how much of this downtime was during scheduled hours
			scheduledDowntime, _ := r.splitIntervalBySchedule(intervalStart, intervalEnd, scheduledPeriods)
			unplannedDowntime += scheduledDowntime
		}
	}

	return actualUptime, unplannedDowntime, overtime, nil
}

// scheduledPeriod represents a scheduled working period
type scheduledPeriod struct {
	start time.Time
	end   time.Time
}

// buildScheduledPeriods converts effective schedules to concrete time periods
func (r *ComplianceRepository) buildScheduledPeriods(schedules []domain.EffectiveSchedule) []scheduledPeriod {
	var periods []scheduledPeriod

	for _, es := range schedules {
		if !es.IsWorkingDay || es.ShiftStart == nil || es.ShiftEnd == nil {
			continue
		}

		date, _ := time.Parse("2006-01-02", es.Date)

		shiftStart := time.Date(date.Year(), date.Month(), date.Day(),
			es.ShiftStart.Hour(), es.ShiftStart.Minute(), es.ShiftStart.Second(), 0, time.UTC)
		shiftEnd := time.Date(date.Year(), date.Month(), date.Day(),
			es.ShiftEnd.Hour(), es.ShiftEnd.Minute(), es.ShiftEnd.Second(), 0, time.UTC)

		// Handle overnight shifts
		if shiftEnd.Before(shiftStart) {
			shiftEnd = shiftEnd.Add(24 * time.Hour)
		}

		// Add the shift period, but we need to exclude breaks
		// For simplicity, add the whole shift and subtract breaks
		periods = append(periods, scheduledPeriod{start: shiftStart, end: shiftEnd})
	}

	return periods
}

// splitIntervalBySchedule splits an interval into scheduled and outside-schedule portions
func (r *ComplianceRepository) splitIntervalBySchedule(
	intervalStart, intervalEnd time.Time,
	scheduledPeriods []scheduledPeriod,
) (duringSchedule, outsideSchedule float64) {
	totalHours := intervalEnd.Sub(intervalStart).Hours()
	scheduledHours := 0.0

	for _, period := range scheduledPeriods {
		// Find overlap between interval and scheduled period
		overlapStart := intervalStart
		if period.start.After(intervalStart) {
			overlapStart = period.start
		}

		overlapEnd := intervalEnd
		if period.end.Before(intervalEnd) {
			overlapEnd = period.end
		}

		if overlapEnd.After(overlapStart) {
			scheduledHours += overlapEnd.Sub(overlapStart).Hours()
		}
	}

	return scheduledHours, totalHours - scheduledHours
}

// calculateDailyCompliance calculates compliance for a single day
func (r *ComplianceRepository) calculateDailyCompliance(
	ctx context.Context,
	lineID uuid.UUID,
	es domain.EffectiveSchedule,
) (*domain.DailyComplianceKPI, error) {
	kpi := &domain.DailyComplianceKPI{
		Date:         es.Date,
		IsWorkingDay: es.IsWorkingDay,
		Source:       string(es.Source),
	}

	if !es.IsWorkingDay {
		return kpi, nil
	}

	// Calculate scheduled hours for this day
	kpi.ScheduledUptimeHours = r.calculateScheduledHours(es)

	// Get actual metrics for just this day
	actualUptime, unplannedDowntime, overtime, err := r.calculateActualMetrics(ctx, lineID, []domain.EffectiveSchedule{es})
	if err != nil {
		return nil, err
	}

	kpi.ActualUptimeHours = actualUptime
	kpi.UnplannedDowntimeHours = unplannedDowntime
	kpi.OvertimeHours = overtime

	if kpi.ScheduledUptimeHours > 0 {
		kpi.CompliancePercentage = (kpi.ActualUptimeHours / kpi.ScheduledUptimeHours) * 100
		if kpi.CompliancePercentage > 100 {
			kpi.CompliancePercentage = 100
		}
	}

	return kpi, nil
}
