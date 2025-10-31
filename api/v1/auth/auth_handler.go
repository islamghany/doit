package auth

import (
	"errors"
	"net/http"

	"doit/internal/config"
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/token"
	"doit/internal/web"
	"doit/pkg/logger"

	"github.com/google/uuid"
)

type Handler struct {
	log          *logger.Logger
	userService  *service.UserService
	tokenService *service.TokenService
	config       *config.Config
}

func NewHandler(log *logger.Logger, userService *service.UserService, tokenService *service.TokenService, config *config.Config) *Handler {
	return &Handler{log: log, userService: userService, tokenService: tokenService, config: config}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var input LoginInput

	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	user, err := h.userService.AuthenticateUser(ctx, model.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return h.handleAuthError(err)
	}

	deviceInfo := model.DeviceInfo{
		IPAddress:  web.GetClientIP(r),
		UserAgent:  web.GetUserAgent(r),
		DeviceName: web.GetDeviceName(r),
	}

	tokens, err := h.tokenService.CreateTokenPair(ctx, *user, deviceInfo)
	if err != nil {
		return web.NewError(err, http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	}
	return web.RespondOK(w, r, response)
}

// Refresh exchanges a refresh token for a new access token
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// 1. Get the refresh token from the header
	refreshToken := web.GetHeader(r, "X-Refresh-Token")
	if refreshToken == "" {
		return web.NewError(errors.New("refresh token is required"), http.StatusBadRequest)
	}

	// 2. Refresh tokens (includes security checks)
	payload, err := h.tokenService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		// if token not valid
		if errors.Is(err, token.ErrInvalidToken) {
			return web.NewError(errors.New("invalid token"), http.StatusUnauthorized)
		}
		// security alert
		if errors.Is(err, service.ErrSecurityAlert) {
			return web.NewError(errors.New("invalid token"), http.StatusUnauthorized)
		}
		return web.NewError(err, http.StatusInternalServerError)
	}

	// 3. Return the new access token
	response := map[string]interface{}{
		"access_token":  payload.AccessToken,
		"refresh_token": payload.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    payload.ExpiresIn,
	}
	return web.RespondOK(w, r, response)
}

// Logout revokes the current refresh token
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// 1. Get the refresh token from the header
	refreshToken := web.GetHeader(r, "X-Refresh-Token")
	if refreshToken == "" {
		return web.NewError(errors.New("refresh token is required"), http.StatusBadRequest)
	}

	err := h.tokenService.Logout(ctx, refreshToken)
	if err != nil {
		return web.NewError(err, http.StatusInternalServerError)
	}

	return web.RespondOK(w, r, nil)
}

// LogoutAll revokes all tokens for the current user
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) error {
	// userID := web.GetUserID(r)
	// 2. Revoke all user tokens
	err := h.tokenService.RevokeAllUserTokens(r.Context(), uuid.New())
	if err != nil {
		return web.NewError(err, http.StatusInternalServerError)
	}

	return web.RespondOK(w, r, nil)
}

// handleAuthError maps service errors to appropriate HTTP responses
func (h *Handler) handleAuthError(err error) error {
	// Map specific service errors to HTTP responses
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return web.NewError(errors.New("invalid credentials"), http.StatusUnauthorized)
	// case errors.Is(err, service.ErrUserInactive):
	// 	return web.NewError(errors.New("user account is inactive"), http.StatusForbidden)
	default:
		return web.NewError(errors.New("internal server error"), http.StatusInternalServerError)
	}
}
