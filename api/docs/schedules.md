# Schedule Management System

## Overview

The schedule management system defines when production lines should be operating. It supports complex scheduling scenarios with base weekly templates, holidays, exceptions, and line-specific overrides.

## Core Concepts

### Base Weekly Schedules

A base schedule defines the normal operating hours for production lines across a week.

**Features**:
- 7-day template (Sunday through Saturday)
- Per-day configuration: working/non-working, shift times, breaks
- Timezone support for multi-location facilities
- Reusable across multiple production lines

**Example**:
- Monday-Friday: 08:00-17:00 with 1-hour lunch break
- Saturday-Sunday: Non-working days

### Holidays

Specific dates when production does not occur, regardless of the base schedule.

**Features**:
- Manual entry (date + name)
- Auto-import from OpenHolidays API
- Applies to all lines using the schedule
- Higher priority than base weekly schedule

**Example**: December 25 (Christmas), January 1 (New Year's Day)

### Schedule Exceptions

Date range overrides that affect **ALL lines** using a schedule.

**Use Case**: Seasonal changes, extended hours during peak periods, facility-wide shutdowns

**Features**:
- Date range (start_date, end_date)
- Modified weekly schedule (overrides base schedule)
- Applies to all lines globally

**Example**:
- December 1-31: Extend hours to Monday-Saturday 08:00-20:00 for holiday rush

### Line Schedule Exceptions

Date range overrides that affect **specific production lines** only.

**Use Case**: Line-specific maintenance, critical orders requiring overtime, staggered breaks

**Features**:
- Date range (start_date, end_date)
- Modified weekly schedule
- Applies only to selected lines
- Highest priority override

**Example**:
- LINE-001 only: December 15-20 operates 24/7 for critical order

## Schedule Priority System

When determining the effective schedule for a specific line on a specific date, the system checks in this order:

```
1. Line-specific exception (highest priority)
   └─ If exists for date → Use this
2. Schedule exception (global)
   └─ If exists for date → Use this
3. Holiday
   └─ If date is holiday → Non-working day
4. Base weekly schedule (lowest priority)
   └─ Use configured day-of-week schedule
```

## Data Model

### Schedules Table
```sql
CREATE TABLE schedules (
  id UUID PRIMARY KEY,
  name VARCHAR(255),
  description TEXT,
  timezone VARCHAR(50),
  created_at TIMESTAMPTZ,
  updated_at TIMESTAMPTZ,
  deleted_at TIMESTAMPTZ
);
```

### Schedule Days (Base Weekly Schedule)
```sql
CREATE TABLE schedule_days (
  id UUID PRIMARY KEY,
  schedule_id UUID REFERENCES schedules(id),
  day_of_week INT,  -- 0=Sunday, 6=Saturday
  is_working_day BOOLEAN,
  shift_start TIME,
  shift_end TIME
);
```

### Schedule Breaks
```sql
CREATE TABLE schedule_breaks (
  id UUID PRIMARY KEY,
  schedule_day_id UUID REFERENCES schedule_days(id),
  break_start TIME,
  break_end TIME,
  break_name VARCHAR(100)  -- e.g., "Lunch", "Coffee Break"
);
```

### Holidays
```sql
CREATE TABLE schedule_holidays (
  id UUID PRIMARY KEY,
  schedule_id UUID REFERENCES schedules(id),
  holiday_date DATE,
  holiday_name VARCHAR(255)
);
```

### Schedule Exceptions (Global)
```sql
CREATE TABLE schedule_exceptions (
  id UUID PRIMARY KEY,
  schedule_id UUID REFERENCES schedules(id),
  start_date DATE,
  end_date DATE,
  exception_name VARCHAR(255),
  -- Contains modified schedule_days for this date range
);
```

### Line Schedule Exceptions (Specific Lines)
```sql
CREATE TABLE line_schedule_exceptions (
  id UUID PRIMARY KEY,
  line_id UUID REFERENCES production_lines(id),
  schedule_id UUID REFERENCES schedules(id),
  start_date DATE,
  end_date DATE,
  exception_name VARCHAR(255),
  -- Contains modified schedule_days for specific lines
);
```

## API Endpoints

### Schedule CRUD

**List Schedules**
```http
GET /api/v1/schedules
Response: [{ id, name, description, timezone, created_at }]
```

**Get Schedule**
```http
GET /api/v1/schedules/{id}
Response: {
  id, name, description, timezone,
  days: [{ day_of_week, is_working_day, shift_start, shift_end, breaks }],
  holidays: [{ date, name }],
  exceptions: [{ start_date, end_date, name, days }]
}
```

**Create Schedule**
```http
POST /api/v1/schedules
Body: {
  name: "Standard Week",
  description: "Mon-Fri 8-5",
  timezone: "America/New_York",
  days: [
    { day_of_week: 1, is_working_day: true, shift_start: "08:00", shift_end: "17:00",
      breaks: [{ break_start: "12:00", break_end: "13:00", break_name: "Lunch" }]
    },
    ...
  ],
  holidays: [{ holiday_date: "2025-12-25", holiday_name: "Christmas" }]
}
```

**Update Schedule**
```http
PUT /api/v1/schedules/{id}
Body: Same as create
```

**Delete Schedule**
```http
DELETE /api/v1/schedules/{id}
Response: 204 No Content
```

### Effective Schedule

**Get Effective Schedule**
```http
GET /api/v1/schedules/{id}/effective?start_date=2025-12-01&end_date=2025-12-31
Response: {
  dates: [
    {
      date: "2025-12-25",
      is_working_day: false,
      reason: "Holiday: Christmas"
    },
    {
      date: "2025-12-15",
      is_working_day: true,
      shift_start: "08:00",
      shift_end: "20:00",
      breaks: [...],
      reason: "Schedule Exception: Holiday Rush"
    }
  ]
}
```

**Get Line Effective Schedule** (includes line-specific exceptions)
```http
GET /api/v1/lines/{line_id}/schedule/effective?start_date=2025-12-01&end_date=2025-12-31
Response: Same format, but includes line-specific exception overrides
```

### Holidays

**Add Holiday**
```http
POST /api/v1/schedules/{id}/holidays
Body: { holiday_date: "2025-12-25", holiday_name: "Christmas" }
```

**Import Holidays from OpenHolidays API**
```http
POST /api/v1/schedules/{id}/holidays/import
Body: { country_code: "US", year: 2025 }
Response: { imported_count: 10, holidays: [...] }
```

**Delete Holiday**
```http
DELETE /api/v1/schedules/{id}/holidays/{holiday_id}
```

### Exceptions

**Add Schedule Exception** (global)
```http
POST /api/v1/schedules/{id}/exceptions
Body: {
  start_date: "2025-12-01",
  end_date: "2025-12-31",
  exception_name: "Holiday Rush",
  days: [{ day_of_week: 1, is_working_day: true, shift_start: "08:00", shift_end: "20:00" }]
}
```

**Add Line Exception** (specific line)
```http
POST /api/v1/lines/{line_id}/schedule/exceptions
Body: Same as schedule exception
```

## Usage Example

### Scenario: Manufacturing Facility

**Base Schedule: "Standard Week"**
- Monday-Friday: 08:00-17:00 (1-hour lunch 12:00-13:00)
- Saturday-Sunday: Non-working

**Holiday: Christmas**
- December 25: No production

**Global Exception: "Holiday Rush"**
- December 1-31: Monday-Saturday 08:00-20:00 (extended hours)

**Line Exception: "Critical Order - LINE-001"**
- December 15-20: LINE-001 operates 24/7 (continuous)

**Effective Schedule Results**:

| Date | LINE-001 | LINE-002 (Standard) | Reason |
|------|----------|---------------------|--------|
| Dec 1 | 08:00-20:00 | 08:00-20:00 | Schedule Exception (Holiday Rush) |
| Dec 15 | 24/7 (continuous) | 08:00-20:00 | Line Exception for LINE-001 |
| Dec 25 | Non-working | Non-working | Holiday (Christmas) |
| Jan 2 | 08:00-17:00 | 08:00-17:00 | Base Schedule |

## Compliance Integration

Schedules are used by the **Compliance System** to calculate:
- Scheduled uptime hours (when lines should be running)
- Compliance percentage (actual vs scheduled uptime)
- Unplanned downtime
- Overtime hours (running outside scheduled hours)

See [Compliance Documentation](./compliance.md) for details.

## See Also

- [Compliance Calculation Methodology](./compliance.md)
- [API Swagger Documentation](http://localhost:8080/swagger/index.html)
- [Architecture Overview](../../docs/architecture.md)
