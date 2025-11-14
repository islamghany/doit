package healthcheck

import "doit/internal/web"

func RegisterRoutes(app *web.WebApp, handler *Handler) {
	app.Handle("GET", "/health", handler.HealthCheck)
	app.Handle("GET", "/ready", handler.ReadyCheck)
}
