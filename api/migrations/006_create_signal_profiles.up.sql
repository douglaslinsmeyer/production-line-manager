-- Migration 006: Create Signal Profiles
-- This migration adds support for configurable signal profiles that define
-- how assembly line devices respond to different operational states.

-- ==============================================================================
-- Table: signal_profiles
-- Core signal profile definitions with states and button behavior
-- ==============================================================================
CREATE TABLE signal_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    version INTEGER NOT NULL DEFAULT 1,

    -- JSONB structure: Array of states with output configurations
    -- Example: [{"name": "On", "outputs": {"redLight": "off", "yellowLight": "off", "greenLight": "on", "buzzer": "off"}}]
    states JSONB NOT NULL,

    -- JSONB structure: Button press cycle arrays
    -- Example: {"shortPressCycle": ["On", "Off"], "longPressCycle": ["Maintenance"]}
    button_behavior JSONB NOT NULL,

    default_state VARCHAR(50) NOT NULL,

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by TEXT,
    deleted_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT chk_version_positive CHECK (version > 0)
);

-- Indexes for signal_profiles
CREATE UNIQUE INDEX idx_signal_profiles_name_active
    ON signal_profiles(name) WHERE deleted_at IS NULL;

CREATE INDEX idx_signal_profiles_deleted_at
    ON signal_profiles(deleted_at);

-- Add comment
COMMENT ON TABLE signal_profiles IS 'Signal profile templates defining states, outputs, and button behavior for assembly line devices';
COMMENT ON COLUMN signal_profiles.states IS 'JSONB array of state definitions with output configurations';
COMMENT ON COLUMN signal_profiles.button_behavior IS 'JSONB object with shortPressCycle and longPressCycle arrays';
COMMENT ON COLUMN signal_profiles.version IS 'Auto-incremented version number for profile configuration changes';

-- ==============================================================================
-- Table: signal_profile_versions
-- Version history for profile configurations (audit trail and rollback support)
-- ==============================================================================
CREATE TABLE signal_profile_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES signal_profiles(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,

    -- Full profile snapshot at this version
    config JSONB NOT NULL,

    -- Change tracking
    changed_by TEXT NOT NULL,
    change_description TEXT,
    changes TEXT[],  -- Array of change descriptions

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique version per profile
    CONSTRAINT uq_profile_version UNIQUE (profile_id, version)
);

-- Indexes for signal_profile_versions
CREATE INDEX idx_profile_versions_profile
    ON signal_profile_versions(profile_id);

CREATE INDEX idx_profile_versions_created
    ON signal_profile_versions(created_at DESC);

-- Add comment
COMMENT ON TABLE signal_profile_versions IS 'Version history for signal profiles, enabling audit trail and rollback functionality';

