package middlewares

import (
	"errors"
	"net/http"
	"strings"

	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
)

func AuthMiddleware(tokenService *service.TokenService) web.MiddleWare {
	return func(handler web.Handler) web.Handler {
		h := func(w http.ResponseWriter, r *http.Request) error {
			// 1. Get the authorization header
			authHeader := web.GetHeader(r, "Authorization")
			if authHeader == "" {
				return web.NewError(errors.New("access token is required"), http.StatusUnauthorized)
			}

			// 2. Extract the access token with bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return web.NewError(errors.New("invalid authorization header"), http.StatusUnauthorized)
			}
			accessToken := parts[1]

			// 3. Verify the access token
			payload, err := tokenService.VerifyAccessToken(r.Context(), accessToken)
			if err != nil {
				return web.NewError(err, http.StatusUnauthorized)
			}

			// 4. Set the user context
			ctx := model.SetUserContext(r.Context(), &model.User{
				ID:           payload.UserID,
				Email:        payload.Email,
				Username:     payload.Username,
				TokenVersion: int32(payload.Version),
			})
			r = r.WithContext(ctx)
			return handler(w, r)
		}
		return h
	}
}
