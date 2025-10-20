# Database Seeding Guide

## Quick Start

```bash
# Seed development data
make seed-dev

# Seed test data
make seed-test

# Complete setup (DB + migrations + seeds)
make setup
```

## Overview

The seeding system populates your database with test data for development and testing environments.

### Key Concepts

- **Seeds ≠ Migrations**: Seeds are for data, migrations are for schema
- **Environment-specific**: Different data for dev, test, and staging
- **Idempotent**: Can run multiple times safely
- **SQL-based**: Simple SQL files with INSERT statements

## Directory Structure

```
internal/data/seeder/
├── seeder.go              # Seeding engine
└── seeds/                 # SQL seed files
    ├── 001_dev_users.sql
    └── 002_dev_todos.sql
```

## Creating Seed Files

### Naming Convention

Files are filtered by environment name in the filename:

- `001_dev_users.sql` - Runs with `make seed-dev`
- `001_test_users.sql` - Runs with `make seed-test`
- `001_users.sql` - Runs in all environments

### File Content

Always make seeds idempotent using `ON CONFLICT`:

```sql
-- Good: Idempotent seed
INSERT INTO users (id, email, username, password_hash, email_verified, is_active)
VALUES
  ('550e8400-e29b-41d4-a716-446655440000', 'admin@example.com', 'admin', '$2a$10$...', true, true),
  ('550e8400-e29b-41d4-a716-446655440001', 'user@example.com', 'user', '$2a$10$...', true, true)
ON CONFLICT (email) DO NOTHING;
```

## Example Seed Files

### Development Users

**`internal/data/seeder/seeds/001_dev_users.sql`**

```sql
-- Development test users
-- All passwords are 'password123' (hashed with bcrypt)

INSERT INTO users (id, email, username, password_hash, email_verified, is_active, metadata)
VALUES
  -- Admin user
  (
    '550e8400-e29b-41d4-a716-446655440000',
    'admin@example.com',
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    true,
    true,
    '{"role": "admin", "department": "engineering"}'::jsonb
  ),
  -- Regular user
  (
    '550e8400-e29b-41d4-a716-446655440001',
    'user@example.com',
    'user',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    true,
    true,
    '{"role": "user"}'::jsonb
  ),
  -- Unverified user
  (
    '550e8400-e29b-41d4-a716-446655440002',
    'unverified@example.com',
    'unverified',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    false,
    true,
    '{}'::jsonb
  )
ON CONFLICT (email) DO NOTHING;
```

### Development Todos

**`internal/data/seeder/seeds/002_dev_todos.sql`**

```sql
-- Development test todos

INSERT INTO todos (id, user_id, title, description, status, priority, tags, due_date)
VALUES
  -- Admin's todos
  (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '550e8400-e29b-41d4-a716-446655440000',
    'Welcome to Doit!',
    'This is your first todo. Try marking it as complete.',
    'pending',
    'medium',
    ARRAY['welcome', 'tutorial'],
    NOW() + INTERVAL '7 days'
  ),
  (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12',
    '550e8400-e29b-41d4-a716-446655440000',
    'Review pull request #123',
    'Check the implementation and provide feedback',
    'in_progress',
    'high',
    ARRAY['work', 'code-review'],
    NOW() + INTERVAL '2 days'
  ),
  -- Regular user's todos
  (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13',
    '550e8400-e29b-41d4-a716-446655440001',
    'Buy groceries',
    'Milk, eggs, bread',
    'pending',
    'low',
    ARRAY['personal', 'shopping'],
    NOW() + INTERVAL '1 day'
  )
ON CONFLICT (id) DO NOTHING;
```

## Workflow Examples

### New Developer Setup

```bash
# 1. Clone the repository
git clone <repo-url>
cd doit

# 2. Start database
make dev-db

# 3. Run migrations
make migrate-up

# 4. Seed development data
make seed-dev

# 5. Start the app
make run

# Or use the shortcut:
make setup && make run
```

### Testing Workflow

```bash
# Setup test database
make migrate-up
make seed-test

# Run tests
go test ./...
```

### Resetting Development Data

```bash
# Option 1: Rerun seeds (updates existing)
make seed-dev

# Option 2: Fresh start
make migrate-fresh  # Drops all tables
make seed-dev       # Repopulate
```

## Best Practices

### ✅ DO

- Use `ON CONFLICT DO NOTHING` for idempotency
- Use fixed UUIDs for predictable testing
- Hash passwords with bcrypt
- Order files with numeric prefixes (001, 002, 003)
- Keep seeds environment-specific
- Use meaningful test data

### ❌ DON'T

- Seed production databases
- Use plaintext passwords
- Create dependencies between environments
- Use random UUIDs (makes testing harder)
- Forget foreign key constraints (seed parents first)

## Generating Password Hashes

Use bcrypt to hash passwords for seed files:

```bash
# Generate a bcrypt hash in Go
go run -ldflags="-s -w" <(cat <<'EOF'
package main
import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run hash.go <password>")
        return
    }
    hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
EOF
) "password123"
```

Or create a helper script:

```bash
# scripts/hash_password.sh
#!/bin/bash
password="${1:-password123}"
go run -ldflags="-s -w" <(cat <<'EOF'
package main
import ("fmt"; "os"; "golang.org/x/crypto/bcrypt")
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
EOF
) "$password"
```

## Troubleshooting

### Error: pattern seeds/\*.sql: no matching files found

**Cause**: Seed files are not in the correct location.

**Solution**: Place seed files in `internal/data/seeder/seeds/` (relative to the seeder package).

### Error: violates foreign key constraint

**Cause**: Child records inserted before parent records.

**Solution**: Order your seed files correctly:

```
001_dev_users.sql      # Parent
002_dev_todos.sql      # Child (references users)
```

### Error: duplicate key value

**Cause**: Trying to insert duplicate data without `ON CONFLICT`.

**Solution**: Add `ON CONFLICT DO NOTHING` to your INSERT statements.

### Seeds not running for my environment

**Cause**: Filename doesn't match environment name.

**Solution**: Include environment in filename:

- For dev: `001_dev_users.sql`
- For test: `001_test_users.sql`

## Commands Reference

| Command          | Description                                   |
| ---------------- | --------------------------------------------- |
| `make seed`      | Run all seed files                            |
| `make seed-dev`  | Run only development seeds (`*_dev_*.sql`)    |
| `make seed-test` | Run only test seeds (`*_test_*.sql`)          |
| `make setup`     | Complete setup: database + migrations + seeds |
| `make dev`       | Full workflow: setup + run application        |

## Advanced: Programmatic Seeding

For complex scenarios, extend the seeder with Go functions:

```go
// internal/data/seeder/custom_seeder.go
package seeder

import "context"

func (s *Seeder) SeedLargeDataset(ctx context.Context) error {
    // Generate 1000 users programmatically
    for i := 0; i < 1000; i++ {
        _, err := s.pool.Exec(ctx, `
            INSERT INTO users (email, username, password_hash)
            VALUES ($1, $2, $3)
            ON CONFLICT DO NOTHING
        `, fmt.Sprintf("user%d@example.com", i),
           fmt.Sprintf("user%d", i),
           "$2a$10$...")

        if err != nil {
            return err
        }
    }
    return nil
}
```

Then call it from `cmd/seed/main.go`:

```go
if *large {
    if err := s.SeedLargeDataset(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Related Documentation

- [MIGRATION_GUIDE.md](./MIGRATION_GUIDE.md) - Database migration documentation
- [README_SQLC_PGX.md](./README_SQLC_PGX.md) - Database query generation
- [SETUP.md](./SETUP.md) - Initial project setup
