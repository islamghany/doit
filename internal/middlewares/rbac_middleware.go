// Package middlewares provides HTTP middleware functions for the API.
package middlewares

import (
	"errors"
	"net/http"

	"doit/internal/model"
	"doit/internal/web"
)

// RequireRole creates middleware that checks if user has one of the required roles
func RequireRole(roles ...model.UserRole) web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		h := func(w http.ResponseWriter, r *http.Request) error {
			user, err := model.GetUserFromContext(r.Context())
			if err != nil {
				return web.NewError(errors.New("unauthorized"), http.StatusUnauthorized)
			}

			// Check if user has one of the required roles
			hasRole := false
			for _, role := range roles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return web.NewError(
					errors.New("forbidden: insufficient permissions"),
					http.StatusForbidden,
				)
			}

			return handler(w, r)
		}
		return h
	}
}

// RequireAdmin is a convenience middleware for admin-only endpoints
func RequireAdmin() web.MiddleWare {
	return RequireRole(model.UserRoleAdmin)
}

// RequireAdminOrModerator allows both admin and moderator roles
func RequireAdminOrModerator() web.MiddleWare {
	return RequireRole(model.UserRoleAdmin, model.UserRoleModerator)
}
