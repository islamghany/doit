package auth

import (
	"doit/internal/config"
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
	"errors"
	"net/http"
)


type Handler struct {
	log     *logger.Logger
	userService *service.UserService
	tokenService *service.TokenService
	config *config.Config
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
		IPAddress: web.GetClientIP(r),
		UserAgent: web.GetUserAgent(r),
		DeviceName: web.GetDeviceName(r),
	}

	tokens, err := h.tokenService.CreateTokenPair(ctx, *user, deviceInfo)
	if err != nil {
		return web.NewError(err, http.StatusInternalServerError)
	}
	response := map[string]interface{}{
		"user": user,
		"tokens": tokens,
	}
	return web.RespondOK(w, r, response)
}

// handleAuthError maps service errors to appropriate HTTP responses
func (h *Handler) handleAuthError( err error) error {
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