-- ==============================================================================
-- Table: device_signal_state
-- Tracks each device's current state, profile version, and override status
-- ==============================================================================
CREATE TABLE device_signal_state (
    device_mac TEXT PRIMARY KEY REFERENCES discovered_devices(mac_address) ON DELETE CASCADE,

    -- Profile tracking
    profile_id UUID REFERENCES signal_profiles(id) ON DELETE SET NULL,
    profile_version INTEGER NOT NULL DEFAULT 1,
    profile_hash TEXT,  -- MD5 hash for validation (optional)

    -- Current state
    current_state VARCHAR(50) NOT NULL,
    is_overridden BOOLEAN NOT NULL DEFAULT FALSE,

    -- Sync tracking
    last_state_change TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_sync TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_version_check TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Version sync status
    version_status TEXT NOT NULL DEFAULT 'unknown'
        CHECK (version_status IN ('up-to-date', 'update-pending', 'failed', 'offline', 'unknown')),

    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for device_signal_state
CREATE INDEX idx_device_signal_state_profile
    ON device_signal_state(profile_id);

CREATE INDEX idx_device_signal_state_override
    ON device_signal_state(is_overridden)
    WHERE is_overridden = TRUE;

CREATE INDEX idx_device_signal_state_version_status
    ON device_signal_state(version_status);

CREATE INDEX idx_device_signal_state_last_check
    ON device_signal_state(last_version_check);

-- Add comment
COMMENT ON TABLE device_signal_state IS 'Tracks device current state, profile version, and synchronization status';
COMMENT ON COLUMN device_signal_state.is_overridden IS 'True if device state was manually changed (not from profile default)';
COMMENT ON COLUMN device_signal_state.version_status IS 'Sync status: up-to-date, update-pending, failed, offline, unknown';

-- ==============================================================================
-- Extend: production_lines
-- Add signal profile assignment to production lines
-- ==============================================================================
ALTER TABLE production_lines
    ADD COLUMN signal_profile_id UUID REFERENCES signal_profiles(id) ON DELETE SET NULL;

-- Index for profile lookup
CREATE INDEX idx_production_lines_profile
    ON production_lines(signal_profile_id)
    WHERE deleted_at IS NULL;

-- Add comment
COMMENT ON COLUMN production_lines.signal_profile_id IS
    'Signal profile assigned to this line. Devices inherit profile when assigned to line.';

-- ==============================================================================
-- Triggers: Auto-update timestamps
-- ==============================================================================

-- Trigger function for signal_profiles updated_at
CREATE OR REPLACE FUNCTION update_signal_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_signal_profiles_updated_at
    BEFORE UPDATE ON signal_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_signal_profiles_updated_at();

-- Trigger function for device_signal_state updated_at
CREATE TRIGGER trigger_update_device_signal_state_updated_at
    BEFORE UPDATE ON device_signal_state
    FOR EACH ROW
    EXECUTE FUNCTION update_signal_profiles_updated_at();

-- ==============================================================================
-- Seed Data: Default Signal Profile
-- Create a default profile matching current device behavior (ON/OFF/MAINTENANCE/ERROR)
-- ==============================================================================
INSERT INTO signal_profiles (
    id,
    name,
    description,
    version,
    states,
    button_behavior,
    default_state,
    updated_by
) VALUES (
    'f47ac10b-58cc-4372-a567-0e02b2c3d479',  -- Fixed UUID for default profile
    'Assembly Line Default',
    'Default signal profile matching standard ON/OFF/MAINTENANCE/ERROR behavior',
    1,
    '[
        {
            "name": "On",
            "outputs": {
                "redLight": "off",
                "yellowLight": "off",
                "greenLight": "on",
                "buzzer": "off"
            }
        },
        {
            "name": "Off",
            "outputs": {
                "redLight": "on",
                "yellowLight": "off",
                "greenLight": "off",
                "buzzer": "off"
            }
        },
        {
            "name": "Maintenance",
            "outputs": {
                "redLight": "off",
                "yellowLight": "on",
                "greenLight": "off",
                "buzzer": "off"
            }
        },
        {
            "name": "Error",
            "outputs": {
                "redLight": "shortBlink",
                "yellowLight": "off",
                "greenLight": "off",
                "buzzer": "on"
            }
        }
    ]'::jsonb,
    '{
        "shortPressCycle": ["On", "Off"],
        "longPressCycle": ["Maintenance"]
    }'::jsonb,
    'Off',
    'system'
);

-- Create initial version history entry for default profile
INSERT INTO signal_profile_versions (
    profile_id,
    version,
    config,
    changed_by,
    change_description,
    changes
) VALUES (
    'f47ac10b-58cc-4372-a567-0e02b2c3d479',
    1,
    '{
        "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
        "name": "Assembly Line Default",
        "description": "Default signal profile matching standard ON/OFF/MAINTENANCE/ERROR behavior",
        "version": 1,
        "states": [
            {
                "name": "On",
                "outputs": {
                    "redLight": "off",
                    "yellowLight": "off",
                    "greenLight": "on",
                    "buzzer": "off"
                }
            },
            {
                "name": "Off",
                "outputs": {
                    "redLight": "on",
                    "yellowLight": "off",
                    "greenLight": "off",
                    "buzzer": "off"
                }
            },
            {
                "name": "Maintenance",
                "outputs": {
                    "redLight": "off",
                    "yellowLight": "on",
                    "greenLight": "off",
                    "buzzer": "off"
                }
            },
            {
                "name": "Error",
                "outputs": {
                    "redLight": "shortBlink",
                    "yellowLight": "off",
                    "greenLight": "off",
                    "buzzer": "on"
                }
            }
        ],
        "buttonBehavior": {
            "shortPressCycle": ["On", "Off"],
            "longPressCycle": ["Maintenance"]
        },
        "defaultState": "Off"
    }'::jsonb,
    'system',
    'Initial profile creation',
    ARRAY['Created default profile with ON, OFF, MAINTENANCE, ERROR states']
);
