package api

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      srv,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	shutdownError := make(chan error)

	go func() {
		logger.Info(ctx, "Starting the HTTP server on", "address", server.Addr)
		shutdownError <- server.ListenAndServe()
	}()

	// Starting the debug server
	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		defer w.Done()

		expvar.NewString("version").Set(cfg.App.Version)
		expvar.NewString("environment").Set(string(cfg.App.Environment))

		logger.Info(ctx, "Starting the debug server on", "address", fmt.Sprintf(":%d", cfg.Server.DebugPort))

		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.DebugPort), debug.Mux()); err != nil {
			logger.Error(ctx, "shutdown", "status", "debug v1 router cloased", "host", cfg.Server.DebugPort)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-shutdownError:
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
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				server.Close()
				return fmt.Errorf("error shutting down the server gracefully: %w", err)
			}
			logger.Info(ctx, "Server stopped gracefully")
			return nil
		}
	}
}
