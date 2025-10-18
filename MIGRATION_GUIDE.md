# Database Migration Guide

## What Happened Today

You encountered a **"Dirty database version"** error. This happens when a migration fails partway through execution, leaving the database in an inconsistent state.

### Root Cause

The `uuid-ossp` PostgreSQL extension wasn't available when the migrations ran, causing them to fail.

### How We Fixed It

1. Manually created the required PostgreSQL extensions (`uuid-ossp` and `pg_trgm`)
2. Dropped the `schema_migrations` table to reset tracking
3. Reran all migrations successfully

---

## Understanding the Dirty Database Error

### What is a "Dirty" Migration?

When golang-migrate runs a migration, it:

1. Marks the version as "dirty" in `schema_migrations` table
2. Executes the SQL statements
3. Marks the version as "clean" when complete

If step 2 fails, the version stays "dirty" and blocks further migrations.

### Common Causes

1. **Syntax errors** in SQL
2. **Missing dependencies** (extensions, functions, etc.)
3. **Constraint violations** (duplicate data, FK violations)
4. **Permission issues**
5. **Connection interruptions**

---

## New Makefile Commands

### Quick Reference

```bash
# Check migration status
make migrate-status

# Fix a dirty migration (after manually fixing the issue)
make migrate-fix version=1

# Reset migration tracking (keeps your data)
make migrate-reset

# Nuclear option: drop everything and start fresh
make migrate-fresh

# Check database connection
make db-check

# Show all tables
make db-tables
```

### Detailed Usage

#### `make migrate-status`

Shows current migration version and lists all available migrations.

```bash
$ make migrate-status
📊 Migration Status:
====================
2

📁 Available migrations:
  - 000001_create_users_table.up.sql
  - 000002_create_todos_table.up.sql
```

#### `make migrate-fix version=N`

Fixes a dirty migration by forcing it to a specific version.

**When to use:**

- When you've manually fixed the database issue
- When a migration was marked dirty but actually completed

```bash
# If migration 2 is dirty but you've fixed the issue
make migrate-fix version=2
```

#### `make migrate-reset`

Drops the `schema_migrations` table, resetting migration tracking. **Your data tables are preserved.**

**When to use:**

- When migrations are in a completely broken state
- When you want to retrack migrations from scratch

```bash
make migrate-reset
# Then run:
make migrate-up
```

#### `make migrate-fresh`

⚠️ **DANGER ZONE** - Drops the entire `public` schema and recreates it, deleting ALL data.

**When to use:**

- Development only
- When you want a completely clean database

```bash
make migrate-fresh  # Will ask for confirmation
```

#### `make db-check`

Verifies database connection and shows PostgreSQL version.

#### `make db-tables`

Lists all tables in your database.

---

## Migration Best Practices

### 1. Test Migrations Locally First

```bash
# Always test in a dev environment
make migrate-up

# If it fails, check what went wrong
make migrate-status

# Fix the issue, then force the version
make migrate-fix version=N
```

### 2. Make Migrations Idempotent

Use `IF EXISTS` and `IF NOT EXISTS`:

```sql
-- ✅ Good
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
DROP TABLE IF EXISTS old_table;

-- ❌ Bad
CREATE TABLE users (...);  -- Fails if table exists
```

### 3. One Logical Change Per Migration

```bash
# ✅ Good
make migrate-create name=add_users_table
make migrate-create name=add_todos_table

# ❌ Bad
make migrate-create name=add_all_tables
```

### 4. Check Dependencies

Before running migrations, ensure:

- Required extensions are installed
- Referenced tables/functions exist
- Sequences are available

```sql
-- At the start of your first migration
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

### 5. Use Transactions (Where Possible)

Most DDL in PostgreSQL is transactional. If one statement fails, the entire migration rolls back.

However, some operations can't be wrapped in transactions:

- `CREATE DATABASE`
- `CREATE INDEX CONCURRENTLY`

### 6. Write Proper Down Migrations

Down migrations must **reverse operations in the correct order**:

```sql
-- ✅ GOOD: Correct order
-- 1. Drop triggers first (they depend on tables)
DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;

-- 2. Drop the table (indexes drop automatically)
DROP TABLE IF EXISTS todos;

-- 3. Drop types/functions last (nothing depends on them)
DROP TYPE IF EXISTS todo_status;
DROP FUNCTION IF EXISTS some_function();

