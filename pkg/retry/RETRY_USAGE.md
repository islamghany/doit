# Retry Logic Usage Guide

This package provides a **generic retry mechanism with exponential backoff** that works with any connection type or operation.

## Table of Contents

- [Quick Start](#quick-start)
- [Core Functions](#core-functions)
- [Configuration](#configuration)
- [Examples](#examples)
- [Best Practices](#best-practices)

---

## Quick Start

The simplest way to use retry logic:

```go
import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "doit/pkg/retry"
)

func main() {
    ctx := context.Background()

    // Use default retry configuration (5 attempts, exponential backoff)
    pool, err := retry.ConnectWithRetry(ctx, retry.DefaultRetryConfig(),
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, "postgres://...")
            if err != nil {
                return nil, err
            }

            // Verify connection
            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )

    if err != nil {
        log.Fatal(err)
    }

    defer pool.Close()
}
```

---

## Core Functions

### 1. `ConnectWithRetry[T any]`

**Generic function that retries any connection creation.**

```go
func ConnectWithRetry[T any](
    ctx context.Context,
    cfg RetryConfig,
    connectFn func(context.Context) (T, error),
) (T, error)
```

**Usage:**

```go
pool, err := retry.ConnectWithRetry(ctx, cfg,
    func(ctx context.Context) (*pgxpool.Pool, error) {
        pool, err := pgxpool.New(ctx, connString)
        if err != nil {
            return nil, err
        }

        // Validate connection
        if err := pool.Ping(ctx); err != nil {
            pool.Close()
            return nil, err
        }

        return pool, nil
    },
)
```

---

### 2. `ExecuteWithRetry`

**Retries operations that don't return a value (only error).**

```go
func ExecuteWithRetry(
    ctx context.Context,
    cfg RetryConfig,
    operation func(context.Context) error,
) error
```

**Usage:**

```go
err := retry.ExecuteWithRetry(ctx, cfg,
    func(ctx context.Context) error {
        return runMigrations(db)
    },
)
```

---

## Configuration

### RetryConfig Structure

```go
type RetryConfig struct {
    MaxAttempts  int           // Max retry attempts (0 = no retries)
    InitialDelay time.Duration // Delay before first retry
    MaxDelay     time.Duration // Maximum delay between retries
    Multiplier   float64       // Backoff multiplier (usually 2.0)
    Jitter       bool          // Add randomness (±20%) to prevent thundering herd
    OnRetry      func(attempt int, err error, delay time.Duration) // Callback
}
```

### Default Configuration

```go
cfg := retry.DefaultRetryConfig()
// Returns:
// - MaxAttempts:  5
// - InitialDelay: 1 second
// - MaxDelay:     30 seconds
// - Multiplier:   2.0
// - Jitter:       true
// - OnRetry:      logs retry attempts
```

### Custom Configuration

```go
cfg := retry.RetryConfig{
    MaxAttempts:  10,
    InitialDelay: 500 * time.Millisecond,
    MaxDelay:     1 * time.Minute,
    Multiplier:   1.5,
    Jitter:       true,
    OnRetry: func(attempt int, err error, delay time.Duration) {
        log.Printf("Retry %d: %v (waiting %v)", attempt, err, delay)
    },
}
```

---

## Examples

### Example 1: PostgreSQL Connection

```go
import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "doit/pkg/retry"
)

func NewDatabasePool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    cfg := retry.DefaultRetryConfig()

    return retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, connString)
            if err != nil {
                return nil, err
            }

            // Validate the connection
            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )
}
```

### Example 2: Redis Connection

```go
import (
    "context"
    "github.com/redis/go-redis/v9"
    "doit/pkg/retry"
)

func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
    cfg := retry.RetryConfig{
        MaxAttempts:  3,
        InitialDelay: 1 * time.Second,
        MaxDelay:     10 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
    }

    return retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*redis.Client, error) {
            client := redis.NewClient(&redis.Options{
                Addr: addr,
            })

            // Validate the connection
            if err := client.Ping(ctx).Err(); err != nil {
                client.Close()
                return nil, err
            }

            return client, nil
        },
    )
}
```

### Example 3: HTTP Client with Health Check

```go
import (
    "context"
    "fmt"
    "net/http"
    "doit/pkg/retry"
)

func NewHTTPClient(ctx context.Context, baseURL string) (*http.Client, error) {
    cfg := retry.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: 2 * time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
    }

    return retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*http.Client, error) {
            client := &http.Client{Timeout: 10 * time.Second}

            // Validate with health check
            resp, err := client.Get(baseURL + "/health")
            if err != nil {
                return nil, err
            }
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                return nil, fmt.Errorf("health check failed: %d", resp.StatusCode)
            }

            return client, nil
        },
    )
}
```

### Example 4: With Timeout

```go
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "doit/pkg/retry"
)

func ConnectWithTimeout(connString string) (*pgxpool.Pool, error) {
    // Total time limit: 2 minutes for all retry attempts
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    cfg := retry.DefaultRetryConfig()

    return retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, connString)
            if err != nil {
                return nil, err
            }

            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )
}
```

### Example 5: Run Migrations with Retry

```go
import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "doit/pkg/retry"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
    cfg := retry.RetryConfig{
        MaxAttempts:  3,
        InitialDelay: 1 * time.Second,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
        Jitter:       false,
    }

    return retry.ExecuteWithRetry(ctx, cfg, func(ctx context.Context) error {
        // Your migration logic
        _, err := db.Exec(ctx, "CREATE TABLE IF NOT EXISTS users (...)")
        return err
    })
}
```

### Example 6: Custom Retry Callback

```go
import (
    "context"
    "log"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "doit/pkg/retry"
)

func ConnectWithCustomLogging(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    cfg := retry.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: 1 * time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
        OnRetry: func(attempt int, err error, delay time.Duration) {
            log.Printf("[WARNING] Database connection attempt %d failed: %v", attempt, err)
            log.Printf("[INFO] Waiting %v before retry...", delay)
        },
    }

    return retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, connString)
            if err != nil {
                return nil, err
            }

            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )
}
```

### Example 7: Multiple Services Startup

```go
import (
    "context"
    "fmt"
    "os"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "doit/pkg/retry"
)

func InitializeServices(ctx context.Context) error {
    cfg := retry.DefaultRetryConfig()

    // Connect to database
    db, err := retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
            if err != nil {
                return nil, err
            }

            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )
    if err != nil {
        return fmt.Errorf("database connection failed: %w", err)
    }

    // Connect to Redis
    redisClient, err := retry.ConnectWithRetry(ctx, cfg,
        func(ctx context.Context) (*redis.Client, error) {
            client := redis.NewClient(&redis.Options{
                Addr: os.Getenv("REDIS_URL"),
            })

            // Validate
            if err := client.Ping(ctx).Err(); err != nil {
                client.Close()
                return nil, err
            }

            return client, nil
        },
    )
    if err != nil {
        return fmt.Errorf("redis connection failed: %w", err)
    }

    fmt.Println("✅ All services connected successfully")
    fmt.Printf("Database: %v\n", db.Stat())
    fmt.Printf("Redis: %v\n", redisClient.Options().Addr)

    return nil
}
```

---

## Best Practices

### 1. **Always Use Context**

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

pool, err := retry.ConnectWithRetry(ctx, cfg, connectFn)
```

### 2. **Set Reasonable Timeouts**

```go
cfg := retry.RetryConfig{
    MaxAttempts:  5,              // Total attempts: 1 + 5 retries = 6
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
}
// Max time: ~1 + 2 + 4 + 8 + 16 + 30 = ~61 seconds
```

### 3. **Enable Jitter in Production**

Prevents thundering herd when multiple instances restart simultaneously:

```go
cfg.Jitter = true  // Adds ±20% randomness
```

### 4. **Log Retry Attempts**

```go
cfg.OnRetry = func(attempt int, err error, delay time.Duration) {
    log.Printf("[RETRY] Attempt %d failed: %v. Retrying in %v", attempt, err, delay)
}
```

### 5. **Clean Up on Failure**

```go
retry.ConnectWithRetry(ctx, cfg,
    func(ctx context.Context) (*pgxpool.Pool, error) {
        pool, err := pgxpool.New(ctx, connString)
        if err != nil {
            return nil, err
        }

        if err := pool.Ping(ctx); err != nil {
            pool.Close() // ← Important! Clean up failed connections
            return nil, err
        }

        return pool, nil
    },
)
```

### 6. **Handle Validation Inside Connect Function**

```go
retry.ConnectWithRetry(ctx, cfg,
    func(ctx context.Context) (*pgxpool.Pool, error) {
        pool, err := pgxpool.New(ctx, connString)
        if err != nil {
            return nil, err
        }

        // Validate immediately
        if err := pool.Ping(ctx); err != nil {
            pool.Close()
            return nil, err // Will trigger retry
        }

        return pool, nil
    },
)
```

### 7. **Use Appropriate Multipliers**

- **Aggressive retry** (quick recovery): `Multiplier = 1.5`
- **Standard** (balanced): `Multiplier = 2.0`
- **Conservative** (avoid overwhelming): `Multiplier = 3.0`

---

## Exponential Backoff Calculation

With `InitialDelay = 1s`, `Multiplier = 2.0`, `MaxDelay = 30s`:

| Attempt | Delay (no jitter) | Delay (with jitter ±20%) |
| ------- | ----------------- | ------------------------ |
| 1       | 1s                | 0.8s - 1.2s              |
| 2       | 2s                | 1.6s - 2.4s              |
| 3       | 4s                | 3.2s - 4.8s              |
| 4       | 8s                | 6.4s - 9.6s              |
| 5       | 16s               | 12.8s - 19.2s            |
| 6       | 30s (capped)      | 24s - 30s                |

---

## Error Handling

The retry mechanism will automatically retry on any error returned from your connect function. You can control what gets retried by:

1. **Returning the error** - Will trigger a retry (up to MaxAttempts)
2. **Returning a wrapped error** - Will still trigger a retry
3. **Context cancellation** - Will stop retries immediately

```go
retry.ConnectWithRetry(ctx, cfg,
    func(ctx context.Context) (*pgxpool.Pool, error) {
        pool, err := pgxpool.New(ctx, connString)
        if err != nil {
            // This error will trigger a retry
            return nil, fmt.Errorf("connection failed: %w", err)
        }

        if err := pool.Ping(ctx); err != nil {
            pool.Close()
            // This error will also trigger a retry
            return nil, fmt.Errorf("ping failed: %w", err)
        }

        return pool, nil // Success - no retry
    },
)
```

### Common Retryable Scenarios

- Network timeouts
- Connection refused
- Temporary database unavailability
- DNS resolution failures
- Connection pool exhaustion

### Non-Retryable Scenarios

You can choose not to retry certain errors by handling them inside your function:

```go
retry.ConnectWithRetry(ctx, cfg,
    func(ctx context.Context) (*pgxpool.Pool, error) {
        pool, err := pgxpool.New(ctx, connString)
        if err != nil {
            // Check if it's an authentication error
            if strings.Contains(err.Error(), "authentication failed") {
                // Don't retry auth errors - fail immediately
                return nil, fmt.Errorf("authentication failed (non-retryable): %w", err)
            }
            return nil, err // Retry other errors
        }

        return pool, nil
    },
)
```

---

## Testing

```go
import (
    "context"
    "testing"
    "time"
    "doit/pkg/retry"
)

func TestDatabaseConnection(t *testing.T) {
    cfg := retry.RetryConfig{
        MaxAttempts:  2,
        InitialDelay: 10 * time.Millisecond,
        MaxDelay:     50 * time.Millisecond,
        Multiplier:   2.0,
        Jitter:       false, // Disable for predictable tests
        OnRetry:      nil,   // Disable logging in tests
    }

    ctx := context.Background()
    pool, err := retry.ConnectWithRetry(ctx, cfg, connectFn)

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if pool == nil {
        t.Fatal("expected pool, got nil")
    }
}
```

---

## Integration with Your Application

Update your `cmd/doit/main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "doit/internal/config"
    "doit/pkg/database"
    "doit/pkg/retry"

    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    ctx := context.Background()

    // Load config
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Setup retry configuration
    retryCfg := retry.RetryConfig{
        MaxAttempts:  5,
        InitialDelay: 2 * time.Second,
        MaxDelay:     30 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
        OnRetry: func(attempt int, err error, delay time.Duration) {
            log.Printf("⚠️  Database connection attempt %d failed: %v. Retrying in %v...",
                attempt, err, delay)
        },
    }

    // Connect to database with retry
    pool, err := retry.ConnectWithRetry(ctx, retryCfg,
        func(ctx context.Context) (*database.Pool, error) {
            dbCfg := database.Config{
                Host:            cfg.Database.Host,
                Port:            cfg.Database.Port,
                Database:        cfg.Database.Name,
                User:            cfg.Database.User,
                Password:        cfg.Database.Password,
                MaxConns:        cfg.Database.MaxOpenConns,
                MinConns:        cfg.Database.MaxIdleConns,
                MaxConnLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
                MaxConnIdleTime: 5 * time.Minute,
                DisableTLS:      cfg.Database.DisableTLS,
                LogLevel:        cfg.App.LogLevel,
            }

            return database.New(ctx, dbCfg)
        },
    )
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer pool.Close()

    log.Println("✅ Connected to database successfully")

    // Your application logic here...
}
```

### Alternative: Simpler Integration

If you want a simpler approach with direct pgxpool:

```go
func main() {
    ctx := context.Background()

    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Build connection string
    connString := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Name,
    )

    // Connect with retry
    pool, err := retry.ConnectWithRetry(ctx, retry.DefaultRetryConfig(),
        func(ctx context.Context) (*pgxpool.Pool, error) {
            pool, err := pgxpool.New(ctx, connString)
            if err != nil {
                return nil, err
            }

            if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, err
            }

            return pool, nil
        },
    )
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer pool.Close()

    log.Println("✅ Database connected successfully")
}
```
