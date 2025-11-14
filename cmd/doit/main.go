// Package main DoIt API
//
// # RESTful API for managing todos with JWT-based authentication
//
// The API implements modern security best practices including:
// - JWT access tokens (short-lived) and refresh tokens (long-lived)
// - Refresh token rotation for enhanced security
// - Password hashing with bcrypt
// - Role-based access control (RBAC)
// - Rate limiting on authentication endpoints
// - CORS and security headers
//
// Authentication Flow:
// 1. Register or Login to receive access token and refresh token
// 2. Use access token in Authorization header: "Bearer {token}"
// 3. When access token expires, use refresh token to get new tokens
// 4. Logout to revoke refresh token
//
// Terms Of Service:
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//	Contact: DoIt Support<support@doit.com>
//	License: MIT http://opensource.org/licenses/MIT
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	SecurityDefinitions:
//	bearer:
//	  type: apiKey
//	  name: Authorization
//	  in: header
//	  description: "Enter your JWT token in the format: Bearer {token}"
//
// swagger:meta
package main

import (
	"context"
	"fmt"
	"os"

	"doit/api"
	"doit/internal/config"
	"doit/internal/web"
	"doit/pkg/logger"
)

// // this vars are set by the compiler via ldflags. e.g. go build -ldflags "-X main.build=production -X main.version=v1.0.0"
// var (
// 	build   = "development"
// 	version = "v0.0.1"
// )

func main() {
	// Create a context for the application.
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	cfg.DevPrint()

	// Setup logger
	log := setupLogger(cfg)

	// Start application
	log.Info(ctx, "Starting application",
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
	)

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
