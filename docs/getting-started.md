# Getting Started

This guide helps you set up the Assembly Line Manager system for local development.

## Prerequisites

### Required
- **Docker & Docker Compose**: [Download](https://www.docker.com/products/docker-desktop)
- **Go 1.22+**: [Download](https://golang.org/dl/) (for API development)
- **Node.js 18+**: [Download](https://nodejs.org/) (for Web UI development)
- **PlatformIO**: [Install](https://platformio.org/install/cli) (for firmware development)

### Optional Tools
- **golang-migrate CLI**: For manual database migration management
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- **golangci-lint**: For Go code linting
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```
- **swag**: For regenerating Swagger API documentation
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

## Quick Start (Docker Compose)

The fastest way to get everything running:

```bash
# Clone the repository
git clone <repository-url>
cd assembly-line-manager

# Start all services (production mode)
make docker-up
```

This starts:
- **PostgreSQL + TimescaleDB** on `localhost:5432`
- **RabbitMQ** with MQTT on `localhost:1883` (TCP), `localhost:9001` (WebSocket), `localhost:15672` (Management UI)
- **Go API** on `localhost:8080`
- **React Web UI** on `localhost:3000`
- **MQTTX Web Client** on `localhost:8090`

### Access Points

- **Web Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **API Documentation (Swagger)**: http://localhost:8080/swagger/index.html
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **MQTTX Web Client**: http://localhost:8090

## Development Mode (Hot Reload)

For active development with auto-reload:

```bash
make docker-up-dev
```

This provides:
- **Web UI** on `localhost:5173` (Vite dev server with HMR)
- **API** with debug logging and volume mounts
- Auto-reload on file changes

## Manual Setup

If you prefer to run services individually:

### 1. Start Infrastructure

```bash
# PostgreSQL + TimescaleDB
docker run -d \
  --name production-line-db \
  -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=production_lines \
  timescale/timescaledb:latest-pg15

# RabbitMQ with MQTT Plugin
docker run -d \
  --name production-line-mqtt \
  -p 1883:1883 \
  -p 9001:9001 \
  -p 15672:15672 \
  -p 5672:5672 \
  rabbitmq:3.13-management-alpine

# Enable MQTT plugins
docker exec production-line-mqtt rabbitmq-plugins enable rabbitmq_mqtt rabbitmq_web_mqtt
```

### 2. Run Database Migrations

```bash
cd api
make migrate-up
```

### 3. Start API

```bash
cd api
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/production_lines?sslmode=disable"
export MQTT_BROKER_URL="tcp://localhost:1883"
make run
```

### 4. Start Web UI

```bash
cd web
npm install
npm run dev
```

### 5. Flash Firmware (Optional)

```bash
cd firmware
pio run --target upload
pio device monitor
```

## Environment Variables

### API Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `DATABASE_URL` | *required* | PostgreSQL connection string |
| `MQTT_BROKER_URL` | *required* | RabbitMQ MQTT broker URL |
| `MQTT_CLIENT_ID` | `production-line-api` | MQTT client identifier |
| `MQTT_QOS` | `1` | MQTT QoS level (0, 1, or 2) |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed CORS origins |

### Web Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:8080/api/v1` | API base URL |
| `VITE_MQTT_WS_URL` | `ws://localhost:9001` | RabbitMQ WebSocket URL |

### Firmware Configuration

Firmware configuration is stored in device NVS. See [Firmware Device Config](../firmware/docs/device-config.md) for details.

## Database Migrations

### Creating New Migrations

```bash
cd api
make migrate-create name=add_some_feature
```

This creates:
- `migrations/XXX_add_some_feature.up.sql` - Apply migration
- `migrations/XXX_add_some_feature.down.sql` - Rollback migration

### Running Migrations

```bash
# Apply all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
migrate -path migrations -database "$DATABASE_URL" version
```

### TimescaleDB Features

The `production_line_status_log` table uses TimescaleDB:
- **Hypertable**: Automatic time-based partitioning (7-day chunks)
- **Compression**: Chunks older than 30 days are compressed
- **Indexes**: Optimized for time-series queries

## Testing

### Run All Tests

```bash
# Test everything
make test

# Test specific component
cd api && make test
cd web && npm test
cd firmware && pio test
```

### API Test Coverage

```bash
cd api
make test-coverage
# Opens coverage.html in browser
```

## Code Formatting and Linting

### API (Go)

```bash
cd api
make fmt    # Format code
make lint   # Run linters
```

### Web UI (TypeScript)

```bash
cd web
npm run lint    # ESLint
npm run format  # Prettier
```

### Firmware (C++)

```bash
cd firmware
pio check  # Static analysis
```

## MQTT Testing

### Using mosquitto_pub/mosquitto_sub

Subscribe to all events:
```bash
mosquitto_sub -h localhost -p 1883 -t "#" -v
```

Publish status command:
```bash
mosquitto_pub -h localhost -p 1883 -t "production-lines/commands/status" \
  -m '{"line_code":"LINE-001","status":"on","source":"manual"}'
```

### Using MQTTX Web Client

1. Open http://localhost:8090
2. Create connection:
   - Host: `mqtt://localhost`
   - Port: `1883`
   - Username: `guest`
   - Password: `guest`
3. Subscribe to topics
4. Publish test messages

## Debugging

### Enable Debug Logging

```bash
# API
LOG_LEVEL=debug make docker-up-dev

# Web UI (browser console)
# Debug logging is automatic in dev mode
```

### Database Debugging

Connect to PostgreSQL:
```bash
psql postgres://postgres:postgres@localhost:5432/production_lines
```

Useful queries:
```sql
-- View all production lines
SELECT * FROM production_lines WHERE deleted_at IS NULL;

-- Recent status changes
SELECT * FROM production_line_status_log
ORDER BY time DESC
LIMIT 20;

-- TimescaleDB hypertable info
SELECT * FROM timescaledb_information.hypertables;
```

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f web
docker-compose logs -f mqtt
```

### Firmware Serial Monitor

```bash
cd firmware
pio device monitor --baud 115200
```

## Common Issues

### Docker Compose Fails to Start

**Issue**: "Port already in use"
```bash
# Check what's using the port
lsof -i :8080  # API
lsof -i :5432  # PostgreSQL
lsof -i :1883  # MQTT

# Stop conflicting services or change ports in docker-compose.yml
```

### Database Connection Refused

**Issue**: API can't connect to PostgreSQL
```bash
# Check if PostgreSQL is running
docker ps | grep production-line-db

# Check PostgreSQL logs
docker logs production-line-db
```

### MQTT Not Receiving Messages

**Issue**: Devices/API not receiving MQTT messages
- Verify RabbitMQ is running: `docker ps | grep mqtt`
- Check RabbitMQ Management UI: http://localhost:15672
- Ensure MQTT plugins are enabled
- Verify topic names match exactly (case-sensitive)
- Check QoS levels

### TimescaleDB Extension Not Found

**Issue**: "extension timescaledb does not exist"
- Use `timescale/timescaledb` image, not vanilla PostgreSQL
- Ensure migrations ran successfully
- Check database logs

### Web UI API Calls Fail

**Issue**: CORS errors or connection refused
- Verify API is running: http://localhost:8080/health
- Check `CORS_ALLOWED_ORIGINS` environment variable
- Ensure `VITE_API_BASE_URL` is correct
- Check browser console for detailed errors

## Next Steps

1. **Read the Architecture**: See [Architecture Overview](./architecture.md)
2. **Explore MQTT Topics**: See [MQTT Topics Reference](./mqtt/topics.md)
3. **Review API Endpoints**: Visit http://localhost:8080/swagger/index.html
4. **Check Component Docs**:
   - [API Documentation](../api/README.md)
   - [Firmware Documentation](../firmware/README.md)
   - [Web UI Documentation](../web/README.md)

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [PlatformIO Documentation](https://docs.platformio.org/)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [RabbitMQ MQTT Plugin](https://www.rabbitmq.com/mqtt.html)
- [MQTT Specification](https://mqtt.org/mqtt-specification/)
