include .env

.PHONY: help build run test clean migrate-up migrate-down sqlc docker-up docker-down

DB_URL := postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
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
	@echo "  migrate-status - Show migration status"
	@echo "  migrate-fix    - Fix dirty migration (usage: make migrate-fix version=N)"
	@echo "  migrate-reset  - Reset migration tracking (keeps data)"
	@echo "  migrate-fresh  - Drop all tables and rerun migrations (⚠️  deletes data!)"
	@echo "  migrate-create - Create new migration (usage: make migrate-create name=add_column)"
	@echo "  db-check       - Check database connection"
	@echo "  db-tables      - Show all database tables"
	@echo ""
	@echo "📦 Code Generation:"
	@echo "  sqlc           - Generate sqlc code"
	@echo ""
	@echo "🐳 Docker:"
	@echo "  docker-up      - Start Docker containers"
	@echo "  docker-down    - Stop Docker containers"
	@echo "  dev-db         - Start development database container"
	@echo "  dev-db-stop    - Stop development database container"

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
	sqlc generate

# Install sqlc (if not installed)
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golang-migrate (if not installed)
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Database migrations
migrate-up:
	migrate -path internal/data/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path internal/data/migrations -database "$(DB_URL)" down

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
docker-up:
	docker-compose -f infra/docker/docker-compose.yml up -d

docker-down:
	docker-compose -f infra/docker/docker-compose.yml down

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
