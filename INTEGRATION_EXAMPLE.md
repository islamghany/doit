# 🔌 Integration Example: Using Fat Services with sqlc + pgx

This guide shows how to integrate the fat services into your API handlers.

## Step 1: Initialize Database & Services in main.go

```go
// cmd/doit/main.go
package main

import (
	"context"
	"doit/api"
	"doit/internal/config"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/database"
	"doit/pkg/logger"
	"fmt"
	"os"
	"time"
)

var build = "development"
var version = "v0.0.1"

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log := setupLogger(cfg)

	log.Info(ctx, "Starting application",
		"build", build,
		"version", version,
		"environment", cfg.App.Environment,
	)

	// Initialize database
	dbPool, err := database.New(ctx, database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		Database:        cfg.Database.Name,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		MaxConns:        cfg.Database.MaxOpenConns,
		MinConns:        5,
		MaxConnLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
		MaxConnIdleTime: 30 * time.Minute,
		DisableTLS:      cfg.Database.DisableTLS,
		LogLevel:        cfg.App.LogLevel,
	})
	if err != nil {
		log.Error(ctx, "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	log.Info(ctx, "Database connection established",
		"host", cfg.Database.Host,
		"database", cfg.Database.Name,
	)

	// Initialize services (Fat Services)
	userService := service.NewUserService(dbPool)
	todoService := service.NewTodoService(dbPool)

	// Start application
	if err := api.Run(ctx, log, cfg, userService, todoService); err != nil {
		log.Error(ctx, "Failed to start the application.", "error", err)
		os.Exit(1)
	}
}

func setupLogger(cfg *config.Config) *logger.Logger {
	var minLevel logger.Level
	switch cfg.App.LogLevel {
	case "debug":
		minLevel = logger.LevelDebug
	case "warn":
		minLevel = logger.LevelWarn
	case "error":
		minLevel = logger.LevelError
	default:
		minLevel = logger.LevelInfo
	}

	traceIDFunc := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			// TODO: Send to error tracking service
		},
	}

	return logger.NewWithEvents(os.Stdout, minLevel, "doit", traceIDFunc, events)
}
```

## Step 2: Update api.go to Accept Services

```go
// api/api.go
package api

import (
	"context"
	"doit/internal/config"
	"doit/internal/service"
	"doit/pkg/logger"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(
	ctx context.Context,
	log *logger.Logger,
	cfg *config.Config,
	userService *service.UserService,
	todoService *service.TodoService,
) error {
	// Create HTTP handler with services
	handler := NewServer(log, cfg, userService, todoService)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Info(ctx, "Starting server",
			"host", cfg.Server.Host,
			"port", cfg.Server.Port,
		)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until error or interrupt
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "Shutdown signal received", "signal", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("could not gracefully shutdown: %w", err)
		}

		log.Info(ctx, "Server shutdown complete")
	}

	return nil
}
```

## Step 3: Update server.go with Services

```go
// api/server.go
package api

import (
	"doit/api/v1"
	"doit/internal/config"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
	"net/http"
)

func NewServer(
	log *logger.Logger,
	cfg *config.Config,
	userService *service.UserService,
	todoService *service.TodoService,
) http.Handler {
	app := web.NewApp()

	// Health check endpoint
	app.Handle("GET", "/healthcheck", func(w http.ResponseWriter, r *http.Request) error {
		return web.RespondOK(w, r, map[string]string{
			"status": "ok",
			"version": "v1.0.0",
		})
	})

	// Mount v1 API routes
	v1.RegisterRoutes(app, log, cfg, userService, todoService)

	return app
}
```

## Step 4: Create v1 Routes

```go
// api/v1/routes.go
package v1

import (
	"doit/api/v1/todo"
	"doit/api/v1/user"
	"doit/internal/config"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
)

func RegisterRoutes(
	app *web.App,
	log *logger.Logger,
	cfg *config.Config,
	userService *service.UserService,
	todoService *service.TodoService,
) {
	// Create handlers (handlers are thin, services are fat)
	userHandler := user.NewHandler(log, userService)
	todoHandler := todo.NewHandler(log, todoService)

	// User routes
	app.Handle("POST", "/api/v1/users", userHandler.CreateUser)
	app.Handle("GET", "/api/v1/users/:id", userHandler.GetUser)
	app.Handle("POST", "/api/v1/users/login", userHandler.Login)

	// Todo routes
	app.Handle("POST", "/api/v1/todos", todoHandler.CreateTodo)
	app.Handle("GET", "/api/v1/todos/:id", todoHandler.GetTodo)
	app.Handle("GET", "/api/v1/todos", todoHandler.ListTodos)
	app.Handle("PATCH", "/api/v1/todos/:id/complete", todoHandler.CompleteTodo)
	app.Handle("GET", "/api/v1/todos/stats", todoHandler.GetStats)
	app.Handle("GET", "/api/v1/todos/search", todoHandler.SearchTodos)
}
```

## Step 5: Example User Handler (Thin Handler)

