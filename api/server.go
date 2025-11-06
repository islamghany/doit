package api

import (
	"errors"
	"net/http"

	"doit/api/v1/auth"
	"doit/internal/cache"
	"doit/internal/config"
	"doit/internal/limiter"
	"doit/internal/middlewares"
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/token"
	"doit/internal/web"
	"doit/pkg/database"
	"doit/pkg/logger"
)

func NewServer(logger *logger.Logger, cfg *config.Config, dbPool *database.Pool, cache cache.Cache) (http.Handler, error) {
	// Helpers
	tokenMaker, err := token.NewJWTToken(cfg.JWT.Secret)
	if err != nil {
		return nil, err
	}
	// Rate Limiter
	rateLimiter := limiter.NewRateLimiter(cache, logger)

	// Services
	userService := service.NewUserService(dbPool)
	tokenService := service.NewTokenService(dbPool, tokenMaker,
		cfg.JWT.AccessTokenExp,
		cfg.JWT.RefreshTokenExp)

	// Middlewares
	corsMiddleware := middlewares.CORSMiddleware(cfg)
	panicMiddleware := middlewares.PanicMiddleware()
	errorMiddleware := middlewares.ErrorMiddleware(logger)
	authMiddleware := middlewares.AuthMiddleware(tokenService)
	securityHeadersMiddleware := middlewares.SecurityHeaders()

	// Web App
	app := web.NewApp(panicMiddleware, errorMiddleware, securityHeadersMiddleware, corsMiddleware)

	// Handlers
	authHandler := auth.NewHandler(logger, userService, tokenService, cfg)

	// Routes
	auth.RegisterRoutes(app, authHandler, rateLimiter)
	app.Handle("GET", "/public", func(w http.ResponseWriter, r *http.Request) error {
		return web.RespondOK(w, r, map[string]string{"status": "ok", "message": "Hello, World!"})
	})

	app.Handle("GET", "/healthcheck", func(w http.ResponseWriter, r *http.Request) error {
		user := model.GetUserContext(r.Context())
		if user == nil {
			return web.NewError(errors.New("user not found"), http.StatusUnauthorized)
		}
		return web.RespondOK(w, r, map[string]string{"status": "ok", "user": user.Email, "user_id": user.ID.String()})
	}, authMiddleware)

	app.Handle("GET", "/admin", func(w http.ResponseWriter, r *http.Request) error {
		user := model.GetUserContext(r.Context())
		if user == nil {
			return web.NewError(errors.New("user not found"), http.StatusUnauthorized)
		}
		return web.RespondOK(w, r, map[string]string{"status": "ok", "user": user.Email, "user_id": user.ID.String()})
	}, authMiddleware, middlewares.RequireAdmin())

	return app, nil
}
