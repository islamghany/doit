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
	authService *service.AuthService
	config *config.Config
}

func NewHandler(log *logger.Logger, authService *service.AuthService, config *config.Config) *Handler {
	return &Handler{log: log, authService: authService, config: config}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var input LoginInput
	
	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	user, err := h.authService.AuthenticateUser(ctx, model.LoginInput{
		Email:    input.Email,
		Password: input.Password,
	})

	if err != nil {
		return h.handleAuthError(err)
	}

	return web.RespondOK(w, r, user)
}

// handleAuthError maps service errors to appropriate HTTP responses
func (h *Handler) handleAuthError( err error) error {
	// Map specific service errors to HTTP responses
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		return web.NewError(errors.New("invalid credentials"), http.StatusUnauthorized)
	case errors.Is(err, service.ErrUserInactive):
		return web.NewError(errors.New("user account is inactive"), http.StatusForbidden)
	default:
		return web.NewError(errors.New("internal server error"), http.StatusInternalServerError)
	}
}