```go
// api/v1/user/handler.go
package user

import (
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
	"net/http"

	"github.com/google/uuid"
)

// Handler is THIN - it only handles HTTP concerns
// All business logic is in the service
type Handler struct {
	log     *logger.Logger
	service *service.UserService
}

func NewHandler(log *logger.Logger, service *service.UserService) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

// CreateUser handles user creation
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse request directly to model.CreateUserInput
	var input model.CreateUserInput
	if err := web.DecodeJSON(r, &input); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Validate request
	if err := web.Validate.Struct(input); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Call service (business logic in service)
	// Service returns model.User (clean domain type)
	user, err := h.service.CreateUser(ctx, input)
	if err != nil {
		h.log.Error(ctx, "Failed to create user", "error", err)
		return err
	}

	// Return response (model.User automatically serializes to JSON)
	return web.RespondCreated(w, r, user)
}

// GetUser retrieves a user by ID
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse path parameter
	userID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Call service (returns clean model.User)
	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		h.log.Error(ctx, "Failed to get user", "error", err, "user_id", userID)
		return err
	}

	// Return response
	return web.RespondOK(w, r, user)
}

// Login authenticates a user
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse to model.LoginInput
	var input model.LoginInput
	if err := web.DecodeJSON(r, &input); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Call service (handles auth logic)
	user, err := h.service.AuthenticateUser(ctx, input)
	if err != nil {
		h.log.Error(ctx, "Authentication failed", "error", err, "email", input.Email)
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	// In real app, generate JWT here
	response := map[string]interface{}{
		"user":  user,
		"token": "jwt-token-here", // TODO: Generate JWT
	}

	return web.RespondOK(w, r, response)
}
```

## Step 6: Example Todo Handler

```go
// api/v1/todo/handler.go
package todo

import (
	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Handler struct {
	log     *logger.Logger
	service *service.TodoService
}

func NewHandler(log *logger.Logger, service *service.TodoService) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Parse directly to model.CreateTodoInput
	var input model.CreateTodoInput
	if err := web.DecodeJSON(r, &input); err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Get user ID from context (set by auth middleware)
	userID := web.GetUserID(ctx) // Implement this in your web package
	input.UserID = userID

	// Service returns model.Todo (clean domain type)
	todo, err := h.service.CreateTodo(ctx, input)
	if err != nil {
		h.log.Error(ctx, "Failed to create todo", "error", err)
		return err
	}

	return web.RespondCreated(w, r, todo)
}

func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	todoID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	todo, err := h.service.GetTodoByID(ctx, todoID)
	if err != nil {
		return err
	}

	return web.RespondOK(w, r, todo)
}

func (h *Handler) ListTodos(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := web.GetUserID(ctx)

	// Parse pagination
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}

	todos, err := h.service.ListUserTodos(ctx, userID, int32(limit), int32(offset))
	if err != nil {
		return err
	}

	return web.RespondOK(w, r, todos)
}

func (h *Handler) CompleteTodo(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := web.GetUserID(ctx)

	todoID, err := uuid.Parse(web.Param(r, "id"))
	if err != nil {
		return web.NewRequestError(err, http.StatusBadRequest)
	}

	// Service handles transaction, locking, and logging
	todo, err := h.service.CompleteTodo(ctx, todoID, userID)
	if err != nil {
		return err
	}

	return web.RespondOK(w, r, todo)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := web.GetUserID(ctx)

	stats, err := h.service.GetTodoStats(ctx, userID)
	if err != nil {
		return err
	}

	return web.RespondOK(w, r, stats)
}

func (h *Handler) SearchTodos(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID := web.GetUserID(ctx)

	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	todos, err := h.service.SearchTodosByTitle(ctx, userID, query, int32(limit))
	if err != nil {
		return err
	}

	return web.RespondOK(w, r, todos)
}
```

## Key Principles

### ✅ Fat Service Pattern Benefits

1. **Single Source of Truth**: All business logic in services
2. **Transaction Management**: Services handle transactions
3. **Testability**: Test services without HTTP layer
4. **Reusability**: Services can be called from handlers, CLI, jobs, etc.
5. **Clear Boundaries**: Handlers = HTTP, Services = Business Logic

### ✅ Handler Responsibilities (THIN)

- Parse HTTP requests
- Validate input format
- Call service methods
- Format HTTP responses
- Handle HTTP-specific concerns

### ✅ Service Responsibilities (FAT)

- Business logic and validation
- Database operations
- Transaction management
- Complex queries and aggregations
- Cross-entity operations
- Error handling

### ✅ Testing

```go
// Test services directly
func TestTodoService_CompleteTodo(t *testing.T) {
    // Setup test database
    pool := setupTestDB(t)
    service := service.NewTodoService(pool)

    // Test business logic
    todo, err := service.CompleteTodo(ctx, todoID, userID)
    assert.NoError(t, err)
    assert.Equal(t, "completed", todo.Status)
}

// Test handlers for HTTP concerns
func TestHandler_CompleteTodo(t *testing.T) {
    // Use mock service
    mockService := &MockTodoService{}
    handler := NewHandler(log, mockService)

    // Test HTTP handling
    req := httptest.NewRequest("PATCH", "/todos/123/complete", nil)
    w := httptest.NewRecorder()

    err := handler.CompleteTodo(w, req)
    assert.NoError(t, err)
    assert.Equal(t, 200, w.Code)
}
```

## Next Steps

1. Run `go mod tidy` to download dependencies
2. Run `make dev-db` to start PostgreSQL
3. Run `make sqlc` to generate code
4. Start the server: `go run cmd/doit/main.go`
5. Test endpoints with curl or Postman

**You're all set! 🚀**
