package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"ping/production-line-api/internal/domain"
)

// SignalProfileRepository handles database operations for signal profiles
type SignalProfileRepository struct {
	db *pgxpool.Pool
}

// NewSignalProfileRepository creates a new SignalProfileRepository
func NewSignalProfileRepository(db *pgxpool.Pool) *SignalProfileRepository {
	return &SignalProfileRepository{db: db}
}

// ========== Profile CRUD Operations ==========

// Create creates a new signal profile
func (r *SignalProfileRepository) Create(ctx context.Context, profile *domain.SignalProfile) error {
	statesJSON, err := json.Marshal(profile.States)
	if err != nil {
		return fmt.Errorf("failed to marshal states: %w", err)
	}

	buttonBehaviorJSON, err := json.Marshal(profile.ButtonBehavior)
	if err != nil {
		return fmt.Errorf("failed to marshal button behavior: %w", err)
	}

	query := `
		INSERT INTO signal_profiles (name, description, version, states, button_behavior, default_state, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		profile.Name,
		profile.Description,
		profile.Version,
		statesJSON,
		buttonBehaviorJSON,
		profile.DefaultState,
		profile.UpdatedBy,
	).Scan(&profile.ID, &profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrProfileNameExists
		}
		return fmt.Errorf("failed to create signal profile: %w", err)
	}

	return nil
}

// GetByID retrieves a signal profile by its ID
func (r *SignalProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SignalProfile, error) {
	query := `
		SELECT id, name, description, version, states, button_behavior, default_state,
		       created_at, updated_at, updated_by, deleted_at
		FROM signal_profiles
		WHERE id = $1 AND deleted_at IS NULL
	`

	var profile domain.SignalProfile
	var statesJSON, buttonBehaviorJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.ID,
		&profile.Name,
		&profile.Description,
		&profile.Version,
		&statesJSON,
		&buttonBehaviorJSON,
		&profile.DefaultState,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get signal profile: %w", err)
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(statesJSON, &profile.States); err != nil {
		return nil, fmt.Errorf("failed to unmarshal states: %w", err)
	}

	if err := json.Unmarshal(buttonBehaviorJSON, &profile.ButtonBehavior); err != nil {
		return nil, fmt.Errorf("failed to unmarshal button behavior: %w", err)
	}

	return &profile, nil
}

// GetByName retrieves a signal profile by its name
func (r *SignalProfileRepository) GetByName(ctx context.Context, name string) (*domain.SignalProfile, error) {
	query := `
		SELECT id, name, description, version, states, button_behavior, default_state,
		       created_at, updated_at, updated_by, deleted_at
		FROM signal_profiles
		WHERE name = $1 AND deleted_at IS NULL
	`

	var profile domain.SignalProfile
	var statesJSON, buttonBehaviorJSON []byte

	err := r.db.QueryRow(ctx, query, name).Scan(
		&profile.ID,
		&profile.Name,
		&profile.Description,
		&profile.Version,
		&statesJSON,
		&buttonBehaviorJSON,
		&profile.DefaultState,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotFound
		}
		return nil, fmt.Errorf("failed to get signal profile by name: %w", err)
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(statesJSON, &profile.States); err != nil {
		return nil, fmt.Errorf("failed to unmarshal states: %w", err)
	}

	if err := json.Unmarshal(buttonBehaviorJSON, &profile.ButtonBehavior); err != nil {
		return nil, fmt.Errorf("failed to unmarshal button behavior: %w", err)
	}

	return &profile, nil
}

// List retrieves all signal profiles
func (r *SignalProfileRepository) List(ctx context.Context) ([]*domain.SignalProfile, error) {
	query := `
		SELECT id, name, description, version, states, button_behavior, default_state,
		       created_at, updated_at, updated_by, deleted_at
		FROM signal_profiles
		WHERE deleted_at IS NULL
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list signal profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*domain.SignalProfile
	for rows.Next() {
		var profile domain.SignalProfile
		var statesJSON, buttonBehaviorJSON []byte

		err := rows.Scan(
			&profile.ID,
			&profile.Name,
			&profile.Description,
			&profile.Version,
			&statesJSON,
			&buttonBehaviorJSON,
			&profile.DefaultState,
			&profile.CreatedAt,
			&profile.UpdatedAt,
			&profile.UpdatedBy,
			&profile.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signal profile: %w", err)
		}

		// Unmarshal JSONB fields
		if err := json.Unmarshal(statesJSON, &profile.States); err != nil {
			return nil, fmt.Errorf("failed to unmarshal states: %w", err)
		}

		if err := json.Unmarshal(buttonBehaviorJSON, &profile.ButtonBehavior); err != nil {
			return nil, fmt.Errorf("failed to unmarshal button behavior: %w", err)
		}

		profiles = append(profiles, &profile)
	}

	return profiles, nil
}

