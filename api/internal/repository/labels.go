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

// LabelRepository handles database operations for labels
type LabelRepository struct {
	db *pgxpool.Pool
}

// NewLabelRepository creates a new LabelRepository
func NewLabelRepository(db *pgxpool.Pool) *LabelRepository {
	return &LabelRepository{db: db}
}

// Create creates a new label
func (r *LabelRepository) Create(ctx context.Context, req domain.CreateLabelRequest) (*domain.Label, error) {
	query := `
		INSERT INTO labels (name, color, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, color, description, created_at, updated_at
	`

	var label domain.Label
	err := r.db.QueryRow(ctx, query, req.Name, req.Color, req.Description).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.Description,
		&label.CreatedAt,
		&label.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, domain.ErrCodeExists
		}
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return &label, nil
}

// GetByID retrieves a label by ID
func (r *LabelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Label, error) {
	query := `
		SELECT id, name, color, description, created_at, updated_at
		FROM labels
		WHERE id = $1
	`

	var label domain.Label
	err := r.db.QueryRow(ctx, query, id).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.Description,
		&label.CreatedAt,
		&label.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return &label, nil
}

// List retrieves all labels
func (r *LabelRepository) List(ctx context.Context) ([]domain.Label, error) {
	query := `
		SELECT id, name, color, description, created_at, updated_at
		FROM labels
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
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

// Update updates a label
func (r *LabelRepository) Update(ctx context.Context, id uuid.UUID, req domain.UpdateLabelRequest) (*domain.Label, error) {
	query := `
		UPDATE labels
		SET name = COALESCE($2, name),
			color = COALESCE($3, color),
			description = COALESCE($4, description),
			updated_at = now()
		WHERE id = $1
		RETURNING id, name, color, description, created_at, updated_at
	`

	var label domain.Label
	err := r.db.QueryRow(ctx, query, id, req.Name, req.Color, req.Description).Scan(
		&label.ID,
		&label.Name,
		&label.Color,
		&label.Description,
		&label.CreatedAt,
		&label.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to update label: %w", err)
	}

	return &label, nil
}

// Delete deletes a label (cascade removes line associations)
func (r *LabelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM labels WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete label: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// AssignToLine assigns labels to a production line (replaces existing)
func (r *LabelRepository) AssignToLine(ctx context.Context, lineID uuid.UUID, labelIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Remove existing labels
	_, err = tx.Exec(ctx, `DELETE FROM production_line_labels WHERE line_id = $1`, lineID)
	if err != nil {
		return fmt.Errorf("failed to remove existing labels: %w", err)
	}

	// Insert new labels
	if len(labelIDs) > 0 {
		query := `
			INSERT INTO production_line_labels (line_id, label_id)
			VALUES ($1, $2)
		`

		for _, labelID := range labelIDs {
			_, err = tx.Exec(ctx, query, lineID, labelID)
			if err != nil {
				return fmt.Errorf("failed to assign label: %w", err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLabelsForLine retrieves all labels for a production line
func (r *LabelRepository) GetLabelsForLine(ctx context.Context, lineID uuid.UUID) ([]domain.Label, error) {
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

// GetLinesForLabel retrieves all production lines with a specific label
func (r *LabelRepository) GetLinesForLabel(ctx context.Context, labelID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT line_id
		FROM production_line_labels
		WHERE label_id = $1
	`

	rows, err := r.db.Query(ctx, query, labelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lines for label: %w", err)
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

	return lineIDs, nil
}
