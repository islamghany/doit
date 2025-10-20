# Database Seeding Setup - Complete ✅

## What Was Added

### 1. Seeder Package Structure

```
internal/data/seeder/
├── seeder.go                    # NEW - Seeding engine
└── seeds/                       # NEW - Seed files directory
    ├── 001_dev_users.sql       # NEW - Development users
    └── 002_dev_todos.sql       # NEW - Development todos
```

### 2. Seed Command

```
cmd/seed/
└── main.go                      # NEW - CLI for running seeds
```

### 3. Makefile Commands

**Added commands:**

```makefile
make seed           # Run all seed files
make seed-dev       # Run development seeds only
make seed-test      # Run test seeds only
make setup          # Complete setup (DB + migrations + seeds)
make dev            # Full workflow (setup + run)
```

### 4. Documentation

**New files:**

- ✅ `SEEDING.md` - Complete seeding guide with examples
- ✅ `SEEDING_SETUP_SUMMARY.md` - This file

**Updated files:**

- ✅ `MIGRATION_GUIDE.md` - Added comprehensive seeding section
- ✅ `QUICK_REFERENCE.md` - Added seeding quick reference
- ✅ `Makefile` - Added seeding commands and help menu

---

## Quick Start

### Run Seeds Now

```bash
# Make sure your database is running and migrated
make migrate-status

# Seed development data
make seed-dev

# Verify data was inserted
make db-tables
```

### Create Your Own Seed File

1. **Create a new SQL file** in `internal/data/seeder/seeds/`:

```bash
touch internal/data/seeder/seeds/003_dev_custom.sql
```

2. **Add idempotent INSERT statements**:

```sql
-- internal/data/seeder/seeds/003_dev_custom.sql
INSERT INTO users (id, email, username, password_hash)
VALUES
  ('550e8400-e29b-41d4-a716-446655440003', 'custom@example.com', 'custom', '$2a$10$...')
ON CONFLICT (email) DO NOTHING;
```

3. **Run the seeder**:

```bash
make seed-dev
```

---

## File Contents Reference

### Seeder Engine

**`internal/data/seeder/seeder.go`**

```go
package seeder

import (
    "context"
    "embed"
    "fmt"
    "path/filepath"
    "strings"
    "github.com/jackc/pgx/v5/pgxpool"
)

//go:embed seeds/*.sql
var seedsFS embed.FS

type Seeder struct {
    pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Seeder {
    return &Seeder{pool: pool}
}

func (s *Seeder) Run(ctx context.Context, env string) error {
    files, err := seedsFS.ReadDir("seeds")
    if err != nil {
        return fmt.Errorf("failed to read seeds directory: %w", err)
    }

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        if env != "" && !strings.Contains(file.Name(), env) {
            continue
        }

        data, err := seedsFS.ReadFile(filepath.Join("seeds", file.Name()))
        if err != nil {
            return fmt.Errorf("failed to read seed file: %w", err)
        }

        if _, err := s.pool.Exec(ctx, string(data)); err != nil {
            return fmt.Errorf("seed %s failed: %w", file.Name(), err)
        }
    }
    return nil
}
```

### Seed CLI Command

**`cmd/seed/main.go`**

```go
package main

import (
    "context"
    "flag"
    "log"
    "os"

    "doit/internal/config"
    "doit/internal/data/seeder"
    "doit/pkg/database"
)

func main() {
    env := flag.String("env", "dev", "Environment (dev, test, prod)")
    flag.Parse()

    ctx := context.Background()

    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create database pool
    pool, err := database.NewPool(ctx, cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer pool.Close()

    // Run seeder
    s := seeder.New(pool)
    log.Printf("Running seeds for environment: %s", *env)

    if err := s.Run(ctx, *env); err != nil {
        log.Fatalf("Seeding failed: %v", err)
    }

    log.Println("✅ Seeding completed successfully")
}
```

### Example Seed Files

**`internal/data/seeder/seeds/001_dev_users.sql`**

