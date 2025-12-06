-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create status log table
CREATE TABLE production_line_status_log (
    time TIMESTAMPTZ NOT NULL DEFAULT now(),
    line_id UUID NOT NULL REFERENCES production_lines(id),
    line_code VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL CHECK (new_status IN ('on', 'off', 'maintenance', 'error')),
    source VARCHAR(20) NOT NULL,
    source_detail JSONB,
    PRIMARY KEY (time, line_id)
);

-- Convert to hypertable (7-day chunks)
SELECT create_hypertable(
    'production_line_status_log',
    'time',
    chunk_time_interval => INTERVAL '7 days'
);

-- Create index for querying by line_id
CREATE INDEX idx_status_log_line
    ON production_line_status_log (line_id, time DESC);

-- Enable compression
ALTER TABLE production_line_status_log SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'line_id'
);

-- Add compression policy (compress chunks older than 30 days)
SELECT add_compression_policy(
    'production_line_status_log',
    INTERVAL '30 days'
);

-- Optional: Add retention policy (uncomment to enable)
-- Keep status logs for 2 years, then automatically drop old chunks
-- SELECT add_retention_policy(
--     'production_line_status_log',
--     INTERVAL '2 years'
-- );

-- Add comments
COMMENT ON TABLE production_line_status_log IS 'Time-series audit log of production line status changes';
COMMENT ON COLUMN production_line_status_log.source IS 'Source of status change: api, mqtt, system';
COMMENT ON COLUMN production_line_status_log.source_detail IS 'Additional details about the source (e.g., user ID, controller ID)';
