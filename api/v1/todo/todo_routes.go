package todo

import (
	"doit/internal/web"
)

func RegisterRoutes(app *web.WebApp, h *Handler, authMiddleware web.MiddleWare) {
	app.Handle("POST", "/api/v1//todos", h.CreateTodo, authMiddleware)
	app.Handle("GET", "/api/v1/todos/{id}", h.GetTodo, authMiddleware)
	app.Handle("PUT", "/api/v1/todos/{id}", h.UpdateTodo, authMiddleware)
	app.Handle("DELETE", "/api/v1/todos/{id}", h.DeleteTodo, authMiddleware)
}
