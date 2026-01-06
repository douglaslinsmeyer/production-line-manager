-- Migration 006 Rollback: Drop Signal Profiles
-- This migration rolls back all signal profile related changes

-- ==============================================================================
-- Drop triggers
-- ==============================================================================
DROP TRIGGER IF EXISTS trigger_update_device_signal_state_updated_at ON device_signal_state;
DROP TRIGGER IF EXISTS trigger_update_signal_profiles_updated_at ON signal_profiles;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_signal_profiles_updated_at();

-- ==============================================================================
-- Remove production_lines extension
-- ==============================================================================
ALTER TABLE production_lines DROP COLUMN IF EXISTS signal_profile_id;

-- ==============================================================================
-- Drop tables in dependency order
-- ==============================================================================

-- Drop device_signal_state (references signal_profiles)
DROP TABLE IF EXISTS device_signal_state;

-- Drop signal_profile_versions (references signal_profiles)
DROP TABLE IF EXISTS signal_profile_versions;

-- Drop signal_profiles (base table)
DROP TABLE IF EXISTS signal_profiles;
