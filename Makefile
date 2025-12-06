.PHONY: build run test test-coverage docs migrate-up migrate-down migrate-create docker-up docker-down docker-build lint fmt clean help

# Variables
BINARY_NAME=server
DOCKER_IMAGE=production-line-api
MIGRATIONS_PATH=migrations
DATABASE_URL?=postgres://postgres:postgres@localhost:5432/production_lines?sslmode=disable

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build            - Build the server binary"
	@echo "  run              - Run the server locally"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  docs             - Generate Swagger documentation"
	@echo "  migrate-up       - Run database migrations"
	@echo "  migrate-down     - Rollback last migration"
	@echo "  migrate-create   - Create new migration (usage: make migrate-create name=my_migration)"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-up        - Start all services with Docker Compose"
	@echo "  docker-down      - Stop all Docker services"
	@echo "  lint             - Run linters"
	@echo "  fmt              - Format code"
	@echo "  clean            - Remove build artifacts"

## build: Build the server binary
build: docs
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "Build complete: bin/$(BINARY_NAME)"

## run: Run the server locally (generates docs first)
run: docs
	@echo "Starting server..."
	@go run cmd/server/main.go

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## docs: Generate Swagger documentation
docs:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@swag fmt
	@echo "Swagger docs generated in docs/"

## migrate-up: Run database migrations
migrate-up:
	@echo "Running migrations..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up
	@echo "Migrations complete"

## migrate-down: Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	@migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1
	@echo "Rollback complete"

## migrate-create: Create new migration file
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migrate-create name=my_migration"; exit 1; fi
	@echo "Creating migration: $(name)"
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "Services started. API available at http://localhost:8080"

## docker-down: Stop all Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Services stopped"

## lint: Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@swag fmt

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf docs/docs.go docs/swagger.json docs/swagger.yaml
	@echo "Clean complete"
