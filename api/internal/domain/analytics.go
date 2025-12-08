package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsQuery represents common parameters for analytics queries
type AnalyticsQuery struct {
	StartTime *time.Time  `json:"start_time,omitempty"`
	EndTime   *time.Time  `json:"end_time,omitempty"`
	LabelIDs  []uuid.UUID `json:"label_ids,omitempty"`
	LineIDs   []uuid.UUID `json:"line_ids,omitempty"`
	Timeframe string      `json:"timeframe,omitempty"` // "24h", "7d", "30d", "custom"
}

// AggregateMetrics represents overall metrics across all queried lines
type AggregateMetrics struct {
	TotalLines         int                `json:"total_lines"`
	TotalUptime        float64            `json:"total_uptime_hours"`
	AverageUptime      float64            `json:"average_uptime_percentage"`
	TotalDowntime      float64            `json:"total_downtime_hours"`
	TotalMaintenance   float64            `json:"total_maintenance_hours"`
	MTTR               float64            `json:"mttr_hours"` // Mean Time To Repair
	TotalInterruptions int                `json:"total_interruptions"`
	StatusDistribution map[Status]float64 `json:"status_distribution"` // Percentage by status
	TimeRange          TimeRange          `json:"time_range"`
}

// LineMetrics represents metrics for a single production line
type LineMetrics struct {
	LineID             uuid.UUID          `json:"line_id"`
	LineCode           string             `json:"line_code"`
	LineName           string             `json:"line_name"`
	Labels             []Label            `json:"labels"`
	UptimeHours        float64            `json:"uptime_hours"`
	UptimePercentage   float64            `json:"uptime_percentage"`
	DowntimeHours      float64            `json:"downtime_hours"`
	MaintenanceHours   float64            `json:"maintenance_hours"`
	ErrorHours         float64            `json:"error_hours"`
	MTTR               float64            `json:"mttr_hours"`
	InterruptionCount  int                `json:"interruption_count"`
	CurrentStatus      Status             `json:"current_status"`
	StatusDistribution map[Status]float64 `json:"status_distribution"`
}

// LabelMetrics represents metrics grouped by label
type LabelMetrics struct {
	Label              Label              `json:"label"`
	LineCount          int                `json:"line_count"`
	AverageUptime      float64            `json:"average_uptime_percentage"`
	TotalUptime        float64            `json:"total_uptime_hours"`
	TotalInterruptions int                `json:"total_interruptions"`
	MTTR               float64            `json:"mttr_hours"`
	StatusDistribution map[Status]float64 `json:"status_distribution"`
}

// DailyKPI represents KPIs for a single 24-hour period
type DailyKPI struct {
	Date             string  `json:"date"` // YYYY-MM-DD format
	UptimeHours      float64 `json:"uptime_hours"`
	UptimePercentage float64 `json:"uptime_percentage"`
	MaintenanceHours float64 `json:"maintenance_hours"`
	InterruptionCount int     `json:"interruption_count"`
	MTTR             float64 `json:"mttr_hours"`
}

// TimeRange represents a time period
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AnalyticsResponse wraps analytics data with metadata
type AnalyticsResponse struct {
	Query    AnalyticsQuery `json:"query"`
	Metrics  interface{}    `json:"metrics"` // Can be AggregateMetrics, []LineMetrics, etc.
	CachedAt *time.Time     `json:"cached_at,omitempty"`
}

// ========== Schedule Compliance Types ==========

// ComplianceQuery represents parameters for compliance queries
type ComplianceQuery struct {
	StartDate string      `json:"start_date"` // YYYY-MM-DD format
	EndDate   string      `json:"end_date"`   // YYYY-MM-DD format
	LabelIDs  []uuid.UUID `json:"label_ids,omitempty"`
	LineIDs   []uuid.UUID `json:"line_ids,omitempty"`
}

// LineComplianceMetrics represents compliance metrics for a single line
type LineComplianceMetrics struct {
	LineID                 uuid.UUID `json:"line_id"`
	LineCode               string    `json:"line_code"`
	LineName               string    `json:"line_name"`
	ScheduleID             *uuid.UUID `json:"schedule_id,omitempty"`
	ScheduleName           *string    `json:"schedule_name,omitempty"`
	ScheduledUptimeHours   float64   `json:"scheduled_uptime_hours"`
	ActualUptimeHours      float64   `json:"actual_uptime_hours"`
	CompliancePercentage   float64   `json:"compliance_percentage"`
	UnplannedDowntimeHours float64   `json:"unplanned_downtime_hours"`
	OvertimeHours          float64   `json:"overtime_hours"`
	ScheduledDays          int       `json:"scheduled_days"`
	WorkingDays            int       `json:"working_days"`
}

// AggregateComplianceMetrics represents overall compliance metrics
type AggregateComplianceMetrics struct {
	TotalLines             int                     `json:"total_lines"`
	LinesWithSchedule      int                     `json:"lines_with_schedule"`
	TotalScheduledHours    float64                 `json:"total_scheduled_hours"`
	TotalActualHours       float64                 `json:"total_actual_hours"`
	AverageCompliance      float64                 `json:"average_compliance_percentage"`
	TotalUnplannedDowntime float64                 `json:"total_unplanned_downtime_hours"`
	TotalOvertime          float64                 `json:"total_overtime_hours"`
	DateRange              DateRange               `json:"date_range"`
	LineMetrics            []LineComplianceMetrics `json:"line_metrics,omitempty"`
}

// DateRange represents a date period (without time)
type DateRange struct {
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD
}

// DailyComplianceKPI represents compliance for a single day
type DailyComplianceKPI struct {
	Date                   string  `json:"date"` // YYYY-MM-DD
	ScheduledUptimeHours   float64 `json:"scheduled_uptime_hours"`
	ActualUptimeHours      float64 `json:"actual_uptime_hours"`
	CompliancePercentage   float64 `json:"compliance_percentage"`
	UnplannedDowntimeHours float64 `json:"unplanned_downtime_hours"`
	OvertimeHours          float64 `json:"overtime_hours"`
	IsWorkingDay           bool    `json:"is_working_day"`
	Source                 string  `json:"source"` // "base", "holiday", "exception", "line_exception"
}
