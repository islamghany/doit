package api

import (
	"doit/internal/config"
	"doit/internal/web"
	"doit/pkg/logger"
	"net/http"
)

func NewServer(logger *logger.Logger, cfg *config.Config) http.Handler {
	app := web.NewApp()

	app.Handle("GET", "/healthcheck", func(w http.ResponseWriter, r *http.Request) error {
		return web.RespondOK(w, r, map[string]string{"status": "ok"})
	})

	return app
}
