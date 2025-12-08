-- Remove schedule_id from production_lines
ALTER TABLE production_lines DROP COLUMN IF EXISTS schedule_id;

-- Drop all schedule-related tables in reverse dependency order
DROP TABLE IF EXISTS line_schedule_exception_breaks;
DROP TABLE IF EXISTS line_schedule_exception_days;
DROP TABLE IF EXISTS line_schedule_exception_lines;
DROP TABLE IF EXISTS line_schedule_exceptions;
DROP TABLE IF EXISTS schedule_exception_breaks;
DROP TABLE IF EXISTS schedule_exception_days;
DROP TABLE IF EXISTS schedule_exceptions;
DROP TABLE IF EXISTS schedule_holidays;
DROP TABLE IF EXISTS schedule_breaks;
DROP TABLE IF EXISTS schedule_days;
DROP TABLE IF EXISTS schedules;
