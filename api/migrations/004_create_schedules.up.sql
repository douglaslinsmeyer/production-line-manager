-- ============================================================================
-- SCHEDULES TABLE (Base weekly schedule templates)
-- ============================================================================
CREATE TABLE schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_schedules_name_active ON schedules(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_schedules_deleted_at ON schedules(deleted_at);

COMMENT ON TABLE schedules IS 'Base production schedules that can be assigned to multiple lines';
COMMENT ON COLUMN schedules.timezone IS 'IANA timezone for interpreting schedule times (e.g., America/New_York)';

-- ============================================================================
-- SCHEDULE_DAYS TABLE (Mon-Sun shift definitions for a schedule)
-- ============================================================================
CREATE TABLE schedule_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    is_working_day BOOLEAN NOT NULL DEFAULT true,
    shift_start TIME,
    shift_end TIME,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_schedule_day UNIQUE (schedule_id, day_of_week),
    CONSTRAINT chk_shift_times CHECK (
        (is_working_day = false AND shift_start IS NULL AND shift_end IS NULL) OR
        (is_working_day = true AND shift_start IS NOT NULL AND shift_end IS NOT NULL)
    )
);

CREATE INDEX idx_schedule_days_schedule ON schedule_days(schedule_id);

COMMENT ON TABLE schedule_days IS 'Daily shift definitions for each day of the week within a schedule';
COMMENT ON COLUMN schedule_days.day_of_week IS 'Day of week: 0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday';
COMMENT ON COLUMN schedule_days.shift_start IS 'Shift start time (HH:MM:SS). NULL if not a working day';
COMMENT ON COLUMN schedule_days.shift_end IS 'Shift end time (HH:MM:SS). Can be less than shift_start for overnight shifts. NULL if not a working day';

-- ============================================================================
-- SCHEDULE_BREAKS TABLE (Breaks within a day shift)
-- ============================================================================
CREATE TABLE schedule_breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_day_id UUID NOT NULL REFERENCES schedule_days(id) ON DELETE CASCADE,
    name VARCHAR(100),
    break_start TIME NOT NULL,
    break_end TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_break_times CHECK (break_start < break_end)
);

CREATE INDEX idx_schedule_breaks_day ON schedule_breaks(schedule_day_id);

COMMENT ON TABLE schedule_breaks IS 'Breaks within a shift for a specific day';
COMMENT ON COLUMN schedule_breaks.name IS 'Optional friendly name for the break (e.g., Lunch, Morning Break)';

-- ============================================================================
-- SCHEDULE_HOLIDAYS TABLE (Dates when NO production occurs)
-- ============================================================================
CREATE TABLE schedule_holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    holiday_date DATE NOT NULL,
    name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_schedule_holiday UNIQUE (schedule_id, holiday_date)
);

CREATE INDEX idx_schedule_holidays_schedule ON schedule_holidays(schedule_id);
CREATE INDEX idx_schedule_holidays_date ON schedule_holidays(holiday_date);

COMMENT ON TABLE schedule_holidays IS 'Specific dates when no production occurs for this schedule';

-- ============================================================================
-- SCHEDULE_EXCEPTIONS TABLE (Date ranges with different weekly schedules)
-- Applies to ALL lines using this schedule
-- ============================================================================
CREATE TABLE schedule_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_exception_dates CHECK (start_date <= end_date)
);

CREATE INDEX idx_schedule_exceptions_schedule ON schedule_exceptions(schedule_id);
CREATE INDEX idx_schedule_exceptions_dates ON schedule_exceptions(start_date, end_date);

COMMENT ON TABLE schedule_exceptions IS 'Date ranges where a different weekly schedule applies to ALL lines using this schedule';

-- ============================================================================
-- SCHEDULE_EXCEPTION_DAYS TABLE (Mon-Sun for exception periods)
-- ============================================================================
CREATE TABLE schedule_exception_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_id UUID NOT NULL REFERENCES schedule_exceptions(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    is_working_day BOOLEAN NOT NULL DEFAULT true,
    shift_start TIME,
    shift_end TIME,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_exception_day UNIQUE (exception_id, day_of_week),
    CONSTRAINT chk_exception_shift_times CHECK (
        (is_working_day = false AND shift_start IS NULL AND shift_end IS NULL) OR
        (is_working_day = true AND shift_start IS NOT NULL AND shift_end IS NOT NULL)
    )
);

