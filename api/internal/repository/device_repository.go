package repository

import (
	"context"
	"time"

	"ping/production-line-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceRepository struct {
	db *pgxpool.Pool
}

// NewDeviceRepository creates a new device repository
func NewDeviceRepository(db *pgxpool.Pool) domain.DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) UpsertDevice(device *domain.DiscoveredDevice) error {
	ctx := context.Background()

	query := `
		INSERT INTO discovered_devices (
			mac_address, device_type, firmware_version, ip_address,
			capabilities, first_seen, last_seen, status, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (mac_address) DO UPDATE SET
			device_type = EXCLUDED.device_type,
			firmware_version = EXCLUDED.firmware_version,
			ip_address = EXCLUDED.ip_address,
			capabilities = EXCLUDED.capabilities,
			last_seen = EXCLUDED.last_seen,
			status = EXCLUDED.status,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		device.MACAddress,
		device.DeviceType,
		device.FirmwareVersion,
		device.IPAddress,
		device.Capabilities,
		device.FirstSeen,
		device.LastSeen,
		device.Status,
		device.Metadata,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)
}

func (r *deviceRepository) GetDeviceByMAC(macAddress string) (*domain.DiscoveredDevice, error) {
	ctx := context.Background()
	var device domain.DiscoveredDevice

	query := `
		SELECT id, mac_address, device_type, firmware_version, ip_address,
		       capabilities, first_seen, last_seen, status, metadata,
		       created_at, updated_at
		FROM discovered_devices
		WHERE mac_address = $1
	`

	err := r.db.QueryRow(ctx, query, macAddress).Scan(
		&device.ID,
		&device.MACAddress,
		&device.DeviceType,
		&device.FirmwareVersion,
		&device.IPAddress,
		&device.Capabilities,
		&device.FirstSeen,
		&device.LastSeen,
		&device.Status,
		&device.Metadata,
		&device.CreatedAt,
		&device.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &device, err
}

func (r *deviceRepository) ListDevices() ([]*domain.DeviceWithAssignment, error) {
	ctx := context.Background()

	query := `
		SELECT
			d.id, d.mac_address, d.device_type, d.firmware_version, d.ip_address,
			d.capabilities, d.first_seen, d.last_seen, d.status, d.metadata,
			d.created_at, d.updated_at,
			dla.line_id as assigned_line_id,
			pl.code as assigned_line_code,
			pl.name as assigned_line_name,
			dla.assigned_at
		FROM discovered_devices d
		LEFT JOIN device_line_assignments dla
			ON d.mac_address = dla.device_mac AND dla.active = true
		LEFT JOIN production_lines pl
			ON dla.line_id = pl.id
		ORDER BY d.last_seen DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.DeviceWithAssignment
	for rows.Next() {
		var device domain.DeviceWithAssignment
		err := rows.Scan(
			&device.ID,
			&device.MACAddress,
			&device.DeviceType,
			&device.FirmwareVersion,
			&device.IPAddress,
			&device.Capabilities,
			&device.FirstSeen,
			&device.LastSeen,
			&device.Status,
			&device.Metadata,
			&device.CreatedAt,
			&device.UpdatedAt,
			&device.AssignedLineID,
			&device.AssignedLineCode,
			&device.AssignedLineName,
			&device.AssignedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}

	return devices, rows.Err()
}

func (r *deviceRepository) UpdateDeviceStatus(macAddress string, status domain.DeviceStatus) error {
	ctx := context.Background()

	query := `
		UPDATE discovered_devices
		SET status = $1, updated_at = NOW()
		WHERE mac_address = $2
	`

	_, err := r.db.Exec(ctx, query, status, macAddress)
	return err
}

func (r *deviceRepository) MarkStaleDevicesOffline(threshold time.Duration) error {
	ctx := context.Background()

	query := `
		UPDATE discovered_devices
		SET status = $1, updated_at = NOW()
		WHERE last_seen < $2 AND status = $3
	`

	cutoff := time.Now().Add(-threshold)
	_, err := r.db.Exec(ctx, query, domain.DeviceStatusOffline, cutoff, domain.DeviceStatusOnline)
	return err
}

func (r *deviceRepository) AssignDeviceToLine(deviceMAC string, lineID uuid.UUID, assignedBy *string) error {
	ctx := context.Background()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Deactivate any existing assignments for this device
	_, err = tx.Exec(ctx, `
		UPDATE device_line_assignments
		SET active = false, updated_at = NOW()
		WHERE device_mac = $1 AND active = true
	`, deviceMAC)
	if err != nil {
		return err
	}

	// Create new assignment
	_, err = tx.Exec(ctx, `
		INSERT INTO device_line_assignments (device_mac, line_id, assigned_by, active)
		VALUES ($1, $2, $3, true)
	`, deviceMAC, lineID, assignedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *deviceRepository) UnassignDevice(deviceMAC string) error {
	ctx := context.Background()

	query := `
		UPDATE device_line_assignments
		SET active = false, updated_at = NOW()
		WHERE device_mac = $1 AND active = true
	`

	_, err := r.db.Exec(ctx, query, deviceMAC)
	return err
}

func (r *deviceRepository) GetDeviceAssignment(deviceMAC string) (*domain.DeviceLineAssignment, error) {
	ctx := context.Background()
	var assignment domain.DeviceLineAssignment

	query := `
		SELECT id, device_mac, line_id, assigned_at, assigned_by, active, created_at, updated_at
		FROM device_line_assignments
		WHERE device_mac = $1 AND active = true
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, deviceMAC).Scan(
		&assignment.ID,
		&assignment.DeviceMAC,
		&assignment.LineID,
		&assignment.AssignedAt,
		&assignment.AssignedBy,
		&assignment.Active,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &assignment, err
}

func (r *deviceRepository) GetLineAssignment(lineID uuid.UUID) (*domain.DeviceLineAssignment, error) {
	ctx := context.Background()
	var assignment domain.DeviceLineAssignment

	query := `
		SELECT id, device_mac, line_id, assigned_at, assigned_by, active, created_at, updated_at
		FROM device_line_assignments
		WHERE line_id = $1 AND active = true
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, lineID).Scan(
		&assignment.ID,
		&assignment.DeviceMAC,
		&assignment.LineID,
		&assignment.AssignedAt,
		&assignment.AssignedBy,
		&assignment.Active,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &assignment, err
}

func (r *deviceRepository) ListAssignments() ([]*domain.DeviceLineAssignment, error) {
	ctx := context.Background()

	query := `
		SELECT id, device_mac, line_id, assigned_at, assigned_by, active, created_at, updated_at
		FROM device_line_assignments
		WHERE active = true
		ORDER BY assigned_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*domain.DeviceLineAssignment
	for rows.Next() {
		var assignment domain.DeviceLineAssignment
		err := rows.Scan(
			&assignment.ID,
			&assignment.DeviceMAC,
			&assignment.LineID,
			&assignment.AssignedAt,
			&assignment.AssignedBy,
			&assignment.Active,
			&assignment.CreatedAt,
			&assignment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, &assignment)
	}

	return assignments, rows.Err()
}
