# Compliance Calculation Methodology

## Overview

The compliance system measures how well production lines adhere to their defined schedules. It calculates scheduled uptime, actual uptime, and derives compliance percentage along with other key performance indicators (KPIs).

## Core Metrics

### 1. Scheduled Uptime Hours

**Definition**: The total hours a production line is scheduled to operate during a given date range.

**Calculation**:
```
For each day in date range:
  1. Get effective schedule for line and date (considers exceptions)
  2. If working day:
     - shift_duration = shift_end - shift_start
     - break_duration = sum(all breaks)
     - scheduled_hours = shift_duration - break_duration
  3. If non-working day (holiday, weekend):
     - scheduled_hours = 0
  4. Add to total scheduled_uptime_hours
```

**Example**:
- Date range: December 1-5 (5 days)
- Schedule: Mon-Fri 08:00-17:00 (9 hours), 1-hour lunch break
- Scheduled uptime per day: 9 - 1 = 8 hours
- Total scheduled uptime: 5 × 8 = 40 hours

### 2. Actual Uptime Hours

**Definition**: The total hours a production line actually operated (status = 'on') during the date range.

**Calculation**:
```sql
SELECT SUM(
  EXTRACT(EPOCH FROM (
    LEAD(time) OVER (PARTITION BY line_id ORDER BY time) - time
  )) / 3600
) as actual_uptime_hours
FROM production_line_status_log
WHERE line_id = ?
  AND time >= start_date
  AND time < end_date
  AND status = 'on';
```

**Logic**:
1. Query `production_line_status_log` TimescaleDB hypertable
2. For each status='on' period, calculate duration until next status change
3. Sum all 'on' durations in hours
4. Filter to date range

**Example**:
- December 1: 'on' for 7.5 hours
- December 2: 'on' for 8 hours
- December 3: 'on' for 6 hours (downtime)
- December 4: 'on' for 8 hours
- December 5: 'on' for 7.5 hours
- Total actual uptime: 37 hours

### 3. Compliance Percentage

**Definition**: The ratio of actual uptime to scheduled uptime, expressed as a percentage.

**Formula**:
```
compliance % = (actual_uptime_hours / scheduled_uptime_hours) × 100
```

**Interpretation**:
- 100%: Line ran exactly as scheduled
- >100%: Line ran overtime (more than scheduled)
- <100%: Line had downtime or ran less than scheduled

**Example** (from above):
```
compliance % = (37 / 40) × 100 = 92.5%
```

### 4. Unplanned Downtime Hours

**Definition**: Time when the line should be running but isn't (excluding planned maintenance).

**Calculation**:
```
unplanned_downtime = scheduled_uptime_hours
                   - actual_uptime_hours
                   - maintenance_hours

where maintenance_hours = sum of time with status='maintenance' during scheduled hours
```

**Example**:
- Scheduled: 40 hours
- Actual: 37 hours
- Maintenance: 1 hour (planned)
- Unplanned downtime: 40 - 37 - 1 = 2 hours

### 5. Maintenance Hours

**Definition**: Planned maintenance time during scheduled operating hours.

**Calculation**:
```sql
SELECT SUM(
  EXTRACT(EPOCH FROM (
    LEAD(time) OVER (PARTITION BY line_id ORDER BY time) - time
  )) / 3600
) as maintenance_hours
FROM production_line_status_log
WHERE line_id = ?
  AND time >= start_date
  AND time < end_date
  AND status = 'maintenance'
  AND time is within scheduled hours;
```

### 6. Overtime Hours

**Definition**: Time when the line operates outside scheduled hours.

**Calculation**:
```
If actual_uptime_hours > scheduled_uptime_hours:
  overtime = actual_uptime_hours - scheduled_uptime_hours
Else:
  overtime = 0
```

**Example**:
- Scheduled: 40 hours
- Actual: 45 hours
- Overtime: 5 hours

### 7. Working Days Count

**Definition**: Number of days within the date range that are working days per the schedule.

**Calculation**:
```
For each day in date range:
  If effective schedule for day is a working day:
    working_days++
```

## Data Sources

### Schedule Data
- **schedules**: Base weekly schedules
- **schedule_days**: Per-day shift configuration
- **schedule_holidays**: Holiday dates
- **schedule_exceptions**: Global overrides
- **line_schedule_exceptions**: Line-specific overrides

### Status Log Data
- **production_line_status_log**: TimescaleDB hypertable with all status changes
  - Columns: time, line_id, status, source
  - Partitioned by time (7-day chunks)
  - Compressed after 30 days

## API Endpoints

### Aggregate Compliance

**Get Overall Compliance**
```http
GET /api/v1/compliance/aggregate?start_date=2025-12-01&end_date=2025-12-31
Response: {
  scheduled_uptime_hours: 160,
  actual_uptime_hours: 148,
  compliance_percentage: 92.5,
  unplanned_downtime_hours: 8,
  maintenance_hours: 4,
  overtime_hours: 0,
  working_days: 20
}
```

