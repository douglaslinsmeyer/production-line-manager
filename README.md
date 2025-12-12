# Production Line Manager

Complete full-stack application for managing production line operations in a manufacturing facility.

## Overview

This monorepo contains:
- **API** (`api/`): Go backend with PostgreSQL + TimescaleDB and MQTT integration
- **Web UI** (`web/`): React TypeScript dashboard with real-time monitoring and analytics

## Features

### Backend (Go API)
- REST API for production line CRUD operations
- Status management with audit trail (TimescaleDB)
- MQTT event publishing and command subscription
- Swagger/OpenAPI documentation
- Structured logging with Zap

### Frontend (React Web UI)
- Real-time dashboard with auto-refresh (TanStack Query)
- Production line management interface
- Status change monitoring with live updates
- Analytics:
  - Status history charts (Recharts)
  - Uptime metrics and MTTR
  - Status distribution visualization
  - Timeline of all changes
- Toast notifications for user feedback
- Responsive mobile-first design

## Quick Start

### Start Everything with Docker Compose

```bash
make docker-up
```

This starts:
- PostgreSQL + TimescaleDB (port 5432)
- MQTT Broker (ports 1883, 9001)
- Go API (port 8080)
- React Web UI (port 3000)
- MQTTX Web Client (port 8090)

Access points:
- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **MQTTX**: http://localhost:8090

### Development Mode (with hot reload)

```bash
make docker-up-dev
```

This provides:
- Web UI on port 5173 (Vite dev server with HMR)
- API with debug logging
- Auto-reload on file changes

## Repository Structure

```
production-line-manager/
├── api/                  # Go backend
│   ├── cmd/              # Application entry point
│   ├── internal/         # Go packages
│   ├── migrations/       # Database migrations
│   ├── Dockerfile        # Production image
│   └── README.md         # API documentation
├── web/                  # React frontend
│   ├── src/              # Source code
│   │   ├── api/          # API client (axios + types)
│   │   ├── components/   # UI components
│   │   ├── pages/        # Route components
│   │   ├── hooks/        # TanStack Query hooks
│   │   └── utils/        # Helpers
│   ├── Dockerfile        # Production image (Nginx)
│   ├── Dockerfile.dev    # Development image
│   └── README.md         # Web documentation
├── docker-compose.yml    # Production services
├── docker-compose.dev.yml # Development services
├── Makefile              # Build orchestration
└── README.md             # This file
```

## Technology Stack

### Backend
- Go 1.25
- Chi v5 (HTTP router)
- pgx v5 (PostgreSQL driver)
- TimescaleDB (time-series)
- Eclipse Paho MQTT
- Zap (structured logging)

### Frontend
- React 18 + TypeScript
- Vite (build tool)
- TanStack Query (data fetching)
- React Router v6
- Tailwind CSS
- Recharts (analytics)
- Headless UI
- react-hot-toast

### Infrastructure
- PostgreSQL 15 + TimescaleDB
- RabbitMQ 3.13 (MQTT + AMQP)
- Docker + Docker Compose
- GitHub Actions CI/CD

## Development

### Build Both Services

```bash
make build
```

### Run Tests

```bash
make test
```

### Lint Code

```bash
make lint
```

### Clean Build Artifacts

```bash
make clean
```

## Docker Commands

```bash
# Production
make docker-up          # Start all services
make docker-down        # Stop all services

# Development (hot reload)
make docker-up-dev      # Start with volume mounts
```

## Status Model

Production lines support 4 statuses:
- `on` - Line is running
- `off` - Line is stopped
- `maintenance` - Line is under maintenance
- `error` - Line has encountered an error

## MQTT Integration

### Topics

**Published by API (Events)**:
- `production-lines/events/created`
- `production-lines/events/updated`
- `production-lines/events/deleted`
- `production-lines/events/status`

**Subscribed by API (Commands)**:
- `production-lines/commands/status`

## Documentation

### Getting Started
- **[Getting Started Guide](docs/getting-started.md)**: Setup and development environment
- **[Architecture Overview](docs/architecture.md)**: System architecture and components
- **[MQTT Topics](docs/mqtt/topics.md)**: Complete MQTT topic reference
- **[MQTT Messages](docs/mqtt/message-formats.md)**: Message schemas and formats

### Component Documentation
- **API**: See [api/README.md](api/README.md) and [api/docs/](api/docs/)
- **Firmware**: See [firmware/README.md](firmware/README.md) and [firmware/docs/](firmware/docs/)
- **Web UI**: See [web/README.md](web/README.md) and [web/docs/](web/docs/)

### API Reference
- **Swagger UI**: http://localhost:8080/swagger/index.html (when running)

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│   Web UI    │────▶│   Go API    │────▶│  PostgreSQL  │
│  (React)    │◀────│   (REST)    │◀────│+ TimescaleDB │
└─────────────┘     └──────┬──────┘     └──────────────┘
                           │
                           │ pub/sub
                           ▼
                    ┌──────────────┐
                    │ MQTT Broker  │
                    │  (RabbitMQ)  │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  ESP32-S3    │
                    │   Devices    │
                    └──────────────┘
```

## Contributing

1. Create feature branch
2. Make changes in `api/` or `web/`
3. Run `make test` and `make lint`
4. Commit and open PR

## License

Proprietary - All rights reserved
