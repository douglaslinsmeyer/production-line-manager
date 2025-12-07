.PHONY: help dev dev-api dev-web build build-api build-web test test-api test-web lint clean docker-up docker-up-dev docker-down

## help: Display this help message
help:
	@echo "Production Line Manager - Monorepo"
	@echo ""
	@echo "Available targets:"
	@echo "  dev              - Start all services in development mode (docker-compose.dev.yml)"
	@echo "  dev-api          - Start only API locally (requires DB and MQTT running)"
	@echo "  dev-web          - Start only web dev server locally"
	@echo "  build            - Build both API and web"
	@echo "  build-api        - Build API binary"
	@echo "  build-web        - Build web production bundle"
	@echo "  test             - Run tests for both API and web"
	@echo "  test-api         - Run API tests"
	@echo "  test-web         - Run web tests"
	@echo "  lint             - Run linters for both services"
	@echo "  docker-up        - Start all services with Docker Compose (production)"
	@echo "  docker-up-dev    - Start all services with Docker Compose (development)"
	@echo "  docker-down      - Stop all Docker services"
	@echo "  clean            - Remove build artifacts"

## dev: Start all services in development mode
dev:
	docker-compose -f docker-compose.dev.yml up

## dev-api: Start only API (assumes DB and MQTT are running)
dev-api:
	cd api && make run

## dev-web: Start only web dev server (assumes API is running)
dev-web:
	cd web && npm run dev

## build: Build both API and web
build: build-api build-web

## build-api: Build API binary
build-api:
	cd api && make build

## build-web: Build web production bundle
build-web:
	cd web && npm run build

## test: Run tests for both services
test: test-api test-web

## test-api: Run API tests
test-api:
	cd api && make test

## test-web: Run web tests
test-web:
	cd web && npm test

## lint: Run linters for both services
lint:
	@echo "Linting API..."
	cd api && make lint
	@echo "Linting web..."
	cd web && npm run lint

## docker-up: Start production services
docker-up:
	docker-compose up -d

## docker-up-dev: Start development services
docker-up-dev:
	docker-compose -f docker-compose.dev.yml up

## docker-down: Stop all services
docker-down:
	docker-compose down
	docker-compose -f docker-compose.dev.yml down

## clean: Remove build artifacts
clean:
	cd api && make clean
	cd web && rm -rf dist node_modules/.vite
	@echo "Clean complete"
