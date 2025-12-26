include .env

.PHONY: help build run test clean migrate-up migrate-down sqlc docker-up docker-down

DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
# Docker variables
IMAGE_NAME := doit-api
IMAGE_TAG := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell git describe --tags --always)

# Default target
help:
	@echo "Available targets:"
	@echo ""
	@echo "🏗️  Build & Run:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  migrate-up     - Run database migrations"
	@echo "  migrate-down   - Rollback database migrations"
	@echo "  migrate-down-last - Rollback last database migration"
	@echo "  migrate-up-last - Run last database migration"
	@echo "  migrate-up-to - Run database migrations up to a specific version"
	@echo "  migrate-down-to - Rollback database migrations down to a specific version"
	@echo "  migrate-version - Show current database migration version"
	@echo "  migrate-status - Show migration status"
	@echo "  migrate-fix    - Fix dirty migration (usage: make migrate-fix version=N)"
	@echo "  migrate-reset  - Reset migration tracking (keeps data)"
	@echo "  migrate-fresh  - Drop all tables and rerun migrations (⚠️  deletes data!)"
	@echo "  migrate-create - Create new migration (usage: make migrate-create name=add_column)"
	@echo "  db-check       - Check database connection"
	@echo "  db-tables      - Show all database tables"
	@echo ""
	@echo "🌱 Database Seeding:"
	@echo "  seed           - Run all seed files"
	@echo "  seed-dev       - Run development seed files only"
	@echo "  seed-test      - Run test seed files only"
	@echo "  setup          - Complete setup (database + migrations + seeds)"
	@echo "  dev            - Full dev workflow (setup + run)"
	@echo ""
	@echo "📦 Code Generation:"
	@echo "  sqlc           - Generate sqlc code"
	@echo "  generate-mocks - Generate mock files for testing"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  swagger-fmt    - Format Swagger comments"
	@echo ""
	@echo "🔍 Code Quality:"
	@echo "  lint           - Run golangci-lint"
	@echo "  lint-fix       - Run golangci-lint with auto-fix"
	@echo "  security       - Run security checks (gosec)"
	@echo "  vuln-check     - Check for vulnerabilities (govulncheck)"
	@echo ""
	@echo "🐳 Docker (Single Container):"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-build-no-cache - Build Docker image without cache"
	@echo "  docker-run     - Run Docker container locally"
	@echo "  docker-inspect - Inspect Docker image labels and metadata"
	@echo "  docker-size    - Show Docker image size"
	@echo "  docker-shell   - Get shell access to container (if using alpine)"
	@echo "  docker-clean   - Remove Docker images"
	@echo ""
	@echo "🐙 Docker Compose (Full Stack):"
	@echo "  compose-up     - Start all services (API, DB, Redis, Prometheus, Grafana)"
	@echo "  compose-up-build - Build and start all services"
	@echo "  compose-down   - Stop all services"
	@echo "  compose-down-v - Stop all services and remove volumes (⚠️  deletes data!)"
	@echo "  compose-logs   - View logs from all services"
	@echo "  compose-logs-api - View logs from API only"
	@echo "  compose-ps     - Show running services"
	@echo "  compose-restart - Restart all services"
	@echo "  compose-restart-api - Restart API only"
	@echo "  compose-shell-api - Get shell in API container"
	@echo "  compose-shell-db - Get shell in PostgreSQL container"
	@echo "  compose-migrate-up - Run migrations in docker-compose environment"
	@echo "  compose-migrate-down - Rollback migrations in docker-compose environment"
	@echo "  compose-setup  - Setup .env file for docker-compose"
	@echo "  compose-health - Check health of all services"
	@echo ""
	@echo "🛠️  Development Database:"
	@echo "  dev-db         - Start development database container"
	@echo "  dev-db-stop    - Stop development database container"
	@echo ""
	@echo "🚀 Local Development (Hybrid Mode - FAST!):"
	@echo "  dev-setup      - Initial setup for local development"
	@echo "  dev-start      - Start everything (infra + migrate + API)"
	@echo "  dev-infra      - Start infrastructure only (DB, Redis, Jaeger, etc.)"
	@echo "  dev-infra-down - Stop infrastructure"
	@echo "  dev-infra-down-v - Stop infrastructure and remove volumes"
	@echo "  dev-infra-ps   - Show infrastructure status"
	@echo "  dev-infra-logs - View infrastructure logs"
	@echo "  dev-migrate    - Run migrations against local Docker DB"
	@echo "  dev-run        - Run Go API locally (connects to Docker infra)"
	@echo "  dev-run-race   - Run Go API with race detector"
	@echo "  dev-build      - Build Go API binary"
	@echo "  dev-seed       - Seed database with development data"
	@echo "  dev-test       - Run tests with local infrastructure"
	@echo "  dev-health     - Check health of all services"
	@echo "  dev-restart    - Restart infrastructure"
	@echo "  dev-clean      - Clean up development environment"

