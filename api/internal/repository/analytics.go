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

// AnalyticsRepository handles database operations for analytics queries
type AnalyticsRepository struct {
	db *pgxpool.Pool
}

// NewAnalyticsRepository creates a new AnalyticsRepository
func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetAggregateMetrics calculates aggregate metrics across multiple lines
func (r *AnalyticsRepository) GetAggregateMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) (*domain.AggregateMetrics, error) {
	whereClause, filterArgs := r.buildWhereClause(query)

	// Build complete args slice with time params first, then filter params
	args := []interface{}{query.EndTime, query.StartTime}
	args = append(args, filterArgs...)

	sqlQuery := fmt.Sprintf(`
		WITH line_filter AS (
			SELECT DISTINCT pl.id
			FROM production_lines pl
			LEFT JOIN production_line_labels pll ON pl.id = pll.line_id
			WHERE pl.deleted_at IS NULL
			%s
		),
		status_intervals AS (
			SELECT
				log.line_id,
				log.time as start_time,
				LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) as end_time,
				log.new_status,
				EXTRACT(EPOCH FROM (
					LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) - log.time
				)) / 3600.0 as duration_hours
			FROM production_line_status_log log
			INNER JOIN line_filter lf ON log.line_id = lf.id
			WHERE log.time >= $2 AND log.time <= $1
		),
		error_periods AS (
			SELECT
				line_id,
				start_time,
				end_time,
				duration_hours
			FROM status_intervals
			WHERE new_status IN ('error', 'maintenance')
		)
		SELECT
			COUNT(DISTINCT si.line_id) as total_lines,
			COALESCE(SUM(CASE WHEN si.new_status = 'on' THEN si.duration_hours ELSE 0 END), 0) as total_uptime,
			COALESCE(SUM(CASE WHEN si.new_status = 'off' THEN si.duration_hours ELSE 0 END), 0) as total_downtime,
			COALESCE(SUM(CASE WHEN si.new_status = 'maintenance' THEN si.duration_hours ELSE 0 END), 0) as total_maintenance,
			COALESCE(AVG(ep.duration_hours), 0) as mttr,
			COALESCE(COUNT(DISTINCT CASE WHEN si.new_status IN ('error', 'maintenance') THEN si.line_id || '-' || si.start_time::text END), 0) as total_interruptions
		FROM status_intervals si
		LEFT JOIN error_periods ep ON si.line_id = ep.line_id
	`, whereClause)

	var metrics domain.AggregateMetrics
	var totalUptime, totalDowntime, totalMaintenance float64

	err := r.db.QueryRow(ctx, sqlQuery, args...).Scan(
		&metrics.TotalLines,
		&totalUptime,
		&totalDowntime,
		&totalMaintenance,
		&metrics.MTTR,
		&metrics.TotalInterruptions,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to calculate aggregate metrics: %w", err)
	}

	metrics.TotalUptime = totalUptime
	metrics.TotalDowntime = totalDowntime
	metrics.TotalMaintenance = totalMaintenance

	// Calculate total hours and percentages
	totalHours := totalUptime + totalDowntime + totalMaintenance
	if totalHours > 0 {
		metrics.AverageUptime = (totalUptime / totalHours) * 100
		metrics.StatusDistribution = map[domain.Status]float64{
			domain.StatusOn:          (totalUptime / totalHours) * 100,
			domain.StatusOff:         (totalDowntime / totalHours) * 100,
			domain.StatusMaintenance: (totalMaintenance / totalHours) * 100,
		}
	} else {
		metrics.StatusDistribution = map[domain.Status]float64{}
	}

	metrics.TimeRange = domain.TimeRange{
		Start: *query.StartTime,
		End:   *query.EndTime,
	}

	return &metrics, nil
}

