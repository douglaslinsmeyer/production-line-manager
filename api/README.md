# Production Line API

A Go-based API backend for managing production line operations in a manufacturing facility.

## Features

- **REST API** for production line CRUD operations
- **Status Management** with audit trail using TimescaleDB
- **MQTT Integration** for event publishing and command subscription
- **Structured Logging** with Zap
- **OpenAPI Documentation** with Swagger UI
- **Docker Compose** setup for local development

## Tech Stack

- **Language**: Go 1.22+
- **HTTP Router**: chi v5
- **Database**: PostgreSQL 15 + TimescaleDB
- **MQTT**: Eclipse Paho MQTT client
- **Logging**: Uber Zap
- **Documentation**: Swaggo

## Status Model

Production lines support the following statuses:
- `on` - Line is running
- `off` - Line is stopped
- `maintenance` - Line is under maintenance
- `error` - Line has encountered an error

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Make (optional, for convenience commands)

## Quick Start

### Local Development with Docker

1. Start all services (PostgreSQL, MQTT, API):
```bash
make docker-up
```

2. Run migrations:
```bash
make migrate-up
```

3. Access the API:
- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health

### Local Development without Docker

1. Start PostgreSQL with TimescaleDB:
```bash
docker run -d \
  -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=production_lines \
  timescale/timescaledb:latest-pg15
```

2. Start MQTT broker:
```bash
docker run -d \
  -p 1883:1883 \
  -v $(pwd)/mosquitto.conf:/mosquitto/config/mosquitto.conf \
  eclipse-mosquitto:2
```

3. Set environment variables:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/production_lines?sslmode=disable"
export MQTT_BROKER_URL="tcp://localhost:1883"
```

4. Run migrations:
```bash
make migrate-up
```

5. Generate Swagger docs:
```bash
make docs
```

6. Run the server:
```bash
make run
```

## Available Make Commands

```bash
make build           # Build the binary
make run             # Run the server locally
make test            # Run tests
make test-coverage   # Run tests with coverage report
make docs            # Generate Swagger documentation
make migrate-up      # Run database migrations
make migrate-down    # Rollback last migration
make docker-up       # Start all services with Docker Compose
make docker-down     # Stop all Docker services
make lint            # Run linters
make fmt             # Format code
```

## API Endpoints

### Production Lines

- `GET /api/v1/lines` - List all production lines
- `GET /api/v1/lines/{id}` - Get a specific production line
- `POST /api/v1/lines` - Create a new production line
- `PUT /api/v1/lines/{id}` - Update a production line
- `DELETE /api/v1/lines/{id}` - Soft delete a production line
- `POST /api/v1/lines/{id}/status` - Update production line status
- `GET /api/v1/lines/{id}/status/history` - Get status change history

### Health

- `GET /health` - Health check endpoint

### Documentation

- `GET /swagger/index.html` - Swagger UI

## MQTT Topics

### Published by API (Events)

- `production-lines/events/created` - Line created
- `production-lines/events/updated` - Line updated
- `production-lines/events/deleted` - Line deleted
- `production-lines/events/status` - Status changed

### Subscribed by API (Commands)

- `production-lines/commands/status` - Set line status from shop floor controllers

## Configuration

Configuration is managed via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | HTTP server port |
| `LOG_LEVEL` | No | `info` | Logging level (debug, info, warn, error) |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `MQTT_BROKER_URL` | Yes | - | MQTT broker address (tcp://host:port) |
| `MQTT_CLIENT_ID` | No | `production-line-api` | MQTT client identifier |
| `MQTT_QOS` | No | `1` | MQTT Quality of Service (0, 1, 2) |
| `SHUTDOWN_TIMEOUT` | No | `10s` | Graceful shutdown timeout |

## Project Structure

```
production-line-api/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── app/             # Application lifecycle
│   ├── config/          # Configuration
│   ├── database/        # Database connection
│   ├── domain/          # Domain types and business logic
│   ├── handler/         # HTTP handlers
│   ├── logger/          # Logging setup
│   ├── mqtt/            # MQTT client and pub/sub
│   ├── repository/      # Data access layer
│   └── service/         # Business logic layer
├── migrations/          # Database migrations
├── scripts/             # Helper scripts
├── docs/                # Generated Swagger docs
├── Dockerfile           # Production Docker image
├── docker-compose.yml   # Local development stack
└── Makefile             # Build and dev commands
```

## Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

## Deployment

The application is designed to be deployed to a Kubernetes cluster. See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## License

Proprietary - All rights reserved
