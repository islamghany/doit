# 🚀 Quick Setup Guide for sqlc + pgx

Follow these steps to get your database-backed API running.

## Prerequisites

- Go 1.24.2 or higher
- Docker (for PostgreSQL)
- Make (optional, for convenience)

## Step 1: Install Required Tools

```bash
# Install sqlc (for code generation)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golang-migrate (for database migrations)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Verify installations
sqlc version
migrate -version
```

## Step 2: Start PostgreSQL Database

### Option A: Using Make (recommended)

```bash
make dev-db
```

### Option B: Manual Docker Command

```bash
docker run --name doit-postgres \
  -e POSTGRES_USER=doit \
  -e POSTGRES_PASSWORD=doit123 \
  -e POSTGRES_DB=doit \
  -p 5432:5432 \
  -d postgres:16-alpine
```

### Verify Database is Running

```bash
docker ps | grep doit-postgres
```

## Step 3: Create Environment File

Create a `.env` file in the project root:

```bash
cat > .env << 'EOF'
# Server Configuration
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10
SERVER_WRITE_TIMEOUT=10
SERVER_IDLE_TIMEOUT=120
SERVER_SHUTDOWN_TIMEOUT=25

# Application Configuration
APP_ENVIRONMENT=development
APP_LOG_LEVEL=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=doit
DB_USER=doit
DB_PASSWORD=doit123
DB_DISABLE_TLS=true
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=300
EOF
```

## Step 4: Run Migrations

**Note**: Currently, golang-migrate needs migrations in a specific format. Since we have our SQL schema in `001_initial_schema.sql`, we need to split it or manually apply it first:

### Option A: Apply Schema Manually

```bash
# Connect to PostgreSQL
docker exec -it doit-postgres psql -U doit -d doit

# Then copy/paste the contents of internal/data/migrations/001_initial_schema.sql
# Or run it directly:
docker exec -i doit-postgres psql -U doit -d doit < internal/data/migrations/001_initial_schema.sql
```

### Option B: Use golang-migrate (if you split the migration)

```bash
# Run migrations
migrate -path internal/data/migrations \
        -database "postgresql://doit:doit123@localhost:5432/doit?sslmode=disable" \
        up
```

## Step 5: Generate sqlc Code

This generates type-safe Go code from your SQL queries:

```bash
# Using Make
make sqlc

# Or directly
sqlc generate
```

**Expected output:**

```
internal/data/db/
├── db.go
├── models.go
├── querier.go
├── activity_logs.sql.go
├── todos.sql.go
└── users.sql.go
```

## Step 6: Verify the Setup

Check that the generated files exist:

```bash
ls -la internal/data/db/
```

You should see:

- `db.go` - Database connection helpers
- `models.go` - Generated struct types
- `querier.go` - Query interface
- `*.sql.go` - Generated query functions

## Step 7: Run the Application

```bash
# Using Make
make run

# Or directly
go run cmd/doit/main.go
```

**Expected output:**

```
level=info msg="Starting application" build=development version=v0.0.1
level=info msg="Database connection established" host=localhost database=doit
level=info msg="Starting server" host=localhost port=8080
```

## Step 8: Test the API

### Health Check

```bash
curl http://localhost:8080/healthcheck
```

**Expected response:**

```json
{
  "status": "ok",
  "version": "v1.0.0"
}
```

### Create a User (once you integrate handlers)

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "testuser",
    "password": "password123",
    "role": "user"
  }'
```

## Troubleshooting

### Error: "failed to connect to database"

- Check if PostgreSQL is running: `docker ps`
- Verify connection details in `.env`
- Check database logs: `docker logs doit-postgres`

### Error: "no required module provides package doit/internal/data/db"

- Run `sqlc generate` to create the db package
- Ensure `sqlc.yaml` is configured correctly

### Error: "table does not exist"

- Run migrations (Step 4)
- Verify tables exist: `docker exec -it doit-postgres psql -U doit -d doit -c "\dt"`

### Error: "sqlc: command not found"

- Install sqlc: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- Ensure `$GOPATH/bin` is in your `PATH`

## Development Workflow

### Making Database Changes

1. **Modify Schema**

   ```bash
   # Edit or create new migration file
   vim internal/data/migrations/002_add_new_table.sql
   ```

2. **Apply Migration**

   ```bash
   docker exec -i doit-postgres psql -U doit -d doit < internal/data/migrations/002_add_new_table.sql
   ```

3. **Add Queries**

   ```bash
   # Add queries to appropriate file
   vim internal/data/queries/your_entity.sql
   ```

4. **Regenerate Code**

   ```bash
   make sqlc
   ```

5. **Use in Services**
   ```go
   // The new queries are automatically available
   result, err := service.queries.YourNewQuery(ctx, params)
   ```

### Running Tests

```bash
make test
```

### Viewing Database

```bash
# Connect to PostgreSQL
docker exec -it doit-postgres psql -U doit -d doit

# List tables
\dt

# Describe a table
\d users

# Query data
SELECT * FROM users;

# Exit
\q
```

### Cleaning Up

```bash
# Stop and remove database
make dev-db-stop

# Clean build artifacts
make clean
```

## Next Steps

Now that your setup is complete, you can:

1. **Integrate handlers** - See `INTEGRATION_EXAMPLE.md`
2. **Add authentication** - Implement JWT middleware
3. **Add more queries** - Extend queries in `internal/data/queries/`
4. **Write tests** - Test your services
5. **Deploy** - Use `infra/docker/` or `infra/k8s/`

## Reference Files

- `README_SQLC_PGX.md` - Advanced features and best practices
- `INTEGRATION_EXAMPLE.md` - How to integrate services with handlers
- `Makefile` - Available commands
- `sqlc.yaml` - sqlc configuration

## Useful Commands

```bash
# Development
make run              # Run application
make build            # Build binary
make test             # Run tests
make sqlc             # Generate sqlc code

# Database
make dev-db           # Start PostgreSQL
make dev-db-stop      # Stop PostgreSQL
make migrate-up       # Run migrations
make migrate-down     # Rollback migrations

# Cleanup
make clean            # Remove build artifacts
```

## Support

For questions or issues:

1. Check `README_SQLC_PGX.md` for advanced topics
2. Review error messages carefully
3. Check PostgreSQL logs: `docker logs doit-postgres`
4. Verify environment variables: `echo $SERVER_PORT`

**You're all set! Start building! 🚀**