// Update updates an existing signal profile
func (r *SignalProfileRepository) Update(ctx context.Context, id uuid.UUID, profile *domain.SignalProfile) error {
	statesJSON, err := json.Marshal(profile.States)
	if err != nil {
		return fmt.Errorf("failed to marshal states: %w", err)
	}

	buttonBehaviorJSON, err := json.Marshal(profile.ButtonBehavior)
	if err != nil {
		return fmt.Errorf("failed to marshal button behavior: %w", err)
	}

	query := `
		UPDATE signal_profiles
		SET name = $1, description = $2, version = $3, states = $4,
		    button_behavior = $5, default_state = $6, updated_by = $7
		WHERE id = $8 AND deleted_at IS NULL
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		profile.Name,
		profile.Description,
		profile.Version,
		statesJSON,
		buttonBehaviorJSON,
		profile.DefaultState,
		profile.UpdatedBy,
		id,
	).Scan(&profile.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrProfileNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrProfileNameExists
		}
		return fmt.Errorf("failed to update signal profile: %w", err)
	}

	return nil
}

// Delete soft-deletes a signal profile
func (r *SignalProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE signal_profiles
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete signal profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}

// IncrementVersion increments the version of a profile and returns the new version
func (r *SignalProfileRepository) IncrementVersion(ctx context.Context, id uuid.UUID) (int, error) {
	query := `
		UPDATE signal_profiles
		SET version = version + 1
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING version
	`

	var newVersion int
	err := r.db.QueryRow(ctx, query, id).Scan(&newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrProfileNotFound
		}
		return 0, fmt.Errorf("failed to increment profile version: %w", err)
	}

	return newVersion, nil
}

// ========== Version Management ==========

// CreateVersion creates a version history entry
func (r *SignalProfileRepository) CreateVersion(ctx context.Context, version *domain.ProfileVersion) error {
	configJSON, err := json.Marshal(version.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO signal_profile_versions (profile_id, version, config, changed_by, change_description, changes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err = r.db.QueryRow(ctx, query,
		version.ProfileID,
		version.Version,
		configJSON,
		version.ChangedBy,
		version.ChangeDescription,
		version.Changes,
	).Scan(&version.ID, &version.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create profile version: %w", err)
	}

	return nil
}

// GetVersions retrieves all version history for a profile
func (r *SignalProfileRepository) GetVersions(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileVersion, error) {
	query := `
		SELECT id, profile_id, version, config, changed_by, change_description, changes, created_at
		FROM signal_profile_versions
		WHERE profile_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile versions: %w", err)
	}
	defer rows.Close()

	var versions []*domain.ProfileVersion
	for rows.Next() {
		var version domain.ProfileVersion
		var configJSON []byte

		err := rows.Scan(
			&version.ID,
			&version.ProfileID,
			&version.Version,
			&configJSON,
			&version.ChangedBy,
			&version.ChangeDescription,
			&version.Changes,
			&version.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan profile version: %w", err)
		}

		// Unmarshal config
		if err := json.Unmarshal(configJSON, &version.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		versions = append(versions, &version)
	}

	return versions, nil
}

// GetVersionByNumber retrieves a specific version of a profile
func (r *SignalProfileRepository) GetVersionByNumber(ctx context.Context, profileID uuid.UUID, version int) (*domain.ProfileVersion, error) {
	query := `
		SELECT id, profile_id, version, config, changed_by, change_description, changes, created_at
		FROM signal_profile_versions
		WHERE profile_id = $1 AND version = $2
	`

	var profileVersion domain.ProfileVersion
	var configJSON []byte

	err := r.db.QueryRow(ctx, query, profileID, version).Scan(
		&profileVersion.ID,
		&profileVersion.ProfileID,
		&profileVersion.Version,
		&configJSON,
		&profileVersion.ChangedBy,
		&profileVersion.ChangeDescription,
		&profileVersion.Changes,
		&profileVersion.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileVersionNotFound
		}
		return nil, fmt.Errorf("failed to get profile version: %w", err)
	}

	// Unmarshal config
	if err := json.Unmarshal(configJSON, &profileVersion.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &profileVersion, nil
}

// ========== Line Assignment ==========