# Build the application
build:
	go build -o bin/doit ./cmd/doit

# Run the application
run:
	go run ./cmd/doit/main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Generate sqlc code
sqlc:
	$(shell go env GOPATH)/bin/sqlc generate

# Install sqlc (if not installed)
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golang-migrate (if not installed)
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install mockgen (if not installed)
install-mockgen:
	go install go.uber.org/mock/mockgen@latest

# Generate mocks for testing
generate-mocks:
	@./scripts/generate-mocks.sh

# Generate Swagger documentation
swagger:
	@echo "📚 Generating Swagger documentation..."
	@command -v swag >/dev/null 2>&1 || (echo "❌ swag not found. Run 'make install-swag' first." && exit 1)
	swag init -g cmd/doit/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger docs generated successfully!"
	@echo "📖 View at: http://localhost:8080/swagger/index.html (after running the app)"

# Format Swagger comments
swagger-fmt:
	@echo "🎨 Formatting Swagger comments..."
	@command -v swag >/dev/null 2>&1 || (echo "❌ swag not found. Run 'make install-swag' first." && exit 1)
	swag fmt
	@echo "✅ Swagger comments formatted!"

# Install swag CLI tool
install-swag:
	@echo "📥 Installing swag..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ swag installed successfully!"

# Database migrations
migrate-up:
	migrate -path internal/data/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path internal/data/migrations -database "$(DB_URL)" down

migrate-down-last:
	migrate -path internal/data/migrations -database "$(DB_URL)" down 1

migrate-up-last:
	migrate -path internal/data/migrations -database "$(DB_URL)" up 1

migrate-up-to:
	migrate -path internal/data/migrations -database "$(DB_URL)" up $(version)

migrate-down-to:
	migrate -path internal/data/migrations -database "$(DB_URL)" down $(version)

migrate-force:
	migrate -path internal/data/migrations -database "$(DB_URL)" force $(version)

migrate-version:
	migrate -path internal/data/migrations -database "$(DB_URL)" version

migrate-create:
	migrate create -ext sql -dir internal/data/migrations -seq ${name}

# Fix dirty migration by forcing to a specific version
# Usage: make migrate-fix version=1
migrate-fix:
	@echo "Current version:"
	@migrate -path internal/data/migrations -database "$(DB_URL)" version || true
	@echo "\nForcing to version $(version)..."
	migrate -path internal/data/migrations -database "$(DB_URL)" force $(version)
	@echo "\nNew version:"
	@migrate -path internal/data/migrations -database "$(DB_URL)" version

# Reset migrations completely (drops schema_migrations table)
# WARNING: This does NOT drop your tables, only resets migration tracking
migrate-reset:
	@echo "⚠️  This will reset migration tracking (drops schema_migrations table)"
	@echo "Your data tables will NOT be affected."
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker exec -it doit_db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP TABLE IF EXISTS schema_migrations;"; \
		echo "✅ Migration tracking reset. Run 'make migrate-up' to reapply."; \
	fi

# Fresh start: drop all tables and rerun migrations
# WARNING: This WILL delete all your data!
migrate-fresh:
	@echo "⚠️  WARNING: This will DROP ALL TABLES and data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker exec -it doit_db psql -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		$(MAKE) migrate-up; \
		echo "✅ Database recreated and migrations applied."; \
	fi

# Check database connection
db-check:
	@docker exec -it doit_db psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT version();"