```sql
-- Development test users
-- Password for all users: 'password123'

INSERT INTO users (id, email, username, password_hash, email_verified, is_active, metadata)
VALUES
  (
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@example.com',
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    true,
    true,
    '{"role": "admin"}'::jsonb
  ),
  (
    '550e8400-e29b-41d4-a716-446655440001',
    'user@example.com',
    'user',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    true,
    true,
    '{"role": "user"}'::jsonb
  )
ON CONFLICT (email) DO NOTHING;
```

**`internal/data/seeder/seeds/002_dev_todos.sql`**

```sql
-- Development test todos

INSERT INTO todos (id, user_id, title, description, status, priority, tags)
VALUES
  (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '550e8400-e29b-41d4-a716-446655440000',
    'Welcome to Doit!',
    'This is your first todo',
    'pending',
    'medium',
    ARRAY['welcome']
  )
ON CONFLICT (id) DO NOTHING;
```

---

## Development Workflow

### Complete Setup (First Time)

```bash
# 1. Start database
make dev-db

# 2. Run migrations
make migrate-up

# 3. Seed data
make seed-dev

# 4. Run application
make run

# Or use the shortcut:
make setup && make run
```

### Daily Development

```bash
# Quick start (if DB already exists)
make dev  # Does: setup + run

# Reseed data (during development)
make seed-dev

# Fresh start (reset everything)
make migrate-fresh
make seed-dev
```

### Testing Workflow

```bash
# Setup test database
make migrate-up
make seed-test

# Run tests
go test ./...
```

---

## Best Practices Reminder

### ✅ DO

- Use `ON CONFLICT DO NOTHING` for idempotency
- Use fixed UUIDs for test data
- Hash passwords with bcrypt
- Name files with numeric prefixes: `001_`, `002_`, etc.
- Include environment in filename: `*_dev_*.sql`, `*_test_*.sql`
- Seed in dependency order (parents before children)

### ❌ DON'T

- Seed production databases
- Use plaintext passwords
- Use random UUIDs (makes testing harder)
- Forget foreign key constraints
- Skip `ON CONFLICT` clauses

---

## Extending the Seeder

### Add Environment-Specific Seeds

Create files with environment prefixes:

```bash
# Test environment (minimal data)
internal/data/seeder/seeds/001_test_users.sql

# Staging environment (realistic data)
internal/data/seeder/seeds/001_staging_users.sql
```

Run with:

```bash
make seed-test      # Runs *_test_* files
make seed-staging   # Need to add to Makefile
```

### Add Programmatic Seeding

For complex scenarios, extend the seeder:

```go
// internal/data/seeder/advanced_seeder.go
package seeder

func (s *Seeder) SeedRandomUsers(ctx context.Context, count int) error {
    for i := 0; i < count; i++ {
        // Generate random test data
        // Insert with proper error handling
    }
    return nil
}
```

---

## Troubleshooting

### "no matching files found"

**Cause**: Seeds are in wrong location.

**Fix**: Move to `internal/data/seeder/seeds/`

```bash
mv internal/data/seeds/* internal/data/seeder/seeds/
```

### "duplicate key value"

**Cause**: Missing `ON CONFLICT` clause.

**Fix**: Add to INSERT statements:

```sql
INSERT INTO users (...)
VALUES (...)
ON CONFLICT (email) DO NOTHING;  -- Add this!
```

### "violates foreign key constraint"

**Cause**: Seeding child before parent.

**Fix**: Rename files to correct order:

```bash
001_dev_users.sql     # Parent first
002_dev_todos.sql     # Child second
```

---

## Next Steps

1. **Test the seeding**:

   ```bash
   make seed-dev
   ```

2. **Verify data**:

   ```bash
   make db-tables
   ```

3. **Add your own seeds** for your specific use case

4. **Update CI/CD** to include seeding:
   ```yaml
   - run: make migrate-up
   - run: make seed-test
   - run: go test ./...
   ```

---

## Documentation Links

- 📚 [SEEDING.md](./SEEDING.md) - Complete seeding guide
- 📚 [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) - Migration best practices
- 📚 [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Quick reference guide

---

**Happy seeding! 🌱**
