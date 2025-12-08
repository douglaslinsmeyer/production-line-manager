-- Create labels table
CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    color VARCHAR(7),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index on label name for search performance
CREATE INDEX idx_labels_name ON labels(name);

-- Add comments
COMMENT ON TABLE labels IS 'Labels for categorizing production lines';
COMMENT ON COLUMN labels.color IS 'Hex color code for UI display (e.g., #FF5733)';

-- Create junction table for many-to-many relationship
CREATE TABLE production_line_labels (
    line_id UUID NOT NULL REFERENCES production_lines(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (line_id, label_id)
);

-- Create indexes for efficient queries
CREATE INDEX idx_production_line_labels_line ON production_line_labels(line_id);
CREATE INDEX idx_production_line_labels_label ON production_line_labels(label_id);
CREATE INDEX idx_production_line_labels_assigned ON production_line_labels(assigned_at);

-- Add comment
COMMENT ON TABLE production_line_labels IS 'Many-to-many relationship between production lines and labels';