// GetLineMetrics calculates metrics for each production line
func (r *AnalyticsRepository) GetLineMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) ([]domain.LineMetrics, error) {
	whereClause, filterArgs := r.buildWhereClause(query)

	// Build complete args slice
	args := []interface{}{query.EndTime, query.StartTime}
	args = append(args, filterArgs...)

	sqlQuery := fmt.Sprintf(`
		WITH line_filter AS (
			SELECT DISTINCT pl.id, pl.code, pl.name, pl.status
			FROM production_lines pl
			LEFT JOIN production_line_labels pll ON pl.id = pll.line_id
			WHERE pl.deleted_at IS NULL
			%s
		),
		status_intervals AS (
			SELECT
				log.line_id,
				log.time as start_time,
				LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) as end_time,
				log.new_status,
				EXTRACT(EPOCH FROM (
					LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) - log.time
				)) / 3600.0 as duration_hours
			FROM production_line_status_log log
			INNER JOIN line_filter lf ON log.line_id = lf.id
			WHERE log.time >= $2 AND log.time <= $1
		)
		SELECT
			lf.id,
			lf.code,
			lf.name,
			lf.status,
			COALESCE(SUM(CASE WHEN si.new_status = 'on' THEN si.duration_hours ELSE 0 END), 0) as uptime_hours,
			COALESCE(SUM(CASE WHEN si.new_status = 'off' THEN si.duration_hours ELSE 0 END), 0) as downtime_hours,
			COALESCE(SUM(CASE WHEN si.new_status = 'maintenance' THEN si.duration_hours ELSE 0 END), 0) as maintenance_hours,
			COALESCE(SUM(CASE WHEN si.new_status = 'error' THEN si.duration_hours ELSE 0 END), 0) as error_hours,
			COALESCE(COUNT(CASE WHEN si.new_status IN ('error', 'maintenance') THEN 1 END), 0) as interruption_count
		FROM line_filter lf
		LEFT JOIN status_intervals si ON lf.id = si.line_id
		GROUP BY lf.id, lf.code, lf.name, lf.status
		ORDER BY lf.code
	`, whereClause)

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get line metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.LineMetrics
	labelRepo := NewLabelRepository(r.db)

	for rows.Next() {
		var m domain.LineMetrics
		var uptime, downtime, maintenance, errorTime float64

		err := rows.Scan(
			&m.LineID, &m.LineCode, &m.LineName, &m.CurrentStatus,
			&uptime, &downtime, &maintenance, &errorTime,
			&m.InterruptionCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line metrics: %w", err)
		}

		m.UptimeHours = uptime
		m.DowntimeHours = downtime
		m.MaintenanceHours = maintenance
		m.ErrorHours = errorTime

		totalHours := uptime + downtime + maintenance + errorTime
		if totalHours > 0 {
			m.UptimePercentage = (uptime / totalHours) * 100
			m.StatusDistribution = map[domain.Status]float64{
				domain.StatusOn:          (uptime / totalHours) * 100,
				domain.StatusOff:         (downtime / totalHours) * 100,
				domain.StatusMaintenance: (maintenance / totalHours) * 100,
				domain.StatusError:       (errorTime / totalHours) * 100,
			}
		} else {
			m.StatusDistribution = map[domain.Status]float64{}
		}

		// Calculate MTTR for this line
		if m.InterruptionCount > 0 {
			m.MTTR = (maintenance + errorTime) / float64(m.InterruptionCount)
		}

		// Load labels for this line
		labels, err := labelRepo.GetLabelsForLine(ctx, m.LineID)
		if err == nil {
			m.Labels = labels
		} else {
			m.Labels = []domain.Label{}
		}

		metrics = append(metrics, m)
	}

	if metrics == nil {
		metrics = []domain.LineMetrics{}
	}

	return metrics, nil
}

