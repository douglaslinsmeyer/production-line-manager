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

// LineRepository handles database operations for production lines
type LineRepository struct {
	db *pgxpool.Pool
}

// NewLineRepository creates a new LineRepository
func NewLineRepository(db *pgxpool.Pool) *LineRepository {
	return &LineRepository{db: db}
}

// Create creates a new production line
func (r *LineRepository) Create(ctx context.Context, req domain.CreateLineRequest) (*domain.ProductionLine, error) {
	query := `
		INSERT INTO production_lines (code, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, code, name, description, status, created_at, updated_at
	`

	var line domain.ProductionLine
	err := r.db.QueryRow(ctx, query, req.Code, req.Name, req.Description).Scan(
		&line.ID,
		&line.Code,
		&line.Name,
		&line.Description,
		&line.Status,
		&line.CreatedAt,
		&line.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, domain.ErrCodeExists
		}
		return nil, fmt.Errorf("failed to create production line: %w", err)
	}

	return &line, nil
}

// GetByID retrieves a production line by its ID
func (r *LineRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProductionLine, error) {
	query := `
		SELECT id, code, name, description, status, created_at, updated_at, deleted_at
		FROM production_lines
		WHERE id = $1 AND deleted_at IS NULL
	`

	var line domain.ProductionLine
	err := r.db.QueryRow(ctx, query, id).Scan(
		&line.ID,
		&line.Code,
		&line.Name,
		&line.Description,
		&line.Status,
		&line.CreatedAt,
		&line.UpdatedAt,
		&line.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get production line by id: %w", err)
	}

	// Load labels for this line
	labels, err := r.getLabelsForLine(ctx, line.ID)
	if err != nil {
		return nil, err
	}
	line.Labels = labels

	return &line, nil
}

// GetByCode retrieves a production line by its code
func (r *LineRepository) GetByCode(ctx context.Context, code string) (*domain.ProductionLine, error) {
	query := `
		SELECT id, code, name, description, status, created_at, updated_at, deleted_at
		FROM production_lines
		WHERE code = $1 AND deleted_at IS NULL
	`

	var line domain.ProductionLine
	err := r.db.QueryRow(ctx, query, code).Scan(
		&line.ID,
		&line.Code,
		&line.Name,
		&line.Description,
		&line.Status,
		&line.CreatedAt,
		&line.UpdatedAt,
		&line.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get production line by code: %w", err)
	}

	return &line, nil
}

