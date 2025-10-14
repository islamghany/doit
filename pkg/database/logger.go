package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/tracelog"
)

// PgxLogger implements tracelog.Logger interface
type PgxLogger struct{}

func NewPgxLogger() *PgxLogger {
	return &PgxLogger{}
}

func (l *PgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	// You can integrate this with your custom logger package
	log.Printf("[PGX %s] %s %v", level, msg, data)
}