-- ❌ BAD: Wrong order
DROP TABLE todos;              -- Works
DROP INDEX idx_todos_status;   -- ERROR: index doesn't exist (dropped with table!)
DROP TRIGGER update_todos;     -- ERROR: trigger doesn't exist (dropped with table!)
```

**Key points:**

- When you `DROP TABLE`, PostgreSQL **automatically drops**:
  - All indexes on that table
  - All triggers on that table
  - All constraints on that table
- Always use `IF EXISTS` to make down migrations idempotent
- Drop in reverse dependency order: triggers → tables → types → functions
- Be careful dropping shared resources (like functions used by multiple tables)

### 7. Keep Backups

```bash
# Before risky migrations
docker exec doit_db pg_dump -U islamghany doit > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore if needed
docker exec -i doit_db psql -U islamghany doit < backup_20231018_120000.sql
```

---

## Common Scenarios & Solutions

### Scenario 1: Migration Fails Partway

```bash
# Check status
make migrate-status
# Output: 2 (dirty)

# Option A: Fix the issue and force version
make migrate-fix version=2

# Option B: Reset and rerun
make migrate-reset
make migrate-up
```

### Scenario 2: Need to Rollback

```bash
# Rollback one version
make migrate-down

# Check new version
make migrate-version
```

### Scenario 3: Migration File Has a Typo

```bash
# 1. Fix the .sql file
# 2. Rollback to before that migration
make migrate-down

# 3. Rerun with the fix
make migrate-up
```

### Scenario 4: Down Migration Fails with "does not exist" Errors

```bash
# Error: index "idx_todos_status" does not exist
# Error: trigger "update_todos" does not exist
```

**Cause:** Trying to drop indexes/triggers after dropping the table (they're auto-dropped with the table).

**Solution:**

```bash
# 1. Fix the down migration file (use IF EXISTS, correct order)
# 2. Fix the dirty state
make migrate-fix version=1

# 3. Test the fixed migration
make migrate-down
make migrate-up
```

### Scenario 5: Can't Force to Version 0

golang-migrate has a bug where forcing to version 0 sometimes fails.

**Solution:**

```bash
# Drop the tracking table instead
make migrate-reset
```

### Scenario 6: Database is Completely Broken

```bash
# Nuclear option (development only!)
make migrate-fresh
```

---

## Troubleshooting Checklist

When you get a dirty migration error:

- [ ] Check `make migrate-status` to see which version is dirty
- [ ] Look at the error message to understand what failed
- [ ] Check if the migration partially executed (`make db-tables`)
- [ ] Manually fix the database issue if needed
- [ ] Use `make migrate-fix version=N` to mark as clean
- [ ] Verify with `make migrate-status`
- [ ] Continue with `make migrate-up`

---

## Environment-Specific Strategies

### Development

- Use `make migrate-fresh` freely
- Test migrations multiple times
- Keep test data in seed files (coming soon!)

### Staging

- Always test migrations here before production
- Keep backups before migrations
- Use `make migrate-fix` carefully

### Production

- **NEVER** use `migrate-fresh` or `migrate-reset`
- Always have rollback plan
- Test in staging first
- Run migrations during low-traffic windows
- Have database backups
- Consider blue-green deployments for zero-downtime

---

## Next Steps

Consider adding:

1. **Database seeding** for test data (see our earlier conversation)
2. **Migration versioning** in CI/CD
3. **Automatic backup** before migrations
4. **Migration smoke tests**

---

## Quick Command Reference

| Command                 | What It Does                  | Safe?                     |
| ----------------------- | ----------------------------- | ------------------------- |
| `migrate-up`            | Run all pending migrations    | ✅ Yes                    |
| `migrate-down`          | Rollback one migration        | ⚠️ Use with care          |
| `migrate-status`        | Show current version          | ✅ Yes                    |
| `migrate-version`       | Show current version (simple) | ✅ Yes                    |
| `migrate-fix version=N` | Force version N               | ⚠️ Use after fixing issue |
| `migrate-reset`         | Drop tracking table           | ⚠️ Dev only               |
| `migrate-fresh`         | Drop all tables               | ❌ Dev only!              |
| `db-check`              | Check connection              | ✅ Yes                    |
| `db-tables`             | List all tables               | ✅ Yes                    |

---

## Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Transaction DDL](https://wiki.postgresql.org/wiki/Transactional_DDL_in_PostgreSQL)
- [Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
