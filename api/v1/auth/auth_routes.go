package auth

import (
	"doit/internal/limiter"
	"doit/internal/middlewares"
	"doit/internal/web"
)

func RegisterRoutes(app *web.WebApp, handler *Handler, rateLimiter *limiter.RateLimiter) {
	registerMiddleware := middlewares.RegisterRateLimitMiddleware(rateLimiter)
	loginMiddleware := middlewares.LoginRateLimitMiddleware(rateLimiter)
	refreshMiddleware := middlewares.RefreshRateLimitMiddleware(rateLimiter)

	app.Handle("POST", "/api/v1/auth/register", handler.Register, registerMiddleware)
	app.Handle("POST", "/api/v1/auth/login", handler.Login, loginMiddleware)
	app.Handle("POST", "/api/v1/auth/refresh", handler.Refresh, refreshMiddleware)
	app.Handle("POST", "/api/v1/auth/logout", handler.Logout)
	app.Handle("POST", "/api/v1/auth/logout-all", handler.LogoutAll)
}