# Show all tables
db-tables:
	@docker exec -it doit_db psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

# Show migration status with better formatting
migrate-status:
	@echo "📊 Migration Status:"
	@echo "===================="
	@migrate -path internal/data/migrations -database "$(DB_URL)" version || echo "❌ Error reading version"
	@echo ""
	@echo "📁 Available migrations:"
	@ls -1 internal/data/migrations/*.up.sql | sed 's/.*\//  - /'

# Docker commands

## Build Docker image with metadata
docker-build: 
	docker build \
		-f infra/docker/dockerfile.service \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		.

docker-build-no-cache: ## Build Docker image without cache
	docker build --no-cache \
		-f infra/docker/dockerfile.service \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(IMAGE_TAG) \
		-t $(IMAGE_NAME):latest \
		.

docker-run: ## Run Docker container locally
	docker run --rm -it \
		-p 8080:8080 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=5432 \
		-e DB_USER=islamghany \
		-e DB_PASSWORD=secret \
		-e DB_NAME=doit \
		-e DB_SSL_MODE=disable \
		-e REDIS_ADDR=host.docker.internal:6379 \
		$(IMAGE_NAME):latest


docker-inspect: ## Inspect Docker image labels and metadata
	@docker inspect $(IMAGE_NAME):latest | jq '.[0].Config.Labels'

docker-size: ## Show Docker image size
	@docker images $(IMAGE_NAME):latest --format "Size: {{.Size}}"

docker-shell: ## Get shell access to container (if using alpine)
	docker run --rm -it --entrypoint /bin/sh $(IMAGE_NAME):latest

docker-clean: ## Remove Docker images
	docker rmi $(IMAGE_NAME):latest $(IMAGE_NAME):$(IMAGE_TAG) || true

# ============================================
# Docker Compose Commands
# ============================================

## Start all services in detached mode
compose-up:
	@echo "🚀 Starting all services..."
	docker-compose up -d
	@echo ""
	@echo "✅ All services started!"
	@echo ""
	@echo "📋 Service URLs:"
	@echo "  🔹 API:         http://localhost:8080"
	@echo "  🔹 Swagger:     http://localhost:8080/swagger/index.html"
	@echo "  🔹 Health:      http://localhost:8080/health"
	@echo "  🔹 Metrics:     http://localhost:8080/metrics"
	@echo "  🔹 Grafana:     http://localhost:3000 (admin/admin)"
	@echo "  🔹 Prometheus:  http://localhost:9090"
	@echo "  🔹 Adminer:     http://localhost:8081 (use --profile tools)"
	@echo ""
	@echo "📊 Check status: make compose-ps"
	@echo "📜 View logs:    make compose-logs"

## Build and start all services
compose-up-build:
	@echo "🔨 Building and starting all services..."
	docker-compose up -d --build
	@echo "✅ All services built and started!"

## Start services with logs visible (foreground)
compose-up-logs:
	@echo "🚀 Starting all services with logs..."
	docker-compose up

## Stop all services
compose-down:
	@echo "🛑 Stopping all services..."
	docker-compose down
	@echo "✅ All services stopped!"

## Stop all services and remove volumes (⚠️  deletes data!)
compose-down-v:
	@echo "⚠️  WARNING: This will delete all data (volumes)!"
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "✅ All services stopped and volumes removed!"; \
	fi

## View logs from all services
compose-logs:
	docker-compose logs -f

## View logs from API only
compose-logs-api:
	docker-compose logs -f api

## View logs from PostgreSQL
compose-logs-db:
	docker-compose logs -f postgres

## View logs from Redis
compose-logs-redis:
	docker-compose logs -f redis

## View logs from Prometheus
compose-logs-prometheus:
	docker-compose logs -f prometheus

## View logs from Grafana
compose-logs-grafana:
	docker-compose logs -f grafana

## Show running services
compose-ps:
	@echo "📊 Running Services:"
	@docker-compose ps
	@echo ""
	@echo "💡 Tip: Use 'make compose-health' to check health status"

## Restart all services
compose-restart:
	@echo "🔄 Restarting all services..."
	docker-compose restart
	@echo "✅ All services restarted!"

## Restart API only
compose-restart-api:
	@echo "🔄 Restarting API..."
	docker-compose restart api
	@echo "✅ API restarted!"

## Restart PostgreSQL
compose-restart-db:
	docker-compose restart postgres

## Restart Redis
compose-restart-redis:
	docker-compose restart redis

## Get shell in API container
compose-shell-api:
	docker-compose exec api /bin/sh

## Get shell in PostgreSQL container
compose-shell-db:
	docker-compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

## Get shell in Redis container
compose-shell-redis:
	docker-compose exec redis redis-cli

## Run migrations in docker-compose environment
compose-migrate-up:
	@echo "📦 Running migrations..."
	docker-compose exec api migrate -path /app/internal/data/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" up
	@echo "✅ Migrations complete!"

## Rollback migrations in docker-compose environment
compose-migrate-down:
	@echo "📦 Rolling back migrations..."
	docker-compose exec api migrate -path /app/internal/data/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable" down 1
	@echo "✅ Rollback complete!"

## Setup .env file for docker-compose
compose-setup:
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "✅ .env file created!"; \
		echo "⚠️  Please update .env with your actual configuration"; \
	else \
		echo "ℹ️  .env file already exists"; \
	fi

## Check health of all services
compose-health:
	@echo "🏥 Checking service health..."
	@echo ""
	@echo "📊 API Health:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ API not responding"
	@echo ""
	@echo "📊 PostgreSQL Health:"
	@docker-compose exec -T postgres pg_isready -U $(DB_USER) -d $(DB_NAME) || echo "❌ PostgreSQL not ready"
	@echo ""
	@echo "📊 Redis Health:"
	@docker-compose exec -T redis redis-cli ping || echo "❌ Redis not responding"
	@echo ""
	@echo "📊 Prometheus Health:"
	@curl -s http://localhost:9090/-/healthy || echo "❌ Prometheus not responding"
	@echo ""
	@echo "📊 Grafana Health:"
	@curl -s http://localhost:3000/api/health | jq . || echo "❌ Grafana not responding"

## Pull latest images
compose-pull:
	@echo "📥 Pulling latest images..."
	docker-compose pull
	@echo "✅ Images updated!"

## Show resource usage
compose-stats:
	@echo "📊 Resource Usage:"
	docker stats --no-stream $$(docker-compose ps -q)

## Start with tools profile (includes Adminer)
compose-up-tools:
	@echo "🚀 Starting all services with tools..."
	docker-compose --profile tools up -d
	@echo "✅ All services started (including tools)!"
	@echo "  🔹 Adminer: http://localhost:8081"

# Development database setup
dev-db:
	docker run --name doit-postgres \
		-e POSTGRES_USER=doit \
		-e POSTGRES_PASSWORD=doit123 \
		-e POSTGRES_DB=doit \
		-p 5432:5432 \
		-d postgres:16-alpine

# Stop development database
dev-db-stop:
	docker stop doit-postgres || true
	docker rm doit-postgres || true

# Database seeding
seed:
	go run ./cmd/seed/main.go

seed-dev:
	go run ./cmd/seed/main.go --env=dev

seed-test:
	go run ./cmd/seed/main.go --env=test

# Complete setup (database, migrations, seeds)
setup: dev-db migrate-up seed-dev
	@echo "✅ Database setup complete!"

# Development workflow
dev: setup run

# Linting and code quality
lint:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint not found. Run 'make install-lint' first." && exit 1)
	$(shell go env GOPATH)/bin/golangci-lint run ./...

lint-fix:
	@command -v golangci-lint >/dev/null 2>&1 || (echo "❌ golangci-lint not found. Run 'make install-lint' first." && exit 1)
	$(shell go env GOPATH)/bin/golangci-lint run --fix ./...

# Install golangci-lint
install-lint:
	@command -v golangci-lint >/dev/null 2>&1 || \
	(echo "Installing golangci-lint..." && \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin)

# Security checks
security:
	@command -v gosec >/dev/null 2>&1 || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	$(shell go env GOPATH)/bin/gosec -fmt=json -out=gosec-report.json -no-fail ./...
	@echo "✅ Security scan complete. Report saved to gosec-report.json"

# Vulnerability check
vuln-check:
	@command -v govulncheck >/dev/null 2>&1 || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	$(shell go env GOPATH)/bin/govulncheck ./...

# Run all checks (CI simulation)
ci: lint test security vuln-check
	@echo "✅ All CI checks passed!"


## Metrics and Tracing

metrics-view-sc:
	expvarmon -ports="localhost:8001" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

metrics-view:
	expvarmon -ports="localhost:4020" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"

# ===========================================
# 🚀 Local Development (Hybrid Mode)
# ===========================================
# Run infrastructure in Docker, Go API locally
# This is MUCH faster for development iterations!
# ===========================================

## dev-infra: Start only infrastructure (DB, Redis, Jaeger, Prometheus, Grafana)
.PHONY: dev-infra
dev-infra:
	@echo "🚀 Starting infrastructure services..."
	docker-compose -f docker-compose.infra.yml up 
	@echo ""
	@echo "✅ Infrastructure started!"
	@echo ""
	@echo "📍 Services available at:"
	@echo "   PostgreSQL:  localhost:5432"
	@echo "   Redis:       localhost:6379"
	@echo "   Jaeger UI:   http://localhost:16686"
	@echo "   Prometheus:  http://localhost:9090"
	@echo "   Grafana:     http://localhost:3000 (admin/admin)"
	@echo "   Adminer:     http://localhost:8081"
	@echo ""
	@echo "📝 Next steps:"
	@echo "   1. Run migrations: make dev-migrate"
	@echo "   2. Start API:      make dev-run"

## dev-infra-down: Stop infrastructure services
.PHONY: dev-infra-down
dev-infra-down:
	@echo "🛑 Stopping infrastructure services..."
	docker-compose -f docker-compose.infra.yml down
	@echo "✅ Infrastructure stopped!"

## dev-infra-down-v: Stop infrastructure and remove volumes (⚠️  deletes data!)
.PHONY: dev-infra-down-v
dev-infra-down-v:
	@echo "🛑 Stopping infrastructure and removing volumes..."
	docker-compose -f docker-compose.infra.yml down -v
	@echo "✅ Infrastructure stopped and volumes removed!"

## dev-infra-logs: View infrastructure logs
.PHONY: dev-infra-logs
dev-infra-logs:
	docker-compose -f docker-compose.infra.yml logs -f

## dev-infra-ps: Show infrastructure status
.PHONY: dev-infra-ps
dev-infra-ps:
	docker-compose -f docker-compose.infra.yml ps

## dev-migrate: Run migrations against local Docker DB
.PHONY: dev-migrate
dev-migrate:
	@echo "🗄️  Running migrations..."
	migrate -path internal/data/migrations -database "postgresql://doit:doit123@localhost:5432/doit?sslmode=disable" up
	@echo "✅ Migrations complete!"

## dev-migrate-down: Rollback migrations
.PHONY: dev-migrate-down
dev-migrate-down:
	@echo "🗄️  Rolling back migrations..."
	migrate -path internal/data/migrations -database "postgresql://doit:doit123@localhost:5432/doit?sslmode=disable" down 1
	@echo "✅ Rollback complete!"

## dev-migrate-reset: Reset all migrations (⚠️  deletes all data!)
.PHONY: dev-migrate-reset
dev-migrate-reset:
	@echo "⚠️  Resetting all migrations..."
	migrate -path internal/data/migrations -database "postgresql://doit:doit123@localhost:5432/doit?sslmode=disable" drop -f
	migrate -path internal/data/migrations -database "postgresql://doit:doit123@localhost:5432/doit?sslmode=disable" up
	@echo "✅ Migrations reset!"

## dev-run: Run Go API locally (connects to Docker infrastructure)
.PHONY: dev-run
dev-run:
	@echo "🚀 Starting API locally..."
	@echo "   Make sure infrastructure is running: make dev-infra"
	@echo ""
	@if [ -f .env.local ]; then \
		set -a && . ./.env.local && set +a && go run ./cmd/doit; \
	else \
		echo "⚠️  .env.local not found!"; \
		echo "   Creating from template..."; \
		cp env.local.example .env.local; \
		echo "✅ Created .env.local - you can customize it if needed"; \
		set -a && . ./.env.local && set +a && go run ./cmd/doit; \
	fi

## dev-run-race: Run Go API locally with race detector
.PHONY: dev-run-race
dev-run-race:
	@echo "🚀 Starting API locally with race detector..."
	@if [ -f .env.local ]; then \
		set -a && . ./.env.local && set +a && go run -race ./cmd/doit; \
	else \
		cp env.local.example .env.local; \
		set -a && . ./.env.local && set +a && go run -race ./cmd/doit; \
	fi

## dev-build: Build Go API for local testing
.PHONY: dev-build
dev-build:
	@echo "🔨 Building API..."
	go build -o bin/doit ./cmd/doit
	@echo "✅ Built: bin/doit"

## dev-start: Start infrastructure, run migrations, and start API (all-in-one)
.PHONY: dev-start
dev-start: dev-infra
	@echo "⏳ Waiting for services to be healthy..."
	@sleep 5
	@$(MAKE) dev-migrate
	@$(MAKE) dev-run

## dev-setup: Initial setup for local development
.PHONY: dev-setup
dev-setup:
	@echo "🔧 Setting up local development environment..."
	@if [ ! -f .env.local ]; then \
		cp env.local.example .env.local; \
		echo "✅ Created .env.local"; \
	else \
		echo "ℹ️  .env.local already exists"; \
	fi
	@echo ""
	@echo "📝 Quick Start:"
	@echo "   1. make dev-infra      # Start infrastructure"
	@echo "   2. make dev-migrate    # Run migrations"
	@echo "   3. make dev-run        # Start API"
	@echo ""
	@echo "   Or all at once:"
	@echo "   make dev-start"

## dev-seed: Seed database with development data
.PHONY: dev-seed
dev-seed:
	@echo "🌱 Seeding database..."
	@if [ -f .env.local ]; then \
		set -a && . ./.env.local && set +a && go run ./cmd/seed; \
	else \
		cp env.local.example .env.local; \
		set -a && . ./.env.local && set +a && go run ./cmd/seed; \
	fi
	@echo "✅ Seeding complete!"

## dev-test: Run tests with local infrastructure
.PHONY: dev-test
dev-test:
	@echo "🧪 Running tests..."
	@if [ -f .env.local ]; then \
		set -a && . ./.env.local && set +a && go test -v -race ./...; \
	else \
		go test -v -race ./...; \
	fi

## dev-clean: Clean up development environment
.PHONY: dev-clean
dev-clean: dev-infra-down
	@echo "🧹 Cleaning up..."
	rm -rf bin/ tmp/
	@echo "✅ Cleanup complete!"

## dev-restart: Restart infrastructure
.PHONY: dev-restart
dev-restart:
	@echo "🔄 Restarting infrastructure..."
	docker-compose -f docker-compose.infra.yml restart
	@echo "✅ Infrastructure restarted!"

## dev-health: Check health of all infrastructure services
.PHONY: dev-health
dev-health:
	@echo "🏥 Checking infrastructure health..."
	@echo ""
	@echo "PostgreSQL:"
	@docker-compose -f docker-compose.infra.yml exec -T postgres pg_isready -U doit -d doit 2>/dev/null && echo "  ✅ Healthy" || echo "  ❌ Not healthy"
	@echo ""
	@echo "Redis:"
	@docker-compose -f docker-compose.infra.yml exec -T redis redis-cli ping 2>/dev/null | grep -q PONG && echo "  ✅ Healthy" || echo "  ❌ Not healthy"
	@echo ""
	@echo "Jaeger:"
	@curl -s http://localhost:16686 > /dev/null && echo "  ✅ Healthy (http://localhost:16686)" || echo "  ❌ Not healthy"
	@echo ""
	@echo "Prometheus:"
	@curl -s http://localhost:9090/-/healthy > /dev/null && echo "  ✅ Healthy (http://localhost:9090)" || echo "  ❌ Not healthy"
	@echo ""
	@echo "Grafana:"
	@curl -s http://localhost:3000/api/health > /dev/null && echo "  ✅ Healthy (http://localhost:3000)" || echo "  ❌ Not healthy"

