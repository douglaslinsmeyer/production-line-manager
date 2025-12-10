-- Create discovered_devices table
CREATE TABLE IF NOT EXISTS discovered_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mac_address TEXT UNIQUE NOT NULL,
    device_type TEXT NOT NULL,
    firmware_version TEXT,
    ip_address TEXT,
    capabilities JSONB,
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'online',
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for discovered_devices
CREATE INDEX idx_discovered_devices_mac ON discovered_devices(mac_address);
CREATE INDEX idx_discovered_devices_last_seen ON discovered_devices(last_seen);
CREATE INDEX idx_discovered_devices_status ON discovered_devices(status);

-- Create device_line_assignments table
CREATE TABLE IF NOT EXISTS device_line_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_mac TEXT NOT NULL,
    line_id UUID NOT NULL REFERENCES production_lines(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_device FOREIGN KEY (device_mac) REFERENCES discovered_devices(mac_address) ON DELETE CASCADE
);

-- Create indexes for device_line_assignments
CREATE INDEX IF NOT EXISTS idx_device_assignments_mac ON device_line_assignments(device_mac);
CREATE INDEX IF NOT EXISTS idx_device_assignments_line ON device_line_assignments(line_id);
CREATE INDEX IF NOT EXISTS idx_device_assignments_active ON device_line_assignments(active);

-- Create partial unique index for active assignments (only one active assignment per device)
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_device_assignment
    ON device_line_assignments(device_mac)
    WHERE active = TRUE;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_discovered_devices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_discovered_devices_updated_at
    BEFORE UPDATE ON discovered_devices
    FOR EACH ROW
    EXECUTE FUNCTION update_discovered_devices_updated_at();

CREATE OR REPLACE FUNCTION update_device_assignments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_device_assignments_updated_at
    BEFORE UPDATE ON device_line_assignments
    FOR EACH ROW
    EXECUTE FUNCTION update_device_assignments_updated_at();

-- Comments
COMMENT ON TABLE discovered_devices IS 'Registry of all ESP32 devices that have announced themselves via MQTT';
COMMENT ON TABLE device_line_assignments IS 'Mapping between devices (by MAC address) and production lines';
COMMENT ON COLUMN discovered_devices.mac_address IS 'Unique MAC address of the device (format: XX:XX:XX:XX:XX:XX)';
COMMENT ON COLUMN discovered_devices.status IS 'Device status: online, offline, unknown';
COMMENT ON COLUMN discovered_devices.capabilities IS 'JSON object describing device capabilities (inputs, outputs, protocols)';
COMMENT ON COLUMN device_line_assignments.active IS 'Only one active assignment per device at a time';
