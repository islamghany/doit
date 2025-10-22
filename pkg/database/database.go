package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	DisableTLS      bool
	LogLevel        string
}

// Pool wraps pgxpool.Pool with additional functionality
type Pool struct {
	*pgxpool.Pool
}

// PoolConn wraps a connection from the pool
type PoolConn struct {
	Conn *pgxpool.Conn
}

// New creates a new database connection pool with optimized settings
func New(ctx context.Context, cfg Config) (*Pool, error) {
	poolConfig, err := buildPoolConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// buildPoolConfig creates an optimized pgxpool configuration
func buildPoolConfig(cfg Config) (*pgxpool.Config, error) {
	dsn := buildDSN(cfg)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Performance optimizations
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	// Logging (optional, based on log level)
	if cfg.LogLevel == "debug" {
		poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   NewPgxLogger(),
			LogLevel: tracelog.LogLevelDebug,
		}
	}

	return poolConfig, nil
}

// BuildDSN constructs the PostgreSQL connection string
// Exported for testing purposes
func BuildDSN(cfg Config) string {
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		sslMode,
		cfg.MaxConns,
	)
}

// buildDSN is kept for backward compatibility (private wrapper)
func buildDSN(cfg Config) string {
	return BuildDSN(cfg)
}

// StatusCheck returns pool statistics for health checks
func (p *Pool) StatusCheck(ctx context.Context) error {
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// Stats returns detailed pool statistics
func (p *Pool) Stats() map[string]interface{} {
	stat := p.Stat()
	return map[string]interface{}{
		"acquired_conns":         stat.AcquiredConns(),
		"canceled_acquire_count": stat.CanceledAcquireCount(),
		"constructing_conns":     stat.ConstructingConns(),
		"empty_acquire_count":    stat.EmptyAcquireCount(),
		"idle_conns":             stat.IdleConns(),
		"max_conns":              stat.MaxConns(),
		"total_conns":            stat.TotalConns(),
		"new_conns_count":        stat.NewConnsCount(),
		"max_lifetime_destroy":   stat.MaxLifetimeDestroyCount(),
		"max_idle_destroy_count": stat.MaxIdleDestroyCount(),
	}
}