// List retrieves all active production lines
func (r *LineRepository) List(ctx context.Context) ([]domain.ProductionLine, error) {
	query := `
		SELECT id, code, name, description, status, created_at, updated_at
		FROM production_lines
		WHERE deleted_at IS NULL
		ORDER BY code ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list production lines: %w", err)
	}
	defer rows.Close()

	var lines []domain.ProductionLine
	for rows.Next() {
		var line domain.ProductionLine
		err := rows.Scan(
			&line.ID,
			&line.Code,
			&line.Name,
			&line.Description,
			&line.Status,
			&line.CreatedAt,
			&line.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan production line: %w", err)
		}
		lines = append(lines, line)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating production lines: %w", err)
	}

	// Return empty slice instead of nil for better JSON marshaling
	if lines == nil {
		lines = []domain.ProductionLine{}
	}

	// Enrich lines with labels
	lines, err = r.enrichWithLabels(ctx, lines)
	if err != nil {
		return nil, err
	}

	return lines, nil
}

// Update updates a production line's details
func (r *LineRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateLineRequest) (*domain.ProductionLine, error) {
	query := `
		UPDATE production_lines
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, code, name, description, status, created_at, updated_at
	`

	var line domain.ProductionLine
	err := r.db.QueryRow(ctx, query, id, req.Name, req.Description).Scan(
		&line.ID,
		&line.Code,
		&line.Name,
		&line.Description,
		&line.Status,
		&line.CreatedAt,
		&line.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update production line: %w", err)
	}

	return &line, nil
}

// UpdateStatus updates a production line's status
func (r *LineRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.Status) (*domain.ProductionLine, error) {
	query := `
		UPDATE production_lines
		SET status = $2,
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, code, name, description, status, created_at, updated_at
	`

	var line domain.ProductionLine
	err := r.db.QueryRow(ctx, query, id, status).Scan(
		&line.ID,
		&line.Code,
		&line.Name,
		&line.Description,
		&line.Status,
		&line.CreatedAt,
		&line.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update production line status: %w", err)
	}

	return &line, nil
}

// Delete soft deletes a production line
func (r *LineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE production_lines
		SET deleted_at = now(),
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete production line: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// enrichWithLabels loads labels for multiple lines in a single query (prevents N+1)
func (r *LineRepository) enrichWithLabels(ctx context.Context, lines []domain.ProductionLine) ([]domain.ProductionLine, error) {
	if len(lines) == 0 {
		return lines, nil
	}

	// Collect all line IDs
	lineIDs := make([]uuid.UUID, len(lines))
	for i, line := range lines {
		lineIDs[i] = line.ID
	}

	// Query all labels for these lines in one query
	query := `
		SELECT pll.line_id, l.id, l.name, l.color, l.description, l.created_at, l.updated_at
		FROM labels l
		INNER JOIN production_line_labels pll ON l.id = pll.label_id
		WHERE pll.line_id = ANY($1)
		ORDER BY pll.line_id, l.name
	`

	rows, err := r.db.Query(ctx, query, lineIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load labels: %w", err)
	}
	defer rows.Close()

	// Build a map of line_id -> []Label
	labelsMap := make(map[uuid.UUID][]domain.Label)
	for rows.Next() {
		var lineID uuid.UUID
		var label domain.Label
		err := rows.Scan(
			&lineID,
			&label.ID,
			&label.Name,
			&label.Color,
			&label.Description,
			&label.CreatedAt,
			&label.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labelsMap[lineID] = append(labelsMap[lineID], label)
	}

	// Attach labels to each line
	for i := range lines {
		if labels, ok := labelsMap[lines[i].ID]; ok {
			lines[i].Labels = labels
		} else {
			lines[i].Labels = []domain.Label{}
		}
	}

	return lines, nil
}

// getLabelsForLine retrieves labels for a single line (helper for GetByID)
func (r *LineRepository) getLabelsForLine(ctx context.Context, lineID uuid.UUID) ([]domain.Label, error) {
	query := `
		SELECT l.id, l.name, l.color, l.description, l.created_at, l.updated_at
		FROM labels l
		INNER JOIN production_line_labels pll ON l.id = pll.label_id
		WHERE pll.line_id = $1
		ORDER BY l.name ASC
	`

	rows, err := r.db.Query(ctx, query, lineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels for line: %w", err)
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var label domain.Label
		err := rows.Scan(
			&label.ID,
			&label.Name,
			&label.Color,
			&label.Description,
			&label.CreatedAt,
			&label.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan label: %w", err)
		}
		labels = append(labels, label)
	}

	if labels == nil {
		labels = []domain.Label{}
	}

	return labels, nil
}

// ListByLabels retrieves production lines that have ANY of the specified labels (OR logic)
func (r *LineRepository) ListByLabels(ctx context.Context, labelIDs []uuid.UUID) ([]domain.ProductionLine, error) {
	query := `
		SELECT DISTINCT pl.id, pl.code, pl.name, pl.description, pl.status,
		       pl.created_at, pl.updated_at
		FROM production_lines pl
		INNER JOIN production_line_labels pll ON pl.id = pll.line_id
		WHERE pl.deleted_at IS NULL
		  AND pll.label_id = ANY($1)
		ORDER BY pl.code ASC
	`

	rows, err := r.db.Query(ctx, query, labelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list lines by labels: %w", err)
	}
	defer rows.Close()

	var lines []domain.ProductionLine
	for rows.Next() {
		var line domain.ProductionLine
		err := rows.Scan(
			&line.ID,
			&line.Code,
			&line.Name,
			&line.Description,
			&line.Status,
			&line.CreatedAt,
			&line.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan line: %w", err)
		}
		lines = append(lines, line)
	}

	if lines == nil {
		lines = []domain.ProductionLine{}
	}

	// Enrich with labels
	return r.enrichWithLabels(ctx, lines)
}