// AssignToLine assigns a profile to a production line
func (r *SignalProfileRepository) AssignToLine(ctx context.Context, lineID uuid.UUID, profileID uuid.UUID) error {
	query := `
		UPDATE production_lines
		SET signal_profile_id = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, profileID, lineID)
	if err != nil {
		return fmt.Errorf("failed to assign profile to line: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetLineProfile retrieves the profile assigned to a production line
func (r *SignalProfileRepository) GetLineProfile(ctx context.Context, lineID uuid.UUID) (*domain.SignalProfile, error) {
	query := `
		SELECT sp.id, sp.name, sp.description, sp.version, sp.states, sp.button_behavior,
		       sp.default_state, sp.created_at, sp.updated_at, sp.updated_by, sp.deleted_at
		FROM signal_profiles sp
		INNER JOIN production_lines pl ON sp.id = pl.signal_profile_id
		WHERE pl.id = $1 AND pl.deleted_at IS NULL AND sp.deleted_at IS NULL
	`

	var profile domain.SignalProfile
	var statesJSON, buttonBehaviorJSON []byte

	err := r.db.QueryRow(ctx, query, lineID).Scan(
		&profile.ID,
		&profile.Name,
		&profile.Description,
		&profile.Version,
		&statesJSON,
		&buttonBehaviorJSON,
		&profile.DefaultState,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileNotAssignedToLine
		}
		return nil, fmt.Errorf("failed to get line profile: %w", err)
	}

	// Unmarshal JSONB fields
	if err := json.Unmarshal(statesJSON, &profile.States); err != nil {
		return nil, fmt.Errorf("failed to unmarshal states: %w", err)
	}

	if err := json.Unmarshal(buttonBehaviorJSON, &profile.ButtonBehavior); err != nil {
		return nil, fmt.Errorf("failed to unmarshal button behavior: %w", err)
	}

	return &profile, nil
}

// GetLinesUsingProfile retrieves all line IDs using a specific profile
func (r *SignalProfileRepository) GetLinesUsingProfile(ctx context.Context, profileID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT id
		FROM production_lines
		WHERE signal_profile_id = $1 AND deleted_at IS NULL
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lines using profile: %w", err)
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

// IsProfileInUse checks if a profile is assigned to any lines
func (r *SignalProfileRepository) IsProfileInUse(ctx context.Context, profileID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM production_lines
			WHERE signal_profile_id = $1 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, profileID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if profile is in use: %w", err)
	}

	return exists, nil
}

// ========== Device State Tracking ==========

// UpsertDeviceState creates or updates device signal state
func (r *SignalProfileRepository) UpsertDeviceState(ctx context.Context, state *domain.DeviceSignalState) error {
	query := `
		INSERT INTO device_signal_state (
			device_mac, profile_id, profile_version, current_state, is_overridden,
			profile_hash, last_state_change, last_sync, last_version_check, version_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (device_mac) DO UPDATE SET
			profile_id = EXCLUDED.profile_id,
			profile_version = EXCLUDED.profile_version,
			current_state = EXCLUDED.current_state,
			is_overridden = EXCLUDED.is_overridden,
			profile_hash = EXCLUDED.profile_hash,
			last_state_change = EXCLUDED.last_state_change,
			last_sync = EXCLUDED.last_sync,
			last_version_check = EXCLUDED.last_version_check,
			version_status = EXCLUDED.version_status
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		state.DeviceMAC,
		state.ProfileID,
		state.ProfileVersion,
		state.CurrentState,
		state.IsOverridden,
		state.ProfileHash,
		state.LastStateChange,
		state.LastSync,
		state.LastVersionCheck,
		state.VersionStatus,
	).Scan(&state.CreatedAt, &state.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert device signal state: %w", err)
	}

	return nil
}

// GetDeviceState retrieves device signal state
func (r *SignalProfileRepository) GetDeviceState(ctx context.Context, deviceMAC string) (*domain.DeviceSignalState, error) {
	query := `
		SELECT device_mac, profile_id, profile_version, current_state, is_overridden,
		       profile_hash, last_state_change, last_sync, last_version_check,
		       version_status, created_at, updated_at
		FROM device_signal_state
		WHERE device_mac = $1
	`

	var state domain.DeviceSignalState
	err := r.db.QueryRow(ctx, query, deviceMAC).Scan(
		&state.DeviceMAC,
		&state.ProfileID,
		&state.ProfileVersion,
		&state.CurrentState,
		&state.IsOverridden,
		&state.ProfileHash,
		&state.LastStateChange,
		&state.LastSync,
		&state.LastVersionCheck,
		&state.VersionStatus,
		&state.CreatedAt,
		&state.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDeviceStateNotFound
		}
		return nil, fmt.Errorf("failed to get device signal state: %w", err)
	}

	return &state, nil
}

