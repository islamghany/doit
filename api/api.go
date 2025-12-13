package api

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"doit/internal/cache"
	"doit/internal/config"
	"doit/internal/debug"
	"doit/pkg/database"
	"doit/pkg/logger"
	"doit/pkg/retry"
)

func Run(ctx context.Context, logger *logger.Logger, cfg *config.Config) error {
	// Initialize database with retry logic
	dbPool, err := retry.ConnectWithRetry(ctx, retry.DefaultRetryConfig(), func(ctx context.Context) (*database.Pool, error) {
		return database.New(ctx, database.Config{
			Host:            cfg.Database.Host,
			Port:            cfg.Database.Port,
			Database:        cfg.Database.Name,
			User:            cfg.Database.User,
			Password:        cfg.Database.Password,
			MaxConns:        cfg.Database.MaxOpenConns,
			MinConns:        5,
			MaxConnLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
			MaxConnIdleTime: 30 * time.Minute,
			DisableTLS:      cfg.Database.DisableTLS,
			LogLevel:        cfg.App.LogLevel,
		})
	})
	if err != nil {
		logger.Error(ctx, "Failed to connect to database", "error", err)
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info(ctx, "Database connection established",
		"host", cfg.Database.Host,
		"database", cfg.Database.Name,
	)
	defer dbPool.Close()

	// Initialize cache with retry logic

	opts := &cache.RedisOptions{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.Database,
		PoolSize: 10,
	}
	cache, err := cache.NewRedisCache(opts)
	if err != nil {
		logger.Error(ctx, "Failed to connect to cache", "error", err)
		return fmt.Errorf("failed to connect to cache: %w", err)
	}
	defer cache.Close()

	// Starting the HTTP server with graceful shutdown
	srv, err := NewServer(logger, cfg, dbPool, cache)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	mainServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      srv,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Create debug server (bind to localhost for security)
	expvar.NewString("version").Set(cfg.App.Version)
	expvar.NewString("environment").Set(string(cfg.App.Environment))

	debugServer := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", cfg.Server.DebugPort),
		Handler: debug.Mux(),
	}
	// Channel to collect server errors
	serverErrors := make(chan error, 2)

	go func() {
		logger.Info(ctx, "Starting the HTTP server", "address", mainServer.Addr)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("main server error: %w", err)
		}
	}()

	// Start debug server
	go func() {
		logger.Info(ctx, "Starting the debug server", "address", debugServer.Addr)
		if err := debugServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("debug server error: %w", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		// Server failed to start
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("error starting the server: %w", err)
		}
		return nil
	case sig := <-shutdownChan:
		{
			logger.Info(ctx, "Shutting down the server...", "signal", sig.String())

			// create a timeout context that carries the deadline
			// 25 is reasonable time to wait for the server to shutdown
			// since kubernetes will wait for 30 seconds before killing the pod
			// and we need to give the server enough time to shutdown
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()

			// Shutdown both servers concurrently
			errChan := make(chan error, 2)

			go func() {
				errChan <- mainServer.Shutdown(shutdownCtx)
			}()

			go func() {
				errChan <- debugServer.Shutdown(shutdownCtx)
			}()

			// Wait for both to complete
			var shutdownErr error
			for i := 0; i < 2; i++ {
				if err := <-errChan; err != nil {
					shutdownErr = err
					logger.Error(ctx, "Error during shutdown", "error", err)
				}
			}

			if shutdownErr != nil {
				return fmt.Errorf("error shutting down servers: %w", shutdownErr)
			}

			logger.Info(ctx, "All servers stopped gracefully")
			return nil
		}
	}
}
