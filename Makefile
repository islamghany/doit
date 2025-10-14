.PHONY: help build run test clean migrate-up migrate-down sqlc docker-up docker-down

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  sqlc         - Generate sqlc code"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  docker-up    - Start Docker containers"
	@echo "  docker-down  - Stop Docker containers"

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
	migrate -path internal/data/migrations -database "postgresql://localhost:5432/doit?sslmode=disable" up

migrate-down:
	migrate -path internal/data/migrations -database "postgresql://localhost:5432/doit?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir internal/data/migrations -seq ${name}

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
