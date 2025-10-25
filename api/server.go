package api

import (
	"doit/api/v1/auth"
	"doit/internal/config"
	"doit/internal/middlewares"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/database"
	"doit/pkg/logger"
	"net/http"
)

func NewServer(logger *logger.Logger, cfg *config.Config, dbPool *database.Pool) http.Handler {

	// Middlewares
	errorMiddleware := middlewares.ErrorMiddleware(logger)

	// Web App
	app := web.NewApp(errorMiddleware)

	// Services
	authService := service.NewAuthService(dbPool, logger)

	// Handlers
	authHandler := auth.NewHandler(logger, authService, cfg)

	// Routes
	auth.RegisterRoutes(app, authHandler)

	app.Handle("GET", "/healthcheck", func(w http.ResponseWriter, r *http.Request) error {
		return web.RespondOK(w, r, map[string]string{"status": "ok"})
	})

	return app
}