### Per-Line Compliance

**Get Compliance by Line**
```http
GET /api/v1/compliance/lines?start_date=2025-12-01&end_date=2025-12-31
Response: [
  {
    line_id: "uuid",
    line_code: "LINE-001",
    scheduled_uptime_hours: 40,
    actual_uptime_hours: 37,
    compliance_percentage: 92.5,
    unplanned_downtime_hours: 2,
    maintenance_hours: 1,
    overtime_hours: 0
  },
  ...
]
```

### Daily Compliance Breakdown

**Get Daily Compliance for Specific Line**
```http
GET /api/v1/compliance/lines/{line_id}/daily?start_date=2025-12-01&end_date=2025-12-31
Response: {
  line_id: "uuid",
  line_code: "LINE-001",
  daily_compliance: [
    {
      date: "2025-12-01",
      scheduled_hours: 8,
      actual_hours: 7.5,
      compliance_percentage: 93.75,
      is_working_day: true
    },
    {
      date: "2025-12-02",
      scheduled_hours: 8,
      actual_hours: 8,
      compliance_percentage: 100.0,
      is_working_day: true
    },
    {
      date: "2025-12-25",
      scheduled_hours: 0,
      actual_hours: 0,
      compliance_percentage: null,
      is_working_day: false,
      note: "Holiday: Christmas"
    }
  ]
}
```

## Calculation Examples

### Example 1: Perfect Compliance

**Setup**:
- Schedule: Mon-Fri 08:00-17:00 (8 hours/day with 1-hour break)
- Date range: Week of December 1-5 (5 working days)
- Actual operation: Line ran exactly as scheduled

**Results**:
```
Scheduled uptime: 5 days × 8 hours = 40 hours
Actual uptime: 40 hours
Maintenance: 0 hours
Compliance: (40 / 40) × 100 = 100%
Unplanned downtime: 0 hours
Overtime: 0 hours
```

### Example 2: Downtime with Maintenance

**Setup**:
- Schedule: Same as Example 1 (40 hours)
- Actual: Line had 2 hours unplanned downtime, 1 hour planned maintenance

**Results**:
```
Scheduled uptime: 40 hours
Actual uptime: 37 hours
Maintenance: 1 hour
Compliance: (37 / 40) × 100 = 92.5%
Unplanned downtime: 40 - 37 - 1 = 2 hours
Overtime: 0 hours
```

### Example 3: Overtime Operation

**Setup**:
- Schedule: Same as Example 1 (40 hours)
- Actual: Line ran 45 hours (including weekends for critical order)

**Results**:
```
Scheduled uptime: 40 hours
Actual uptime: 45 hours
Maintenance: 0 hours
Compliance: (45 / 40) × 100 = 112.5%
Unplanned downtime: 0 hours
Overtime: 45 - 40 = 5 hours
```

### Example 4: Holiday Week

**Setup**:
- Schedule: Mon-Fri 08:00-17:00 (8 hours/day)
- Date range: December 23-27 (includes Christmas on Dec 25)
- Actual: Line shut down for holidays

**Results**:
```
Working days: 4 (Dec 25 is holiday)
Scheduled uptime: 4 × 8 = 32 hours
Actual uptime: 0 hours
Compliance: (0 / 32) × 100 = 0%
Unplanned downtime: 32 hours
Note: Lines typically shut down for holidays, so 0% compliance may be expected
```

## Performance Considerations

### TimescaleDB Optimization
- Hypertable partitioning enables fast time-range queries
- Compression reduces storage for historical data
- Continuous aggregates can pre-calculate daily summaries

### Caching Strategy
- Compliance calculations can be expensive for large date ranges
- Consider caching results for completed periods (e.g., past months)
- Invalidate cache when schedule or status log changes

### Query Optimization
- Use indexes on (line_id, time) in status_log
- Limit date ranges to reasonable periods (e.g., max 1 year)
- Consider async calculation for large reports

## Interpretation Guidelines

### Compliance Targets
- **>95%**: Excellent compliance
- **90-95%**: Good compliance
- **80-90%**: Fair compliance, investigate causes
- **<80%**: Poor compliance, requires immediate attention

### Investigating Low Compliance
1. Check unplanned_downtime_hours (equipment failures?)
2. Review status_log for error periods
3. Verify schedule accuracy (is schedule realistic?)
4. Check if overtime is compensating for downtime

### Handling Outliers
- Compliance >100% indicates overtime (may be intentional)
- Compliance 0% on working days indicates complete shutdown
- Null compliance on non-working days is normal

## See Also

- [Schedule Management System](./schedules.md)
- [TimescaleDB Hypertables](https://docs.timescale.com/use-timescale/latest/hypertables/)
- [Analytics Dashboard](../../web/docs/features.md#compliance)
- [API Swagger Documentation](http://localhost:8080/swagger/index.html)
