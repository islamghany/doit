package auth

import "doit/internal/web"

func RegisterRoutes(app *web.WebApp, handler *Handler) {
	app.Handle("POST", "/api/v1/auth/login", handler.Login)
}