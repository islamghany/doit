package middlewares

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"doit/internal/config"
	"doit/internal/web"
)

func CORSMiddleware(config *config.Config) web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		return func(w http.ResponseWriter, r *http.Request) error {
			origin := r.Header.Get("Origin")

			isAllOriginsAllowed := slices.Contains(config.Security.AllowedOrigins, "*")
			if !isAllOriginsAllowed && !slices.Contains(config.Security.AllowedOrigins, origin) {
				return web.NewError(errors.New("origin not allowed"), http.StatusForbidden)
			}

			w.Header().Set("Access-Control-Allow-Origin", strings.Join(config.Security.AllowedOrigins, ","))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return nil
			}

			return handler(w, r)
		}
	}
}
