package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ping/production-line-api/internal/domain"
)

// StatusLogRepository handles database operations for status change logs
type StatusLogRepository struct {
	db *pgxpool.Pool
}

// NewStatusLogRepository creates a new StatusLogRepository
func NewStatusLogRepository(db *pgxpool.Pool) *StatusLogRepository {
	return &StatusLogRepository{db: db}
}

// Insert records a status change in the audit log
func (r *StatusLogRepository) Insert(ctx context.Context, change *domain.StatusChange) error {
	// Convert source_detail to JSONB
	var sourceDetailJSON []byte
	var err error
	if change.SourceDetail != nil {
		sourceDetailJSON, err = json.Marshal(change.SourceDetail)
		if err != nil {
			return fmt.Errorf("failed to marshal source_detail: %w", err)
		}
	}

	query := `
		INSERT INTO production_line_status_log
			(time, line_id, line_code, old_status, new_status, source, source_detail)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Exec(
		ctx,
		query,
		change.Time,
		change.LineID,
		change.LineCode,
		change.OldStatus,
		change.NewStatus,
		change.Source,
		sourceDetailJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert status log: %w", err)
	}

	return nil
}

// GetHistory retrieves status change history for a production line
func (r *StatusLogRepository) GetHistory(ctx context.Context, lineID uuid.UUID, limit int) ([]domain.StatusChange, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}

	query := `
		SELECT time, line_id, line_code, old_status, new_status, source, source_detail
		FROM production_line_status_log
		WHERE line_id = $1
		ORDER BY time DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, lineID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history: %w", err)
	}
	defer rows.Close()

	var history []domain.StatusChange
	for rows.Next() {
		var change domain.StatusChange
		var sourceDetailJSON []byte

		err := rows.Scan(
			&change.Time,
			&change.LineID,
			&change.LineCode,
			&change.OldStatus,
			&change.NewStatus,
			&change.Source,
			&sourceDetailJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status change: %w", err)
		}

		// Unmarshal source_detail if present
		if sourceDetailJSON != nil {
			var detail interface{}
			if err := json.Unmarshal(sourceDetailJSON, &detail); err != nil {
				return nil, fmt.Errorf("failed to unmarshal source_detail: %w", err)
			}
			change.SourceDetail = detail
		}

		history = append(history, change)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status history: %w", err)
	}

	// Return empty slice instead of nil for better JSON marshaling
	if history == nil {
		history = []domain.StatusChange{}
	}

	return history, nil
}

// GetHistoryByTimeRange retrieves status changes within a time range
func (r *StatusLogRepository) GetHistoryByTimeRange(
	ctx context.Context,
	lineID uuid.UUID,
	startTime, endTime interface{},
	limit int,
) ([]domain.StatusChange, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT time, line_id, line_code, old_status, new_status, source, source_detail
		FROM production_line_status_log
		WHERE line_id = $1
		  AND time >= $2
		  AND time <= $3
		ORDER BY time DESC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, lineID, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get status history by time range: %w", err)
	}
	defer rows.Close()

	var history []domain.StatusChange
	for rows.Next() {
		var change domain.StatusChange
		var sourceDetailJSON []byte

		err := rows.Scan(
			&change.Time,
			&change.LineID,
			&change.LineCode,
			&change.OldStatus,
			&change.NewStatus,
			&change.Source,
			&sourceDetailJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status change: %w", err)
		}

		if sourceDetailJSON != nil {
			var detail interface{}
			if err := json.Unmarshal(sourceDetailJSON, &detail); err != nil {
				return nil, fmt.Errorf("failed to unmarshal source_detail: %w", err)
			}
			change.SourceDetail = detail
		}

		history = append(history, change)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating status history: %w", err)
	}

	if history == nil {
		history = []domain.StatusChange{}
	}

	return history, nil
}
