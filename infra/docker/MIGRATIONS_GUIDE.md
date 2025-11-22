# Database Migrations in Docker: Complete Guide

> **How to run database migrations in Docker/Docker Compose environments**

---

## 📋 Table of Contents

- [The Problem](#the-problem)
- [Solution Overview](#solution-overview)
- [Approach 1: Separate Migration Service](#approach-1-separate-migration-service-docker-compose)
- [Approach 2: Entrypoint Script](#approach-2-entrypoint-script-built-into-api)
- [Approach 3: Manual Migrations](#approach-3-manual-migrations)
- [Approach 4: Init Container](#approach-4-init-container-kubernetes-style)
- [Which Approach to Use](#which-approach-to-use)
- [Troubleshooting](#troubleshooting)

---

## The Problem

When you start your application with Docker Compose, you have this situation:

```
┌──────────────────────────────────────────────────────────┐
│  Docker Compose Starts:                                  │
│                                                          │
│  1. PostgreSQL container starts ✅                       │
│  2. PostgreSQL becomes healthy ✅                        │
│  3. API container starts ✅                              │
│  4. API tries to connect to database...                 │
│                                                          │
│     ❌ Tables don't exist!                               │
│     ❌ Schema not created!                               │
│     ❌ App crashes or returns errors!                    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**Why this happens:**

- Database is empty when first started
- Migrations haven't run yet
- Your app expects tables to exist

---

## Solution Overview

We need to run migrations **after** PostgreSQL is ready but **before** the API starts processing requests.

```
✅ Correct Flow:

1. PostgreSQL starts
2. PostgreSQL becomes healthy
3. 🔑 Migrations run (creates schema)
4. API starts
5. API works perfectly!
```

---

## Approach 1: Separate Migration Service (Docker Compose)

**Best for:** Development, simple deployments

### Implementation

We've added a `migrations` service to `docker-compose.yml`:

```yaml
services:
  # Migration service (runs first)
  migrations:
    image: migrate/migrate:latest
    container_name: doit-migrations
    depends_on:
      postgres:
        condition: service_healthy # Wait for DB
    environment:
      DATABASE_URL: "postgresql://${DB_USER:-doit}:${DB_PASSWORD:-doit123}@postgres:5432/${DB_NAME:-doit}?sslmode=disable"
    volumes:
      - ./internal/data/migrations:/migrations
    command:
      - "-path=/migrations"
      - "-database=${DATABASE_URL}"
      - "up"
    networks:
      - doit_network
    restart: on-failure

  # API service (starts after migrations)
  api:
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      migrations:
        condition: service_completed_successfully # 🔑 Wait for migrations!
    # ... rest of API config
```

### How It Works

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  postgres (healthy)                                     │
│      ↓                                                  │
│  migrations (runs once, exits)                          │
│      ↓                                                  │
│  api (starts only after migrations succeed)             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Usage

```bash
# Start everything (migrations run automatically)
make compose-up

# Check migration status
docker logs doit-migrations

# Migrations will show:
# ✓ 000001_create_users_table.up.sql
# ✓ 000002_create_todos_table.up.sql
# ✓ 000003_create_refresh_tokens_table.up.sql
```

### Pros & Cons

**Pros:**

- ✅ Simple to understand
- ✅ Migrations visible as separate service
- ✅ Easy to debug (check logs)
- ✅ Works with any Docker Compose setup

**Cons:**

- ⚠️ Extra container (though it exits after running)
- ⚠️ Need to manage migration image version
- ⚠️ Slightly more complex docker-compose.yml

---

## Approach 2: Entrypoint Script (Built into API)

**Best for:** Production, self-contained images

### Implementation

We've added an entrypoint script that runs before your application starts.

**Files:**

- `infra/docker/entrypoint.sh` - Script that runs migrations
- `infra/docker/dockerfile.service` - Updated to include migrate binary and entrypoint

### Dockerfile Changes

```dockerfile
# In builder stage: Download migrate binary
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /build/migrate

# In runtime stage: Copy everything needed
COPY --from=builder --chown=appuser:appuser /build/migrate ./migrate
COPY --chown=appuser:appuser ./internal/data/migrations ./internal/data/migrations
COPY --chown=appuser:appuser ./infra/docker/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Use entrypoint
ENTRYPOINT ["./entrypoint.sh"]
```

### Entrypoint Script Logic

```bash
#!/bin/sh

1. Wait for database to be ready (pg_isready)
2. Run migrations (migrate -path=/app/internal/data/migrations up)
3. Start application (exec ./app)
```

### How It Works

```
Container starts
    ↓
entrypoint.sh executes
    ↓
Waits for PostgreSQL (pg_isready)
    ↓
Runs migrations (migrate up)
    ↓
Starts your app (./app)
```

### Usage

```bash
# Build new image
make docker-build

# Run with docker-compose
make compose-up

# Migrations run automatically inside API container!
# Check logs:
make compose-logs-api

# You'll see:
# 🚀 Starting doit API...
# ⏳ Waiting for database to be ready...
# ✅ Database is ready!
# 📦 Running database migrations...
# ✅ Migrations completed successfully!
# 🎉 Starting application...
```

### Skip Migrations (Optional)

```yaml
# In docker-compose.yml, add environment variable:
api:
  environment:
    SKIP_MIGRATIONS: "true" # Skip migrations
```

### Pros & Cons

**Pros:**

- ✅ Self-contained (everything in one image)
- ✅ No extra services needed
- ✅ Works in any environment (Docker, K8s, ECS)
- ✅ Production-ready pattern
- ✅ Can skip migrations if needed

**Cons:**

- ⚠️ Slightly larger image (~10MB for migrate binary)
- ⚠️ Migrations run on every container start (but idempotent)
- ⚠️ Need to rebuild image for new migrations

---

## Approach 3: Manual Migrations

**Best for:** Quick testing, development

### Usage

```bash
# Start services (without migrations)
make compose-up

# Run migrations manually
make compose-migrate-up

# Or with docker-compose directly
docker-compose exec api migrate -path=/app/internal/data/migrations -database="postgresql://doit:doit123@postgres:5432/doit?sslmode=disable" up
```

### Pros & Cons

**Pros:**

- ✅ Full control over when migrations run
- ✅ Good for testing migrations
- ✅ Simple to understand

**Cons:**

- ❌ Easy to forget
- ❌ Not automated
- ❌ Team members might miss this step
- ❌ Not suitable for production

---

## Approach 4: Init Container (Kubernetes Style)

**Best for:** Kubernetes deployments

### Implementation

This is what we'll use when we get to Phase 5 (Kubernetes):

```yaml
# Kubernetes deployment
spec:
  initContainers:
    - name: migrations
      image: migrate/migrate:latest
      command:
        - "migrate"
        - "-path=/migrations"
        - "-database=$(DATABASE_URL)"
        - "up"
      volumeMounts:
        - name: migrations
          mountPath: /migrations

  containers:
    - name: api
      image: doit-api:latest
      # ... API config
```

### How It Works

```
Pod starts
    ↓
Init container (migrations) runs
    ↓
Init container completes
    ↓
Main container (API) starts
```

---

## Which Approach to Use?

| Environment       | Recommended Approach           | Why                          |
| ----------------- | ------------------------------ | ---------------------------- |
| **Development**   | Approach 1 (Separate service)  | Easy to see what's happening |
| **Production**    | Approach 2 (Entrypoint script) | Self-contained, reliable     |
| **CI/CD Testing** | Approach 1 or 2                | Both work well               |
| **Kubernetes**    | Approach 4 (Init container)    | Native K8s pattern           |
| **Quick Testing** | Approach 3 (Manual)            | Full control                 |

### Our Current Setup

We've implemented **both Approach 1 and 2** in this project!

**Default (Approach 1):** Separate migration service

- Just run `make compose-up` and everything works
- Easy to see migration logs: `docker logs doit-migrations`

**Alternative (Approach 2):** Built-in migrations

- Rebuild image: `make docker-build`
- Migrations run automatically in API container

---

## Migration Flow Diagram

### Current Setup (Approach 1)

```
Time →

t0:  docker-compose up
     ├─ postgres starts
     ├─ redis starts
     └─ migrations waits...

t1:  postgres healthy ✅
     └─ migrations starts

t2:  migrations runs
     ├─ 000001_create_users_table.up.sql ✅
     ├─ 000002_create_todos_table.up.sql ✅
     └─ 000003_create_refresh_tokens_table.up.sql ✅

t3:  migrations exits (success) ✅
     └─ api starts

t4:  api healthy ✅
     └─ Ready to accept requests!
```

---

## Troubleshooting

### Issue 1: Migrations Service Keeps Restarting

```bash
# Check logs
docker logs doit-migrations

# Common causes:
# 1. Database not ready yet
# 2. Wrong credentials
# 3. Network issues

# Solution: Check DATABASE_URL
docker-compose exec migrations env | grep DATABASE_URL
```

### Issue 2: API Starts But Tables Don't Exist

```bash
# Check if migrations ran
docker logs doit-migrations

# Check migration status in DB
docker-compose exec postgres psql -U doit -d doit -c "SELECT * FROM schema_migrations;"

# Manually run migrations
make compose-migrate-up
```

### Issue 3: Migration Fails with "Dirty Database"

```bash
# Check version
docker-compose exec postgres psql -U doit -d doit -c "SELECT * FROM schema_migrations;"

# Force to specific version (replace N with version number)
docker-compose exec api migrate -path=/app/internal/data/migrations -database="postgresql://doit:doit123@postgres:5432/doit?sslmode=disable" force N

# Then try again
make compose-migrate-up
```

### Issue 4: Entrypoint Script Permission Denied

```bash
# Make sure script is executable
chmod +x infra/docker/entrypoint.sh

# Rebuild image
make docker-build

# Try again
make compose-up-build
```

### Issue 5: Migrations Take Too Long

```bash
# Increase start_period in docker-compose.yml
api:
  healthcheck:
    start_period: 60s  # Give more time for migrations
```

---

## Best Practices

### 1. **Always Make Migrations Idempotent**

```sql
-- ✅ Good: Check before creating
CREATE TABLE IF NOT EXISTS users (...);

-- ❌ Bad: Will fail if table exists
CREATE TABLE users (...);
```

### 2. **Use Transactions**

```sql
-- At top of migration
BEGIN;

-- Your changes
CREATE TABLE ...;
ALTER TABLE ...;

-- At end
COMMIT;
```

### 3. **Test Rollbacks**

```bash
# Test up
make compose-migrate-up

# Test down
make compose-migrate-down

# Test up again
make compose-migrate-up
```

### 4. **Version Control Everything**

```
✅ Commit migrations
✅ Commit entrypoint.sh
✅ Commit docker-compose.yml changes
```

### 5. **Monitor Migration Status**

```bash
# Add this to your health check
/health/ready:
  - Check database connection
  - Check schema_migrations table
  - Return healthy only if migrations applied
```

---

## Testing Your Setup

### Test Approach 1 (Separate Service)

```bash
# Clean start
make compose-down-v

# Start fresh
make compose-up

# Check migration logs
docker logs doit-migrations

# Expected output:
# ✓ 000001_create_users_table.up.sql
# ✓ 000002_create_todos_table.up.sql
# ✓ 000003_create_refresh_tokens_table.up.sql

# Check API started
make compose-logs-api

# Verify tables exist
docker-compose exec postgres psql -U doit -d doit -c "\dt"
```

### Test Approach 2 (Entrypoint)

```bash
# Rebuild with entrypoint
make docker-build

# Start
make compose-up-build

# Watch API logs for migration output
make compose-logs-api

# Look for:
# 🚀 Starting doit API...
# ⏳ Waiting for database to be ready...
# ✅ Database is ready!
# 📦 Running database migrations...
# ✅ Migrations completed successfully!
```

---

## Summary

We've implemented **automated database migrations** in Docker with two approaches:

**Approach 1 (Current Default):** Separate migration service

- ✅ Easy to debug
- ✅ Clear separation of concerns
- ✅ Works immediately with `make compose-up`

**Approach 2 (Alternative):** Built-in entrypoint script

- ✅ Self-contained
- ✅ Production-ready
- ✅ No extra services needed

Both are production-grade solutions. Choose based on your needs!

---

## Related Documentation

- [Docker Compose Implementation](./DOCKER_COMPOSE_IMPLEMENTATION.md)
- [Migration Guide](../../internal/data/docs/MIGRATION_GUIDE.md)
- [Database Setup](../../internal/data/docs/SEEDING.md)

---

**No more manual migration steps!** 🎉

Your database schema is now automatically created when you run `make compose-up`!
