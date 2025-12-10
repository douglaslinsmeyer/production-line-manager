-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_device_assignments_updated_at ON device_line_assignments;
DROP TRIGGER IF EXISTS trigger_update_discovered_devices_updated_at ON discovered_devices;

-- Drop functions
DROP FUNCTION IF EXISTS update_device_assignments_updated_at();
DROP FUNCTION IF EXISTS update_discovered_devices_updated_at();

-- Drop tables (cascade will handle foreign keys)
DROP TABLE IF EXISTS device_line_assignments CASCADE;
DROP TABLE IF EXISTS discovered_devices CASCADE;