// UpdateDeviceVersionStatus updates only the version status of a device
func (r *SignalProfileRepository) UpdateDeviceVersionStatus(ctx context.Context, deviceMAC string, status string) error {
	query := `
		UPDATE device_signal_state
		SET version_status = $1, last_version_check = NOW()
		WHERE device_mac = $2
	`

	result, err := r.db.Exec(ctx, query, status, deviceMAC)
	if err != nil {
		return fmt.Errorf("failed to update device version status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDeviceStateNotFound
	}

	return nil
}

// GetDevicesWithProfile retrieves all devices using a specific profile
func (r *SignalProfileRepository) GetDevicesWithProfile(ctx context.Context, profileID uuid.UUID) ([]*domain.DeviceSignalState, error) {
	query := `
		SELECT device_mac, profile_id, profile_version, current_state, is_overridden,
		       profile_hash, last_state_change, last_sync, last_version_check,
		       version_status, created_at, updated_at
		FROM device_signal_state
		WHERE profile_id = $1
	`

	rows, err := r.db.Query(ctx, query, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices with profile: %w", err)
	}
	defer rows.Close()

	var states []*domain.DeviceSignalState
	for rows.Next() {
		var state domain.DeviceSignalState
		err := rows.Scan(
			&state.DeviceMAC,
			&state.ProfileID,
			&state.ProfileVersion,
			&state.CurrentState,
			&state.IsOverridden,
			&state.ProfileHash,
			&state.LastStateChange,
			&state.LastSync,
			&state.LastVersionCheck,
			&state.VersionStatus,
			&state.CreatedAt,
			&state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device signal state: %w", err)
		}
		states = append(states, &state)
	}

	return states, nil
}

// GetOverriddenDevices retrieves all devices with active overrides
func (r *SignalProfileRepository) GetOverriddenDevices(ctx context.Context) ([]*domain.DeviceSignalState, error) {
	query := `
		SELECT device_mac, profile_id, profile_version, current_state, is_overridden,
		       profile_hash, last_state_change, last_sync, last_version_check,
		       version_status, created_at, updated_at
		FROM device_signal_state
		WHERE is_overridden = TRUE
		ORDER BY last_state_change DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get overridden devices: %w", err)
	}
	defer rows.Close()

	var states []*domain.DeviceSignalState
	for rows.Next() {
		var state domain.DeviceSignalState
		err := rows.Scan(
			&state.DeviceMAC,
			&state.ProfileID,
			&state.ProfileVersion,
			&state.CurrentState,
			&state.IsOverridden,
			&state.ProfileHash,
			&state.LastStateChange,
			&state.LastSync,
			&state.LastVersionCheck,
			&state.VersionStatus,
			&state.CreatedAt,
			&state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device signal state: %w", err)
		}
		states = append(states, &state)
	}

	return states, nil
}

// ClearDeviceOverride clears the override flag for a device
func (r *SignalProfileRepository) ClearDeviceOverride(ctx context.Context, deviceMAC string) error {
	query := `
		UPDATE device_signal_state
		SET is_overridden = FALSE
		WHERE device_mac = $1
	`

	result, err := r.db.Exec(ctx, query, deviceMAC)
	if err != nil {
		return fmt.Errorf("failed to clear device override: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDeviceStateNotFound
	}

	return nil
}

// GetDevicesNeedingUpdate retrieves devices with outdated profile versions
func (r *SignalProfileRepository) GetDevicesNeedingUpdate(ctx context.Context, profileID uuid.UUID, currentVersion int) ([]*domain.DeviceSignalState, error) {
	query := `
		SELECT device_mac, profile_id, profile_version, current_state, is_overridden,
		       profile_hash, last_state_change, last_sync, last_version_check,
		       version_status, created_at, updated_at
		FROM device_signal_state
		WHERE profile_id = $1 AND profile_version < $2
		ORDER BY last_version_check DESC
	`

	rows, err := r.db.Query(ctx, query, profileID, currentVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices needing update: %w", err)
	}
	defer rows.Close()

	var states []*domain.DeviceSignalState
	for rows.Next() {
		var state domain.DeviceSignalState
		err := rows.Scan(
			&state.DeviceMAC,
			&state.ProfileID,
			&state.ProfileVersion,
			&state.CurrentState,
			&state.IsOverridden,
			&state.ProfileHash,
			&state.LastStateChange,
			&state.LastSync,
			&state.LastVersionCheck,
			&state.VersionStatus,
			&state.CreatedAt,
			&state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device signal state: %w", err)
		}
		states = append(states, &state)
	}

	return states, nil
}
