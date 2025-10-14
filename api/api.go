package api

import (
	"context"
	"doit/internal/config"
	"doit/pkg/logger"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, logger *logger.Logger, cfg *config.Config) error {

	// Starting the HTTP server with graceful shutdown
	srv := NewServer(logger, cfg)

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
