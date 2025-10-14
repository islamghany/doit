package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is an interface that both pgxpool.Pool and pgx.Tx implement
// This allows services to work with both transactions and regular connections
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// TxOptions represents transaction options
type TxOptions struct {
	IsoLevel   pgx.TxIsoLevel
	AccessMode pgx.TxAccessMode
	ReadOnly   bool
}

// DefaultTxOptions returns the default transaction options
func DefaultTxOptions() TxOptions {
	return TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
		ReadOnly:   false,
	}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func WithTransaction(ctx context.Context, pool *pgxpool.Pool, opts TxOptions, fn func(pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   opts.IsoLevel,
		AccessMode: opts.AccessMode,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Execute the function
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// WithSerializableTransaction executes a function in a serializable transaction
// Useful for operations that require the highest isolation level
func WithSerializableTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	opts := TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return WithTransaction(ctx, pool, opts, fn)
}

// WithReadOnlyTransaction executes a function in a read-only transaction
// Useful for complex queries that need consistent snapshot
func WithReadOnlyTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	opts := TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
		ReadOnly:   true,
	}
	return WithTransaction(ctx, pool, opts, fn)
}
