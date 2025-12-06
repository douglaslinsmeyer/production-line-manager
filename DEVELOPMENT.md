# Development Guide

This document provides detailed information for developers working on the Production Line API.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Local Development Setup](#local-development-setup)
- [Running the Application](#running-the-application)
- [Database Migrations](#database-migrations)
- [Testing](#testing)
- [Code Style and Linting](#code-style-and-linting)
- [MQTT Integration](#mqtt-integration)
- [API Documentation](#api-documentation)
- [Debugging](#debugging)

---

## Prerequisites

- **Go 1.22+**: [Download](https://golang.org/dl/)
- **Docker & Docker Compose**: [Download](https://www.docker.com/products/docker-desktop)
- **golang-migrate CLI** (optional): For manual migration management
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint** (optional): For code linting
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **swag** (optional): For generating Swagger docs
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

---

## Project Structure

```
production-line-api/
├── cmd/
│   └── server/               # Application entry point
│       └── main.go           # Main function with wiring
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection setup
│   ├── domain/               # Domain types, errors, events
│   ├── handler/              # HTTP handlers and router
│   ├── logger/               # Logging configuration
│   ├── mqtt/                 # MQTT client, publisher, subscriber
│   ├── repository/           # Data access layer (pgx)
│   └── service/              # Business logic layer
├── migrations/               # Database migration files
│   ├── 001_*.sql            # Production lines table
│   └── 002_*.sql            # TimescaleDB hypertable for status log
├── docs/                     # Generated Swagger documentation
├── .github/workflows/        # CI/CD pipelines
├── Dockerfile                # Production Docker image
├── docker-compose.yml        # Local development stack
├── mosquitto.conf            # MQTT broker configuration
├── Makefile                  # Build and development commands
└── README.md                 # User documentation
```

### Layer Responsibilities

- **Handler**: HTTP request/response handling, routing, validation
- **Service**: Business logic, orchestration, transaction management
- **Repository**: Database queries, data access
- **Domain**: Core types, business rules, constants

---

## Local Development Setup

### Option 1: Docker Compose (Recommended)

Start all services (PostgreSQL, MQTT, API):

```bash
make docker-up
```

This starts:
- **PostgreSQL + TimescaleDB** on `localhost:5432`
- **MQTT (Mosquitto)** on `localhost:1883` (TCP) and `localhost:9001` (WebSocket)
- **API** on `localhost:8080`

### Option 2: Manual Setup

1. **Start PostgreSQL + TimescaleDB**:
   ```bash
   docker run -d \
     --name production-line-db \
     -p 5432:5432 \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=postgres \
     -e POSTGRES_DB=production_lines \
     timescale/timescaledb:latest-pg15
   ```

2. **Start MQTT Broker**:
   ```bash
   docker run -d \
     --name production-line-mqtt \
     -p 1883:1883 \
     -p 9001:9001 \
     -v $(pwd)/mosquitto.conf:/mosquitto/config/mosquitto.conf \
     eclipse-mosquitto:2
   ```

3. **Set environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your local configuration
   ```

4. **Run database migrations**:
   ```bash
   make migrate-up
   ```

5. **Start the API**:
   ```bash
   make run
   ```

---

## Running the Application

### Using Make

```bash
# Build the binary
make build

# Run the server (generates docs first)
make run

# Run with Docker Compose
make docker-up
```

### Using Go commands

```bash
# Generate Swagger docs
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Run directly
go run cmd/server/main.go

# Build binary
go build -o bin/server cmd/server/main.go
```

### Environment Variables

Required:
- `DATABASE_URL`: PostgreSQL connection string
- `MQTT_BROKER_URL`: MQTT broker address

Optional:
- `PORT`: HTTP server port (default: `8080`)
- `LOG_LEVEL`: Logging level (default: `info`)
- `MQTT_CLIENT_ID`: MQTT client identifier (default: `production-line-api`)
- `MQTT_QOS`: MQTT QoS level (default: `1`)
- `SHUTDOWN_TIMEOUT`: Graceful shutdown timeout (default: `10s`)

---

## Database Migrations

### Creating a New Migration

```bash
make migrate-create name=add_some_feature
```

This creates two files:
- `migrations/XXX_add_some_feature.up.sql` - Apply migration
- `migrations/XXX_add_some_feature.down.sql` - Rollback migration

### Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down
```

### Manual Migration Management

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" version
```

### TimescaleDB-Specific Features

The status log table uses TimescaleDB features:
- **Hypertable**: Automatic time-based partitioning (7-day chunks)
- **Compression**: Chunks older than 30 days are compressed
- **Retention** (optional): Automatically drop chunks older than specified period

---

## Testing

### Run All Tests

```bash
make test
```

### Run with Coverage

```bash
make test-coverage
# Opens coverage.html in browser
```

### Test Structure

```
internal/
├── repository/
│   ├── lines_test.go         # Repository unit tests
│   └── status_log_test.go
├── service/
│   └── lines_test.go          # Service unit tests (with mocks)
└── handler/
    └── lines_test.go          # Handler integration tests
```

### Testing Best Practices

1. **Repository tests**: Use testcontainers for real database
2. **Service tests**: Mock repository layer
3. **Handler tests**: Use `httptest.Server`
4. **Integration tests**: Full stack with docker-compose

---

## Code Style and Linting

### Format Code

```bash
make fmt
```

### Run Linters

```bash
make lint
```

### Pre-commit Checklist

Before committing:
```bash
make fmt
make lint
make test
make build
```

---

## MQTT Integration

### Topic Structure

**Published by API (Events)**:
- `production-lines/events/created` - Line created (QoS 0)
- `production-lines/events/updated` - Line updated (QoS 0)
- `production-lines/events/deleted` - Line deleted (QoS 0)
- `production-lines/events/status` - Status changed (QoS 0)

**Subscribed by API (Commands)**:
- `production-lines/commands/status` - Set line status (QoS 1)

### Testing MQTT Locally

Using `mosquitto_pub` and `mosquitto_sub`:

**Subscribe to events**:
```bash
mosquitto_sub -h localhost -p 1883 -t "production-lines/events/#" -v
```

**Publish status command**:
```bash
mosquitto_pub -h localhost -p 1883 -t "production-lines/commands/status" \
  -m '{"code": "LINE-A1", "status": "on"}'
```

### MQTT Connection Management

- Auto-reconnect enabled with exponential backoff
- QoS 1 for commands (at-least-once delivery)
- QoS 0 for events (best-effort, occasional miss acceptable)
- Connection status reflected in logs

---

## API Documentation

### Generate Swagger Docs

```bash
make docs
```

### View Documentation

1. Start the server
2. Navigate to http://localhost:8080/swagger/index.html

### Adding Swagger Annotations

Example handler annotation:
```go
// Create godoc
// @Summary Create production line
// @Description Create a new production line
// @Tags lines
// @Accept json
// @Produce json
// @Param line body domain.CreateLineRequest true "Production line details"
// @Success 201 {object} Response{data=domain.ProductionLine}
// @Failure 400 {object} Response{error=APIError}
// @Router /lines [post]
func (h *LineHandler) Create(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

---

## Debugging

### Enabling Debug Logging

Set `LOG_LEVEL=debug` in environment:

```bash
LOG_LEVEL=debug make run
```

Debug logs include:
- Request/response details
- SQL queries with parameters
- MQTT message payloads
- Service layer decisions

### Database Debugging

Connect to PostgreSQL:
```bash
psql postgres://postgres:postgres@localhost:5432/production_lines
```

Useful queries:
```sql
-- View all production lines
SELECT * FROM production_lines WHERE deleted_at IS NULL;

-- View recent status changes
SELECT * FROM production_line_status_log
ORDER BY time DESC
LIMIT 20;

-- Check TimescaleDB hypertable info
SELECT * FROM timescaledb_information.hypertables;

-- Check compression policy
SELECT * FROM timescaledb_information.compression_settings;
```

### MQTT Debugging

Monitor all MQTT traffic:
```bash
mosquitto_sub -h localhost -p 1883 -t "#" -v
```

### Common Issues

1. **"Connection refused" on startup**: Ensure PostgreSQL and MQTT are running
2. **Migration errors**: Check if migrations were already applied (`migrate version`)
3. **MQTT not receiving commands**: Verify topic spelling and QoS settings
4. **TimescaleDB extension not found**: Use `timescale/timescaledb` image, not vanilla PostgreSQL

---

## Performance Considerations

### Database

- Connection pooling: Max 25 connections, Min 5
- Use prepared statements (pgx handles this automatically)
- Indexes on frequently queried columns
- TimescaleDB compression for old data

### MQTT

- Event publishing is best-effort (failures logged, don't block requests)
- Buffered message queue during disconnections
- Auto-reconnect with exponential backoff

### HTTP

- Request timeout: 15 seconds
- Write timeout: 15 seconds
- Idle timeout: 60 seconds
- Compression middleware enabled

---

## Contributing

1. Create a feature branch from `develop`
2. Make your changes
3. Run tests and linters
4. Commit with conventional commit messages
5. Open a pull request

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat(service): add idempotent status changes

Status changes are now idempotent - setting a line to its current
status returns success without logging a change event.

Closes #123
```

---

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [pgx Documentation](https://github.com/jackc/pgx)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [MQTT Specification](https://mqtt.org/mqtt-specification/)
- [Swagger/OpenAPI](https://swagger.io/specification/)
