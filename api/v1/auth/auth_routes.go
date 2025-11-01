package auth

import "doit/internal/web"

func RegisterRoutes(app *web.WebApp, handler *Handler) {
	app.Handle("POST", "/api/v1/auth/register", handler.Register)
	app.Handle("POST", "/api/v1/auth/login", handler.Login)
	app.Handle("POST", "/api/v1/auth/refresh", handler.Refresh)
	app.Handle("POST", "/api/v1/auth/logout", handler.Logout)
	app.Handle("POST", "/api/v1/auth/logout-all", handler.LogoutAll)
}