CREATE INDEX idx_schedule_exception_days_exception ON schedule_exception_days(exception_id);

COMMENT ON TABLE schedule_exception_days IS 'Daily shift definitions for exception periods';

-- ============================================================================
-- SCHEDULE_EXCEPTION_BREAKS TABLE (Breaks for exception days)
-- ============================================================================
CREATE TABLE schedule_exception_breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_day_id UUID NOT NULL REFERENCES schedule_exception_days(id) ON DELETE CASCADE,
    name VARCHAR(100),
    break_start TIME NOT NULL,
    break_end TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_exception_break_times CHECK (break_start < break_end)
);

CREATE INDEX idx_schedule_exception_breaks_day ON schedule_exception_breaks(exception_day_id);

COMMENT ON TABLE schedule_exception_breaks IS 'Breaks within exception day shifts';

-- ============================================================================
-- LINE_SCHEDULE_EXCEPTIONS TABLE (Line-specific exceptions)
-- Only applies to specific lines
-- ============================================================================
CREATE TABLE line_schedule_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_line_exception_dates CHECK (start_date <= end_date)
);

CREATE INDEX idx_line_schedule_exceptions_schedule ON line_schedule_exceptions(schedule_id);
CREATE INDEX idx_line_schedule_exceptions_dates ON line_schedule_exceptions(start_date, end_date);

COMMENT ON TABLE line_schedule_exceptions IS 'Date ranges where a different schedule applies to SPECIFIC lines only';

-- ============================================================================
-- LINE_SCHEDULE_EXCEPTION_LINES TABLE (Which lines get the exception)
-- ============================================================================
CREATE TABLE line_schedule_exception_lines (
    exception_id UUID NOT NULL REFERENCES line_schedule_exceptions(id) ON DELETE CASCADE,
    line_id UUID NOT NULL REFERENCES production_lines(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (exception_id, line_id)
);

CREATE INDEX idx_line_exception_lines_line ON line_schedule_exception_lines(line_id);

COMMENT ON TABLE line_schedule_exception_lines IS 'Junction table linking line-specific exceptions to the lines they apply to';

-- ============================================================================
-- LINE_SCHEDULE_EXCEPTION_DAYS TABLE (Mon-Sun for line-specific exceptions)
-- ============================================================================
CREATE TABLE line_schedule_exception_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_id UUID NOT NULL REFERENCES line_schedule_exceptions(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    is_working_day BOOLEAN NOT NULL DEFAULT true,
    shift_start TIME,
    shift_end TIME,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_line_exception_day UNIQUE (exception_id, day_of_week),
    CONSTRAINT chk_line_exception_shift_times CHECK (
        (is_working_day = false AND shift_start IS NULL AND shift_end IS NULL) OR
        (is_working_day = true AND shift_start IS NOT NULL AND shift_end IS NOT NULL)
    )
);

CREATE INDEX idx_line_exception_days_exception ON line_schedule_exception_days(exception_id);

COMMENT ON TABLE line_schedule_exception_days IS 'Daily shift definitions for line-specific exception periods';

-- ============================================================================
-- LINE_SCHEDULE_EXCEPTION_BREAKS TABLE (Breaks for line-specific exception days)
-- ============================================================================
CREATE TABLE line_schedule_exception_breaks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exception_day_id UUID NOT NULL REFERENCES line_schedule_exception_days(id) ON DELETE CASCADE,
    name VARCHAR(100),
    break_start TIME NOT NULL,
    break_end TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_line_exception_break_times CHECK (break_start < break_end)
);

CREATE INDEX idx_line_exception_breaks_day ON line_schedule_exception_breaks(exception_day_id);

COMMENT ON TABLE line_schedule_exception_breaks IS 'Breaks within line-specific exception day shifts';

-- ============================================================================
-- ADD SCHEDULE_ID TO PRODUCTION_LINES
-- ============================================================================
ALTER TABLE production_lines
    ADD COLUMN schedule_id UUID REFERENCES schedules(id) ON DELETE SET NULL;

CREATE INDEX idx_production_lines_schedule ON production_lines(schedule_id) WHERE deleted_at IS NULL;

COMMENT ON COLUMN production_lines.schedule_id IS 'The schedule this line follows. NULL means no schedule assigned';
