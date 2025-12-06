-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create production_lines table
CREATE TABLE production_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'off'
        CHECK (status IN ('on', 'off', 'maintenance', 'error')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX idx_production_lines_code ON production_lines(code);
CREATE INDEX idx_production_lines_status ON production_lines(status)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_production_lines_deleted_at ON production_lines(deleted_at);

-- Add comment
COMMENT ON TABLE production_lines IS 'Production lines in the manufacturing facility';
COMMENT ON COLUMN production_lines.deleted_at IS 'Soft delete timestamp - NULL means active';