// GetLabelMetrics calculates metrics grouped by label
func (r *AnalyticsRepository) GetLabelMetrics(
	ctx context.Context,
	query domain.AnalyticsQuery,
) ([]domain.LabelMetrics, error) {
	sqlQuery := `
		WITH status_intervals AS (
			SELECT
				pll.label_id,
				log.line_id,
				log.time as start_time,
				LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) as end_time,
				log.new_status,
				EXTRACT(EPOCH FROM (
					LEAD(log.time, 1, $1) OVER (PARTITION BY log.line_id ORDER BY log.time) - log.time
				)) / 3600.0 as duration_hours
			FROM production_line_status_log log
			INNER JOIN production_lines pl ON log.line_id = pl.id
			INNER JOIN production_line_labels pll ON pl.id = pll.line_id
			WHERE log.time >= $2 AND log.time <= $1
			  AND pl.deleted_at IS NULL
		)
		SELECT
			l.id,
			l.name,
			l.color,
			l.description,
			l.created_at,
			l.updated_at,
			COUNT(DISTINCT si.line_id) as line_count,
			COALESCE(SUM(CASE WHEN si.new_status = 'on' THEN si.duration_hours ELSE 0 END), 0) as total_uptime,
			COALESCE(COUNT(CASE WHEN si.new_status IN ('error', 'maintenance') THEN 1 END), 0) as total_interruptions
		FROM labels l
		LEFT JOIN status_intervals si ON l.id = si.label_id
		GROUP BY l.id, l.name, l.color, l.description, l.created_at, l.updated_at
		ORDER BY l.name
	`

	rows, err := r.db.Query(ctx, sqlQuery, query.EndTime, query.StartTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get label metrics: %w", err)
	}
	defer rows.Close()

	var metrics []domain.LabelMetrics
	for rows.Next() {
		var m domain.LabelMetrics

		err := rows.Scan(
			&m.Label.ID, &m.Label.Name, &m.Label.Color, &m.Label.Description,
			&m.Label.CreatedAt, &m.Label.UpdatedAt,
			&m.LineCount, &m.TotalUptime, &m.TotalInterruptions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label metrics: %w", err)
		}

		// Calculate average uptime percentage
		if m.LineCount > 0 && m.TotalUptime > 0 {
			m.AverageUptime = (m.TotalUptime / float64(m.LineCount)) * 100
		}

		if m.TotalInterruptions > 0 {
			m.MTTR = m.TotalUptime / float64(m.TotalInterruptions)
		}

		metrics = append(metrics, m)
	}

	if metrics == nil {
		metrics = []domain.LabelMetrics{}
	}

	return metrics, nil
}

// GetDailyKPIs retrieves KPIs for each 24-hour period in the range
func (r *AnalyticsRepository) GetDailyKPIs(
	ctx context.Context,
	lineID uuid.UUID,
	startTime, endTime time.Time,
) ([]domain.DailyKPI, error) {
	sqlQuery := `
		WITH daily_intervals AS (
			SELECT
				DATE(log.time) as date,
				log.time as start_time,
				LEAD(log.time, 1, $4) OVER (PARTITION BY log.line_id ORDER BY log.time) as end_time,
				log.new_status,
				EXTRACT(EPOCH FROM (
					LEAD(log.time, 1, $4) OVER (PARTITION BY log.line_id ORDER BY log.time) - log.time
				)) / 3600.0 as duration_hours
			FROM production_line_status_log log
			WHERE log.line_id = $1
			  AND log.time >= $2
			  AND log.time <= $4
		)
		SELECT
			date,
			COALESCE(SUM(CASE WHEN new_status = 'on' THEN duration_hours ELSE 0 END), 0) as uptime_hours,
			COALESCE(SUM(CASE WHEN new_status = 'maintenance' THEN duration_hours ELSE 0 END), 0) as maintenance_hours,
			COALESCE(COUNT(CASE WHEN new_status IN ('error', 'maintenance') THEN 1 END), 0) as interruption_count,
			COALESCE(AVG(CASE WHEN new_status IN ('error', 'maintenance') THEN duration_hours END), 0) as mttr
		FROM daily_intervals
		GROUP BY date
		ORDER BY date
	`

	rows, err := r.db.Query(ctx, sqlQuery, lineID, startTime, endTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily KPIs: %w", err)
	}
	defer rows.Close()

	var kpis []domain.DailyKPI
	for rows.Next() {
		var kpi domain.DailyKPI
		var uptime, maintenance, mttr float64

		err := rows.Scan(&kpi.Date, &uptime, &maintenance, &kpi.InterruptionCount, &mttr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily KPI: %w", err)
		}

		kpi.UptimeHours = uptime
		kpi.MaintenanceHours = maintenance
		kpi.MTTR = mttr

		totalHours := 24.0
		kpi.UptimePercentage = (uptime / totalHours) * 100

		kpis = append(kpis, kpi)
	}

	if kpis == nil {
		kpis = []domain.DailyKPI{}
	}

	return kpis, nil
}

// buildWhereClause constructs WHERE clause based on query filters
func (r *AnalyticsRepository) buildWhereClause(query domain.AnalyticsQuery) (string, []interface{}) {
	var conditions []string
	args := []interface{}{}
	argPos := 3 // Starting position (after $1=endTime, $2=startTime)

	if len(query.LabelIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("AND pll.label_id = ANY($%d)", argPos))
		args = append(args, query.LabelIDs)
		argPos++
	}

	if len(query.LineIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("AND pl.id = ANY($%d)", argPos))
		args = append(args, query.LineIDs)
		argPos++
	}

	whereClause := strings.Join(conditions, " ")
	return whereClause, args
}
