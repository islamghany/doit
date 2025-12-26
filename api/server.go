package api

import (
	"errors"
	"net/http"

	"doit/api/v1/auth"
	healthcheck "doit/api/v1/health_check"
	"doit/api/v1/todo"
	_ "doit/docs" // Import generated swagger docs
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

	httpSwagger "github.com/swaggo/http-swagger"
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
	todoService := service.NewTodoService(dbPool)

	// Middlewares
	corsMiddleware := middlewares.CORSMiddleware(cfg)
	panicMiddleware := middlewares.PanicMiddleware()
	errorMiddleware := middlewares.ErrorMiddleware(logger)
	authMiddleware := middlewares.AuthMiddleware(tokenService)
	securityHeadersMiddleware := middlewares.SecurityHeaders()
	metricsMiddleware := middlewares.Metrics()
	tracingMiddleware := middlewares.Tracing()

	// Web App
	app := web.NewApp(
		tracingMiddleware,         // 1. Outermost - captures everything
		errorMiddleware,           // 2. Error handling
		metricsMiddleware,         // 3. Prometheus metrics
		panicMiddleware,           // 4. Panic recovery
		securityHeadersMiddleware, // 5. Security headers
		corsMiddleware,            // 6. CORS handling
	)

	// Handlers
	authHandler := auth.NewHandler(logger, userService, tokenService, cfg)
	healthcheckHandler := healthcheck.NewHandler(logger, dbPool, cache, cfg.App.Version)
	todoHandler := todo.NewHandler(logger, todoService)

	// Routes
	auth.RegisterRoutes(app, authHandler, rateLimiter)
	healthcheck.RegisterRoutes(app, healthcheckHandler)
	todo.RegisterRoutes(app, todoHandler, authMiddleware)

	// Swagger documentation endpoint
	// Access at: http://localhost:8080/swagger/index.html
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)
	app.Handle("GET", "/swagger/*", func(w http.ResponseWriter, r *http.Request) error {
		swaggerHandler.ServeHTTP(w, r)
		return nil
	})

	app.Handle("GET", "/public", func(w http.ResponseWriter, r *http.Request) error {
		logger.Info(r.Context(), "Public endpoint hit")
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
