package api

import (
	"net/http"

	"doit/api/v1/auth"
	"doit/internal/config"
	"doit/internal/middlewares"
	"doit/internal/service"
	"doit/internal/token"
	"doit/internal/web"
	"doit/pkg/database"
	"doit/pkg/logger"
)

func NewServer(logger *logger.Logger, cfg *config.Config, dbPool *database.Pool) (http.Handler, error) {
	// Helpers
	tokenMaker, err := token.NewJWTToken(cfg.JWT.Secret)
	if err != nil {
		return nil, err
	}

	// Middlewares
	errorMiddleware := middlewares.ErrorMiddleware(logger)

	// Web App
	app := web.NewApp(errorMiddleware)

	// Services
	userService := service.NewUserService(dbPool)
	tokenService := service.NewTokenService(dbPool, tokenMaker,
		cfg.JWT.AccessTokenExp,
		cfg.JWT.RefreshTokenExp)

	// Handlers
	authHandler := auth.NewHandler(logger, userService, tokenService, cfg)

	// Routes
	auth.RegisterRoutes(app, authHandler)

	app.Handle("GET", "/healthcheck", func(w http.ResponseWriter, r *http.Request) error {
		return web.RespondOK(w, r, map[string]string{"status": "ok"})
	})

	return app, nil
}
