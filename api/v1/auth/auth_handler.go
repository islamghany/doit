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

// Register godoc
// @Summary      Register a new user account
// @Description  Create a new user account with email, username, and password. Passwords must be at least 8 characters and should include uppercase, lowercase, numbers, and special characters. Email and username must be unique.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body RegisterInput true "User registration information"
// @Success      200 {object} model.User "Successfully created user account"
// @Failure      400 {object} ErrorResponse "Invalid input, duplicate email, or duplicate username"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var input RegisterInput
	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	// Create user
	user, err := h.userService.CreateUser(ctx, model.CreateUserInput{
		Email:    input.Email,
		Username: input.Username,
		Password: input.Password,
		Role:     model.UserRole(input.Role),
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateEmail):
			return web.NewError(errors.New("email already exists"), http.StatusBadRequest)
		case errors.Is(err, service.ErrDuplicateUsername):
			return web.NewError(errors.New("username already exists"), http.StatusBadRequest)
		case errors.Is(err, service.ErrFailedToCreateUser):
			return web.NewError(errors.New("failed to create user"), http.StatusInternalServerError)
		default:
			return web.NewError(err, http.StatusInternalServerError)
		}
	}

	return web.RespondOK(w, r, user)
}

// Login godoc
// @Summary      Authenticate and login user
// @Description  Authenticate user credentials and receive JWT access token and refresh token. The access token is short-lived (15 minutes) and used for API requests. The refresh token is long-lived (7 days) and used to obtain new access tokens.
// @Description  Rate limit: 5 requests per minute per IP address
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        credentials body LoginInput true "User login credentials"
// @Success      200 {object} LoginResponse "Login successful with user info and tokens"
// @Failure      400 {object} ErrorResponse "Invalid request body"
// @Failure      401 {object} ErrorResponse "Invalid credentials (wrong email or password)"
// @Failure      429 {object} ErrorResponse "Rate limit exceeded"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/login [post]
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

// Refresh godoc
// @Summary      Refresh access token
// @Description  Exchange a valid refresh token for a new access token and refresh token pair. This implements refresh token rotation for enhanced security - the old refresh token becomes invalid after use.
// @Description  Security: If an already-used refresh token is detected, all tokens for that user are revoked (possible token theft).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        X-Refresh-Token header string true "Refresh token from login response"
// @Success      200 {object} RefreshResponse "New token pair generated successfully"
// @Failure      400 {object} ErrorResponse "Missing refresh token header"
// @Failure      401 {object} ErrorResponse "Invalid, expired, or already-used refresh token"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/refresh [post]
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

// Logout godoc
// @Summary      Logout user (single device)
// @Description  Revoke the current refresh token, logging out the user from this specific device/session. Other active sessions remain valid.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        X-Refresh-Token header string true "Refresh token to revoke"
// @Success      200 {object} MessageResponse "Logged out successfully"
// @Failure      400 {object} ErrorResponse "Missing refresh token header"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/logout [post]
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

// LogoutAll godoc
// @Summary      Logout user from all devices
// @Description  Revoke all refresh tokens for the authenticated user, logging them out from all devices and sessions. Use this when account security is compromised.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     bearer
// @Success      200 {object} MessageResponse "Logged out from all devices successfully"
// @Failure      401 {object} ErrorResponse "Unauthorized - missing or invalid access token"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /auth/logout/all [post]
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
