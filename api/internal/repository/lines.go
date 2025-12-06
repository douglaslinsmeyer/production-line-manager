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
