package user

import (
	"context"
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type Handler struct {
	log         *logger.Logger
	userService *service.UserService
}

func NewHandler(log *logger.Logger, userService *service.UserService) *Handler {
	return &Handler{
		log:         log,
		userService: userService,
	}
}

// GetUser retrieves a user by ID
// Example: GET /api/v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Extract ID from URL path (simple extraction without chi)
	// In production, this would be handled by your router
	pathParts := strings.Split(r.URL.Path, "/")
	userIDStr := pathParts[len(pathParts)-1]
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return web.NewError(errors.New("invalid user ID format"), http.StatusBadRequest)
	}

	// Get user from service
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return h.handleUserError(ctx, err)
	}

	return web.RespondOK(w, r, user)
}

// CreateUser creates a new user
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var input CreateUserRequest
	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	// Create user via service
	user, err := h.userService.CreateUser(ctx, model.CreateUserInput{
		Email:    input.Email,
		Username: input.Username,
		Password: input.Password,
		Metadata: input.Metadata,
	})
	if err != nil {
		return h.handleUserError(ctx, err)
	}

	return web.RespondCreated(w, r, user)
}

// UpdateUser updates an existing user
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	userIDStr := pathParts[len(pathParts)-1]
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return web.NewError(errors.New("invalid user ID format"), http.StatusBadRequest)
	}

	var input UpdateUserRequest
	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	// Update user via service
	user, err := h.userService.UpdateUser(ctx, userID, model.UpdateUserInput{
		Email:    input.Email,
		Username: input.Username,
		IsActive: input.IsActive,
		Metadata: input.Metadata,
	})
	if err != nil {
		return h.handleUserError(ctx, err)
	}

	return web.RespondOK(w, r, user)
}

// ListUsers retrieves a paginated list of users
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// TODO: Parse query parameters for pagination
	filter := model.UserFilter{
		Limit:  10,
		Offset: 0,
	}

	users, err := h.userService.ListUsers(ctx, filter)
	if err != nil {
		return h.handleUserError(ctx, err)
	}

	return web.RespondOK(w, r, users)
}

// handleUserError maps service errors to appropriate HTTP responses
// This is the recommended pattern for clean error handling
func (h *Handler) handleUserError(ctx context.Context, err error) error {
	// Check for specific error messages or types
	errMsg := err.Error()
	
	switch {
	case strings.Contains(errMsg, "user not found"):
		return web.NewError(errors.New("user not found"), http.StatusNotFound)
	case errors.Is(err, service.ErrDuplicateEmail):
		return web.NewError(errors.New("email already exists"), http.StatusConflict)
	case errors.Is(err, service.ErrInvalidInput):
		return web.NewError(errors.New("invalid input"), http.StatusBadRequest)
	case strings.Contains(errMsg, "validation failed"):
		return web.NewError(err, http.StatusBadRequest)
	default:
		// Log unexpected errors for debugging
		h.log.Error(ctx, "user service error", "error", err)
		return web.NewError(errors.New("internal server error"), http.StatusInternalServerError)
	}
}
