package main

import (
	"context"
	"doit/api"
	"doit/internal/config"
	"doit/internal/web"
	"doit/pkg/logger"
	"fmt"
	"os"
)

// this vars are set by the compiler via ldflags. e.g. go build -ldflags "-X main.build=production -X main.version=v1.0.0"
var build = "development"
var version = "v0.0.1"

func main() {
	// Create a context for the application.
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log := setupLogger(cfg)

	// Start application
	log.Info(ctx, "Starting application",
		"build", build,
		"version", version,
		"environment", cfg.App.Environment,
	)

	fmt.Println("cfg", cfg)
	if err := api.Run(ctx, log, cfg); err != nil {
		log.Error(ctx, "Failed to start the application.", "error", err)
		os.Exit(1)
	}
}

func setupLogger(cfg *config.Config) *logger.Logger {
	// Determine log level
	var minLevel logger.Level
	switch cfg.App.LogLevel {
	case "debug":
		minLevel = logger.LevelDebug
	case "warn":
		minLevel = logger.LevelWarn
	case "error":
		minLevel = logger.LevelError
	default:
		minLevel = logger.LevelInfo
	}

	// TraceID function
	traceIDFunc := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	// Event callbacks
	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			// TODO: Send to error tracking service (e.g., Sentry)
			// For now, just ensure it's logged
		},
	}

	return logger.NewWithEvents(os.Stdout, minLevel, "doit", traceIDFunc, events)
}
