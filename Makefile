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
