# OWASP Top 10 Security Guide for doit API

> **Complete security implementation guide based on OWASP Top 10 (2021)**
>
> This document explains each security risk and provides actionable implementation steps for our Go REST API.

---

## 📋 Table of Contents

1. [A01:2021 - Broken Access Control](#a012021---broken-access-control)
2. [A02:2021 - Cryptographic Failures](#a022021---cryptographic-failures)
3. [A03:2021 - Injection](#a032021---injection)
4. [A04:2021 - Insecure Design](#a042021---insecure-design)
5. [A05:2021 - Security Misconfiguration](#a052021---security-misconfiguration)
6. [A06:2021 - Vulnerable and Outdated Components](#a062021---vulnerable-and-outdated-components)
7. [A07:2021 - Identification and Authentication Failures](#a072021---identification-and-authentication-failures)
8. [A08:2021 - Software and Data Integrity Failures](#a082021---software-and-data-integrity-failures)
9. [A09:2021 - Security Logging and Monitoring Failures](#a092021---security-logging-and-monitoring-failures)
10. [A10:2021 - Server-Side Request Forgery (SSRF)](#a102021---server-side-request-forgery-ssrf)

---

## A01:2021 - Broken Access Control

### 🔴 **CRITICAL - Highest Priority**

### What It Is

Users can access resources they shouldn't have permission to access. This includes:

- Viewing or modifying other users' data
- Accessing admin functions without authorization
- Bypassing access control checks by modifying URLs, IDs, or internal state

### Real Examples in Your Project

**VULNERABLE CODE:**

```go
// ❌ BAD: Any authenticated user can get any todo by ID
func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) error {
    todoID := uuid.MustParse(chi.URLParam(r, "id"))
    todo, err := h.service.GetTodoByID(ctx, todoID)
    return web.RespondOK(w, r, todo)
}
```

**Attack Scenario:**

```bash
# User A creates todo with ID: abc-123
# User B can access it by guessing/enumerating IDs:
curl -H "Authorization: Bearer <user-b-token>" \
     https://api.example.com/todos/abc-123
# ❌ Returns User A's todo!
```

### ✅ Current Status in Your Project

**Good practices already implemented:**

- ✅ JWT authentication middleware (`auth_middleware.go`)
- ✅ User context propagation
- ✅ `CompleteTodo` has ownership verification (line 217-219 in `todo_service.go`)
- ✅ `DeleteTodo` and `BulkDeleteTodos` check userID

**Missing protections:**

- ❌ `GetTodoByID` doesn't verify ownership
- ❌ `UpdateTodo` doesn't verify ownership
- ❌ No role-based access control (RBAC)
- ❌ `GetOverdueTodos` accessible to all users (should be admin-only)

### 🛠️ Implementation Plan

#### Task 1: Add Ownership Verification to GetTodoByID

**Create new service method:**

```go
// internal/service/todo_service.go

// GetTodoByIDWithOwnership retrieves a todo and verifies ownership
func (s *TodoService) GetTodoByIDWithOwnership(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
    todo, err := s.querier.GetTodoByID(ctx, todoID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("todo not found")
        }
        return nil, fmt.Errorf("failed to get todo: %w", err)
    }

    // Verify ownership
    if todo.UserID != userID {
        return nil, fmt.Errorf("unauthorized: todo does not belong to user")
    }

    return s.toTodoModel(todo), nil
}
```

#### Task 2: Add Ownership Verification to UpdateTodo

**Update existing method:**

```go
// internal/service/todo_service.go

// UpdateTodo updates a todo with ownership verification
func (s *TodoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, input model.UpdateTodoInput) (*model.Todo, error) {
    // First, verify ownership
    existingTodo, err := s.querier.GetTodoByID(ctx, todoID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("todo not found")
        }
        return nil, fmt.Errorf("failed to get todo: %w", err)
    }

    if existingTodo.UserID != userID {
        return nil, fmt.Errorf("unauthorized: todo does not belong to user")
    }

    // Build update params (rest of existing code)
    params := db.UpdateTodoParams{
        ID: todoID,
    }

    // ... rest of your existing update logic
}
```

#### Task 3: Implement Role-Based Access Control (RBAC)

**Step 1: Add roles to users table**

```sql
-- internal/data/migrations/000004_add_user_roles.up.sql

ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index for role-based queries
CREATE INDEX idx_users_role ON users(role);

-- Add check constraint
ALTER TABLE users ADD CONSTRAINT check_user_role
    CHECK (role IN ('user', 'admin', 'moderator'));
```

```sql
-- internal/data/migrations/000004_add_user_roles.down.sql

ALTER TABLE users DROP CONSTRAINT check_user_role;
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN role;
```

**Step 2: Update user model**

```go
// internal/model/user.go

type UserRole string

const (
    UserRoleUser      UserRole = "user"
    UserRoleAdmin     UserRole = "admin"
    UserRoleModerator UserRole = "moderator"
)

type User struct {
    ID           uuid.UUID `json:"id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    Role         UserRole  `json:"role"`          // ADD THIS
    TokenVersion int32     `json:"token_version"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

**Step 3: Create authorization middleware**

```go
// internal/middlewares/rbac_middleware.go

package middlewares

import (
    "errors"
    "net/http"

    "doit/internal/model"
    "doit/internal/web"
)

// RequireRole creates middleware that checks if user has required role
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
```

**Step 4: Apply to admin routes**

```go
// api/v1/todo/todo_routes.go

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware web.MiddleWare) {
    r.Route("/todos", func(r chi.Router) {
        // Public routes (none for todos)

        // Protected user routes
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)

            r.Get("/", handler.ListTodos)
            r.Post("/", handler.CreateTodo)
            r.Get("/{id}", handler.GetTodo)
            r.Put("/{id}", handler.UpdateTodo)
            r.Delete("/{id}", handler.DeleteTodo)
        })

        // Admin-only routes
        r.Group(func(r chi.Router) {
            r.Use(authMiddleware)
            r.Use(middlewares.RequireAdmin()) // ADD THIS

            r.Get("/admin/overdue", handler.GetOverdueTodos)
            r.Get("/admin/stats", handler.GetAllUsersStats)
            r.Delete("/admin/bulk-delete", handler.AdminBulkDelete)
        })
    })
}
```

#### Task 4: Prevent IDOR (Insecure Direct Object Reference)

**Add resource ownership check helper:**

```go
// internal/service/authorization.go

package service

import (
    "fmt"
    "github.com/google/uuid"
)

// VerifyOwnership checks if a resource belongs to a user
func VerifyOwnership(resourceUserID, requestUserID uuid.UUID, resourceType string) error {
    if resourceUserID != requestUserID {
        return fmt.Errorf("unauthorized: %s does not belong to user", resourceType)
    }
    return nil
}

// VerifyOwnershipOrAdmin checks ownership or admin role
func VerifyOwnershipOrAdmin(resourceUserID, requestUserID uuid.UUID, isAdmin bool, resourceType string) error {
    if isAdmin {
        return nil // Admins can access anything
    }
    return VerifyOwnership(resourceUserID, requestUserID, resourceType)
}
```

---

## A02:2021 - Cryptographic Failures

### 🔴 **CRITICAL**

### What It Is

Failure to properly protect sensitive data through encryption. This includes:

- Storing passwords in plain text
- Using weak hashing algorithms (MD5, SHA1)
- Not using TLS/HTTPS
- Exposing sensitive data in logs or error messages
- Using weak encryption keys

### ✅ Current Status in Your Project

**Good practices:**

- ✅ Password hashing with bcrypt (`pkg/password_hash/hash.go`)
- ✅ JWT tokens for authentication
- ✅ Refresh token storage in database

**Missing protections:**

- ❌ No TLS/HTTPS configuration
- ❌ No encryption of sensitive data at rest (e.g., metadata field)
- ❌ Secrets might be in environment variables (better: use AWS Secrets Manager)
- ❌ No password strength requirements

### 🛠️ Implementation Plan

#### Task 1: Enforce Strong Password Requirements

```go
// pkg/validator/password.go

package validator

import (
    "fmt"
    "regexp"
    "unicode"
)

type PasswordStrength struct {
    MinLength        int
    RequireUppercase bool
    RequireLowercase bool
    RequireNumber    bool
    RequireSpecial   bool
}

var DefaultPasswordStrength = PasswordStrength{
    MinLength:        12,
    RequireUppercase: true,
    RequireLowercase: true,
    RequireNumber:    true,
    RequireSpecial:   true,
}

func ValidatePasswordStrength(password string, rules PasswordStrength) error {
    if len(password) < rules.MinLength {
        return fmt.Errorf("password must be at least %d characters long", rules.MinLength)
    }

    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if rules.RequireUppercase && !hasUpper {
        return fmt.Errorf("password must contain at least one uppercase letter")
    }
    if rules.RequireLowercase && !hasLower {
        return fmt.Errorf("password must contain at least one lowercase letter")
    }
    if rules.RequireNumber && !hasNumber {
        return fmt.Errorf("password must contain at least one number")
    }
    if rules.RequireSpecial && !hasSpecial {
        return fmt.Errorf("password must contain at least one special character")
    }

    // Check against common passwords
    if IsCommonPassword(password) {
        return fmt.Errorf("password is too common, please choose a stronger password")
    }

    return nil
}

var commonPasswords = map[string]bool{
    "password123":    true,
    "Password123":    true,
    "Password123!":   true,
    "Admin123":       true,
    "Welcome123":     true,
    "Qwerty123":      true,
    // Add more from: https://github.com/danielmiessler/SecLists
}

func IsCommonPassword(password string) bool {
    return commonPasswords[password]
}
```

**Apply in user service:**

```go
// internal/service/user_service.go

func (s *UserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // Validate password strength
    if err := validator.ValidatePasswordStrength(input.Password, validator.DefaultPasswordStrength); err != nil {
        return nil, fmt.Errorf("password validation failed: %w", err)
    }

    // Rest of your existing logic...
}
```

#### Task 2: Configure HTTPS/TLS

```go
// api/server.go

package api

import (
    "crypto/tls"
    "net/http"
    "time"
)

// StartTLS starts the server with TLS
func (s *Server) StartTLS(certFile, keyFile string) error {
    // Configure TLS
    tlsConfig := &tls.Config{
        MinVersion:               tls.VersionTLS13, // Only TLS 1.3
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_AES_128_GCM_SHA256,
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
    }

    server := &http.Server{
        Addr:         s.config.Server.Address,
        Handler:      s.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
        TLSConfig:    tlsConfig,
    }

    s.logger.Info("starting HTTPS server", "address", s.config.Server.Address)
    return server.ListenAndServeTLS(certFile, keyFile)
}
```

#### Task 3: Sanitize Sensitive Data from Logs

```go
// internal/web/response.go

// Never log passwords, tokens, or sensitive fields
func (e *Error) LogSafe() map[string]interface{} {
    return map[string]interface{}{
        "status":  e.Status,
        "message": e.Message,
        // ❌ Don't log: error details, stack traces, user data
    }
}

// internal/web/request.go

// Redact sensitive fields before logging
func RedactSensitiveFields(data map[string]interface{}) map[string]interface{} {
    sensitiveFields := []string{"password", "token", "secret", "api_key", "credit_card"}

    redacted := make(map[string]interface{})
    for k, v := range data {
        if contains(sensitiveFields, k) {
            redacted[k] = "[REDACTED]"
        } else {
            redacted[k] = v
        }
    }
    return redacted
}
```

#### Task 4: Rotate JWT Secrets

```go
// internal/config/config.go

type JWTConfig struct {
    AccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
    RefreshSecret string        `env:"JWT_REFRESH_SECRET,required"`
    SecretVersion int           `env:"JWT_SECRET_VERSION" envDefault:"1"`
    AccessTTL     time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
    RefreshTTL    time.Duration `env:"JWT_REFRESH_TTL" envDefault:"7d"`
}

// Support for multiple versions during rotation
type JWTSecrets struct {
    Current  string
    Previous string // For graceful rotation
    Version  int
}
```

---

## A03:2021 - Injection

### 🔴 **CRITICAL**

### What It Is

Untrusted data is sent to an interpreter as part of a command or query. Includes:

- SQL Injection
- NoSQL Injection
- OS Command Injection
- LDAP Injection

### ✅ Current Status in Your Project

**Good practices:**

- ✅ Using `sqlc` with parameterized queries (automatically prevents SQL injection)
- ✅ Using `pgx` which uses prepared statements

**Potential risks:**

- ⚠️ User input in search queries
- ⚠️ Metadata field accepts arbitrary JSON

### 🛠️ Implementation Plan

#### Task 1: Validate and Sanitize All Inputs

```go
// pkg/validator/input.go

package validator

import (
    "fmt"
    "regexp"
    "strings"
)

var (
    // Alphanumeric with common punctuation
    SafeTextRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\_\.\,\!\?\'\"]+$`)

    // For search queries
    SearchQueryRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-\_]+$`)

    // SQL injection patterns to reject
    SQLInjectionPatterns = []string{
        "--", "/*", "*/", "xp_", "sp_",
        "';", "\";", "OR ", "AND ",
        "UNION ", "SELECT ", "DROP ",
        "INSERT ", "UPDATE ", "DELETE ",
    }
)

func ValidateSearchQuery(query string) error {
    if len(query) > 100 {
        return fmt.Errorf("search query too long")
    }

    // Check for SQL injection patterns
    upperQuery := strings.ToUpper(query)
    for _, pattern := range SQLInjectionPatterns {
        if strings.Contains(upperQuery, pattern) {
            return fmt.Errorf("invalid characters in search query")
        }
    }

    return nil
}

func SanitizeString(input string, maxLength int) string {
    // Remove null bytes
    input = strings.ReplaceAll(input, "\x00", "")

    // Limit length
    if len(input) > maxLength {
        input = input[:maxLength]
    }

    // Trim whitespace
    return strings.TrimSpace(input)
}
```

**Apply to search handler:**

```go
// internal/service/todo_service.go

func (s *TodoService) SearchTodosByTitle(ctx context.Context, userID uuid.UUID, query string, limit int32) ([]*model.Todo, error) {
    // Validate search query
    if err := validator.ValidateSearchQuery(query); err != nil {
        return nil, fmt.Errorf("invalid search query: %w", err)
    }

    // Sanitize
    query = validator.SanitizeString(query, 100)

    // Your existing search logic...
}
```

#### Task 2: Validate JSON Metadata

```go
// pkg/validator/json.go

package validator

import (
    "encoding/json"
    "fmt"
)

const (
    MaxJSONDepth = 5
    MaxJSONSize  = 10 * 1024 // 10KB
)

func ValidateJSON(data interface{}) error {
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }

    if len(jsonBytes) > MaxJSONSize {
        return fmt.Errorf("JSON too large (max %d bytes)", MaxJSONSize)
    }

    // Check depth
    if err := checkJSONDepth(data, 0); err != nil {
        return err
    }

    return nil
}

func checkJSONDepth(data interface{}, currentDepth int) error {
    if currentDepth > MaxJSONDepth {
        return fmt.Errorf("JSON too deeply nested (max depth %d)", MaxJSONDepth)
    }

    switch v := data.(type) {
    case map[string]interface{}:
        for _, value := range v {
            if err := checkJSONDepth(value, currentDepth+1); err != nil {
                return err
            }
        }
    case []interface{}:
        for _, value := range v {
            if err := checkJSONDepth(value, currentDepth+1); err != nil {
                return err
            }
        }
    }

    return nil
}
```

---

## A04:2021 - Insecure Design

### 🟡 **MEDIUM**

### What It Is

Missing or ineffective security controls in the design phase. This is about architectural flaws, not implementation bugs.

### Examples in Your Context

- No rate limiting (allow brute force attacks)
- No account lockout after failed login attempts
- No email verification (anyone can register with any email)
- No audit logging (can't track who did what)

### 🛠️ Implementation Plan

#### Task 1: Implement Rate Limiting

```go
// internal/middlewares/rate_limit_middleware.go

package middlewares

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"

    "doit/internal/web"
    "github.com/go-redis/redis/v8"
)

type RateLimiter struct {
    redis *redis.Client
    // Fallback to in-memory if Redis unavailable
    memory     map[string]*rateLimitEntry
    memoryLock sync.RWMutex
}

type rateLimitEntry struct {
    count      int
    resetTime  time.Time
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
    return &RateLimiter{
        redis:  redis,
        memory: make(map[string]*rateLimitEntry),
    }
}

// RateLimitMiddleware limits requests per IP address
func (rl *RateLimiter) RateLimitMiddleware(maxRequests int, window time.Duration) web.MiddleWare {
    return func(handler web.Handler) web.Handler {
        h := func(w http.ResponseWriter, r *http.Request) error {
            ip := web.GetClientIP(r)
            key := fmt.Sprintf("rate_limit:%s", ip)

            allowed, remaining, resetTime, err := rl.checkRateLimit(r.Context(), key, maxRequests, window)
            if err != nil {
                // Log error but don't block request
                // (fail open, not fail closed)
                return handler(w, r)
            }

            // Add rate limit headers
            w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
            w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

            if !allowed {
                w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(resetTime).Seconds())))
                return web.NewError(
                    fmt.Errorf("rate limit exceeded"),
                    http.StatusTooManyRequests,
                )
            }

            return handler(w, r)
        }
        return h
    }
}

func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, max int, window time.Duration) (bool, int, time.Time, error) {
    if rl.redis != nil {
        return rl.checkRateLimitRedis(ctx, key, max, window)
    }
    return rl.checkRateLimitMemory(key, max, window)
}

func (rl *RateLimiter) checkRateLimitRedis(ctx context.Context, key string, max int, window time.Duration) (bool, int, time.Time, error) {
    pipe := rl.redis.Pipeline()

    // Increment counter
    incr := pipe.Incr(ctx, key)

    // Set expiry only on first request
    pipe.Expire(ctx, key, window)

    // Get TTL
    ttl := pipe.TTL(ctx, key)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, 0, time.Time{}, err
    }

    count := incr.Val()
    remaining := max - int(count)
    if remaining < 0 {
        remaining = 0
    }

    resetTime := time.Now().Add(ttl.Val())
    allowed := count <= int64(max)

    return allowed, remaining, resetTime, nil
}

func (rl *RateLimiter) checkRateLimitMemory(key string, max int, window time.Duration) (bool, int, time.Time, error) {
    rl.memoryLock.Lock()
    defer rl.memoryLock.Unlock()

    now := time.Now()
    entry, exists := rl.memory[key]

    if !exists || now.After(entry.resetTime) {
        // New window
        rl.memory[key] = &rateLimitEntry{
            count:     1,
            resetTime: now.Add(window),
        }
        return true, max - 1, now.Add(window), nil
    }

    entry.count++
    remaining := max - entry.count
    if remaining < 0 {
        remaining = 0
    }
    allowed := entry.count <= max

    return allowed, remaining, entry.resetTime, nil
}

// Per-user rate limiting (stricter for auth endpoints)
func (rl *RateLimiter) UserRateLimitMiddleware(maxRequests int, window time.Duration) web.MiddleWare {
    return func(handler web.Handler) web.Handler {
        h := func(w http.ResponseWriter, r *http.Request) error {
            user, err := model.GetUserFromContext(r.Context())
            if err != nil {
                // Not authenticated, skip
                return handler(w, r)
            }

            key := fmt.Sprintf("rate_limit:user:%s", user.ID.String())

            allowed, remaining, resetTime, err := rl.checkRateLimit(r.Context(), key, maxRequests, window)
            if err != nil {
                return handler(w, r)
            }

            w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
            w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
            w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

            if !allowed {
                return web.NewError(
                    fmt.Errorf("rate limit exceeded"),
                    http.StatusTooManyRequests,
                )
            }

            return handler(w, r)
        }
        return h
    }
}
```

**Apply to routes:**

```go
// api/v1/routes.go

func RegisterRoutes(r chi.Router, handler *Handler, rateLimiter *middlewares.RateLimiter) {
    // Auth endpoints: strict rate limiting
    r.Group(func(r chi.Router) {
        // 5 login attempts per 15 minutes per IP
        r.Use(rateLimiter.RateLimitMiddleware(5, 15*time.Minute))

        r.Post("/auth/login", handler.Login)
        r.Post("/auth/register", handler.Register)
    })

    // General API: 100 requests per minute per IP
    r.Group(func(r chi.Router) {
        r.Use(rateLimiter.RateLimitMiddleware(100, 1*time.Minute))
        r.Mount("/todos", todoRoutes)
    })
}
```

#### Task 2: Implement Account Lockout

```go
// internal/service/auth_security.go

package service

import (
    "context"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
)

type AuthSecurity struct {
    redis *redis.Client
}

func NewAuthSecurity(redis *redis.Client) *AuthSecurity {
    return &AuthSecurity{redis: redis}
}

const (
    MaxFailedAttempts = 5
    LockoutDuration   = 15 * time.Minute
)

func (as *AuthSecurity) RecordFailedLogin(ctx context.Context, email string) error {
    key := fmt.Sprintf("failed_login:%s", email)

    // Increment counter
    count, err := as.redis.Incr(ctx, key).Result()
    if err != nil {
        return err
    }

    // Set expiry on first attempt
    if count == 1 {
        as.redis.Expire(ctx, key, LockoutDuration)
    }

    // Check if account should be locked
    if count >= MaxFailedAttempts {
        lockKey := fmt.Sprintf("account_locked:%s", email)
        as.redis.Set(ctx, lockKey, "1", LockoutDuration)
    }

    return nil
}

func (as *AuthSecurity) IsAccountLocked(ctx context.Context, email string) (bool, time.Duration, error) {
    key := fmt.Sprintf("account_locked:%s", email)

    ttl, err := as.redis.TTL(ctx, key).Result()
    if err != nil {
        return false, 0, err
    }

    if ttl > 0 {
        return true, ttl, nil
    }

    return false, 0, nil
}

func (as *AuthSecurity) ClearFailedAttempts(ctx context.Context, email string) error {
    key := fmt.Sprintf("failed_login:%s", email)
    lockKey := fmt.Sprintf("account_locked:%s", email)

    pipe := as.redis.Pipeline()
    pipe.Del(ctx, key)
    pipe.Del(ctx, lockKey)
    _, err := pipe.Exec(ctx)

    return err
}

func (as *AuthSecurity) GetFailedAttempts(ctx context.Context, email string) (int, error) {
    key := fmt.Sprintf("failed_login:%s", email)
    count, err := as.redis.Get(ctx, key).Int()
    if err == redis.Nil {
        return 0, nil
    }
    return count, err
}
```

**Update login handler:**

```go
// api/v1/auth/auth_handler.go

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()

    var input LoginInput
    if err := web.Decode(w, r, &input); err != nil {
        return web.NewError(err, http.StatusBadRequest)
    }

    // Check if account is locked
    locked, lockDuration, err := h.authSecurity.IsAccountLocked(ctx, input.Email)
    if err != nil {
        h.log.Error("failed to check account lock", "error", err)
    }
    if locked {
        return web.NewError(
            fmt.Errorf("account temporarily locked due to too many failed attempts. Try again in %v",
                lockDuration.Round(time.Minute)),
            http.StatusTooManyRequests,
        )
    }

    // Attempt authentication
    user, err := h.userService.AuthenticateUser(ctx, model.LoginInput{
        Email:    input.Email,
        Password: input.Password,
    })

    if err != nil {
        // Record failed attempt
        if err := h.authSecurity.RecordFailedLogin(ctx, input.Email); err != nil {
            h.log.Error("failed to record failed login", "error", err)
        }

        // Get remaining attempts
        attempts, _ := h.authSecurity.GetFailedAttempts(ctx, input.Email)
        remaining := MaxFailedAttempts - attempts

        if remaining > 0 {
            return web.NewError(
                fmt.Errorf("invalid credentials. %d attempts remaining", remaining),
                http.StatusUnauthorized,
            )
        }

        return h.handleAuthError(err)
    }

    // Clear failed attempts on successful login
    h.authSecurity.ClearFailedAttempts(ctx, input.Email)

    // Rest of login logic...
}
```

---

## A05:2021 - Security Misconfiguration

### 🟡 **MEDIUM**

### What It Is

Insecure default configurations, incomplete setups, open cloud storage, verbose error messages, outdated software.

### 🛠️ Implementation Plan

#### Task 1: Security Headers Middleware

```go
// internal/middlewares/security_headers.go

package middlewares

import (
    "net/http"
    "doit/internal/web"
)

// SecurityHeaders adds security-related HTTP headers
func SecurityHeaders() web.MiddleWare {
    return func(handler web.Handler) web.Handler {
        h := func(w http.ResponseWriter, r *http.Request) error {
            // Prevent clickjacking
            w.Header().Set("X-Frame-Options", "DENY")

            // Prevent MIME type sniffing
            w.Header().Set("X-Content-Type-Options", "nosniff")

            // Enable XSS protection (legacy browsers)
            w.Header().Set("X-XSS-Protection", "1; mode=block")

            // Strict Transport Security (HSTS)
            // Forces HTTPS for 1 year
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

            // Content Security Policy
            w.Header().Set("Content-Security-Policy",
                "default-src 'self'; "+
                "script-src 'self'; "+
                "style-src 'self' 'unsafe-inline'; "+
                "img-src 'self' data: https:; "+
                "font-src 'self'; "+
                "connect-src 'self'; "+
                "frame-ancestors 'none'")

            // Referrer Policy
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

            // Permissions Policy (formerly Feature-Policy)
            w.Header().Set("Permissions-Policy",
                "geolocation=(), "+
                "microphone=(), "+
                "camera=()")

            // Remove server identification
            w.Header().Del("Server")
            w.Header().Del("X-Powered-By")

            return handler(w, r)
        }
        return h
    }
}
```

#### Task 2: Environment-Specific Configuration

```go
// internal/config/config.go

type Environment string

const (
    EnvDevelopment Environment = "development"
    EnvStaging     Environment = "staging"
    EnvProduction  Environment = "production"
)

type Config struct {
    Environment Environment `env:"ENVIRONMENT" envDefault:"development"`
    Debug       bool        `env:"DEBUG" envDefault:"false"`

    Server ServerConfig
    Database DatabaseConfig
    JWT JWTConfig
    Security SecurityConfig
}

type SecurityConfig struct {
    EnableCORS          bool          `env:"ENABLE_CORS" envDefault:"true"`
    AllowedOrigins      []string      `env:"ALLOWED_ORIGINS" envSeparator:","`
    RateLimitEnabled    bool          `env:"RATE_LIMIT_ENABLED" envDefault:"true"`
    MaxRequestsPerMin   int           `env:"MAX_REQUESTS_PER_MIN" envDefault:"100"`
    EnableTLS           bool          `env:"ENABLE_TLS" envDefault:"false"`
    TLSCertFile         string        `env:"TLS_CERT_FILE"`
    TLSKeyFile          string        `env:"TLS_KEY_FILE"`
}

func (c *Config) IsProduction() bool {
    return c.Environment == EnvProduction
}

func (c *Config) IsDevelopment() bool {
    return c.Environment == EnvDevelopment
}
```

#### Task 3: Sanitize Error Messages in Production

```go
// internal/web/error.go

func (e *Error) ToHTTPResponse(isProduction bool) map[string]interface{} {
    response := map[string]interface{}{
        "error": map[string]interface{}{
            "message": e.Message,
            "status":  e.Status,
        },
    }

    if !isProduction {
        // Include detailed error information in development
        response["error"].(map[string]interface{})["details"] = e.Details
        response["error"].(map[string]interface{})["trace"] = e.StackTrace
    }

    return response
}
```

---

## A07:2021 - Identification and Authentication Failures

### 🔴 **CRITICAL**

### What It Is

Failures in authentication mechanisms that allow attackers to compromise passwords, keys, or session tokens.

### ✅ Current Status

**Good:**

- ✅ JWT with refresh tokens
- ✅ Token rotation
- ✅ Password hashing with bcrypt

**Missing:**

- ❌ No multi-factor authentication (MFA)
- ❌ No session management/device tracking
- ❌ No password reset functionality
- ❌ No email verification

### 🛠️ Implementation Plan

#### Task 1: Add Device/Session Tracking

Already mostly implemented in your `refresh_tokens` table! Enhance it:

```sql
-- internal/data/migrations/000005_enhance_device_tracking.up.sql

ALTER TABLE refresh_tokens ADD COLUMN last_used_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE refresh_tokens ADD COLUMN last_ip_address INET;
CREATE INDEX idx_refresh_tokens_last_used ON refresh_tokens(last_used_at);
```

#### Task 2: Implement Email Verification

```sql
-- internal/data/migrations/000006_email_verification.up.sql

ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_verification_token VARCHAR(255);
ALTER TABLE users ADD COLUMN email_verification_expires_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_users_verification_token ON users(email_verification_token);
```

```go
// internal/service/email_verification.go

package service

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/google/uuid"
)

func (s *UserService) GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
    // Generate secure random token
    token := make([]byte, 32)
    if _, err := rand.Read(token); err != nil {
        return "", err
    }
    tokenStr := hex.EncodeToString(token)

    // Store in database with expiry (24 hours)
    err := s.querier.SetEmailVerificationToken(ctx, db.SetEmailVerificationTokenParams{
        ID:                           userID,
        EmailVerificationToken:       &tokenStr,
        EmailVerificationExpiresAt:   time.Now().Add(24 * time.Hour),
    })

    return tokenStr, err
}

func (s *UserService) VerifyEmail(ctx context.Context, token string) error {
    user, err := s.querier.GetUserByVerificationToken(ctx, token)
    if err != nil {
        return fmt.Errorf("invalid or expired verification token")
    }

    // Check expiry
    if user.EmailVerificationExpiresAt.Before(time.Now()) {
        return fmt.Errorf("verification token expired")
    }

    // Mark email as verified
    err = s.querier.VerifyUserEmail(ctx, user.ID)
    if err != nil {
        return fmt.Errorf("failed to verify email: %w", err)
    }

    return nil
}
```

---

## A09:2021 - Security Logging and Monitoring Failures

### 🟡 **MEDIUM**

### What It Is

Insufficient logging and monitoring, which prevents detection of breaches and attacks.

### 🛠️ Implementation Plan

#### Task 1: Comprehensive Audit Logging

```go
// internal/service/audit_log.go

package service

import (
    "context"
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type AuditAction string

const (
    ActionUserLogin          AuditAction = "user.login"
    ActionUserLogout         AuditAction = "user.logout"
    ActionUserRegister       AuditAction = "user.register"
    ActionUserUpdate         AuditAction = "user.update"
    ActionUserDelete         AuditAction = "user.delete"
    ActionTodoCreate         AuditAction = "todo.create"
    ActionTodoUpdate         AuditAction = "todo.update"
    ActionTodoDelete         AuditAction = "todo.delete"
    ActionPasswordChange     AuditAction = "password.change"
    ActionTokenRefresh       AuditAction = "token.refresh"
    ActionUnauthorizedAccess AuditAction = "security.unauthorized_access"
    ActionRateLimitExceeded  AuditAction = "security.rate_limit_exceeded"
)

type AuditLog struct {
    ID          uuid.UUID              `json:"id"`
    UserID      *uuid.UUID             `json:"user_id,omitempty"`
    Action      AuditAction            `json:"action"`
    ResourceType string                `json:"resource_type,omitempty"`
    ResourceID  *uuid.UUID             `json:"resource_id,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    Success     bool                   `json:"success"`
    ErrorMsg    string                 `json:"error_msg,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
}

type AuditLogger struct {
    logger *logger.Logger
    // Could also write to database table for persistence
}

func NewAuditLogger(logger *logger.Logger) *AuditLogger {
    return &AuditLogger{logger: logger}
}

func (al *AuditLogger) Log(ctx context.Context, log AuditLog) {
    log.ID = uuid.New()
    log.Timestamp = time.Now()

    logData, _ := json.Marshal(log)

    al.logger.Info("audit_log",
        "action", log.Action,
        "user_id", log.UserID,
        "success", log.Success,
        "data", string(logData),
    )

    // Optionally: write to database for long-term storage
    // al.writeToDatabase(ctx, log)
}

// Helper methods
func (al *AuditLogger) LogUserAction(ctx context.Context, userID uuid.UUID, action AuditAction, success bool, ipAddress, userAgent string) {
    al.Log(ctx, AuditLog{
        UserID:    &userID,
        Action:    action,
        Success:   success,
        IPAddress: ipAddress,
        UserAgent: userAgent,
    })
}

func (al *AuditLogger) LogSecurityEvent(ctx context.Context, action AuditAction, ipAddress, userAgent, errorMsg string) {
    al.Log(ctx, AuditLog{
        Action:    action,
        Success:   false,
        IPAddress: ipAddress,
        UserAgent: userAgent,
        ErrorMsg:  errorMsg,
    })
}
```

**Apply in handlers:**

```go
// api/v1/auth/auth_handler.go

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()

    // ... authentication logic ...

    if err != nil {
        // Log failed login attempt
        h.auditLogger.LogSecurityEvent(ctx,
            service.ActionUserLogin,
            web.GetClientIP(r),
            web.GetUserAgent(r),
            "invalid credentials",
        )
        return h.handleAuthError(err)
    }

    // Log successful login
    h.auditLogger.LogUserAction(ctx,
        user.ID,
        service.ActionUserLogin,
        true,
        web.GetClientIP(r),
        web.GetUserAgent(r),
    )

    return web.RespondOK(w, r, response)
}
```

#### Task 2: Security Alerting

```go
// internal/service/security_alerts.go

package service

import (
    "context"
    "fmt"
)

type AlertSeverity string

const (
    SeverityLow      AlertSeverity = "low"
    SeverityMedium   AlertSeverity = "medium"
    SeverityHigh     AlertSeverity = "high"
    SeverityCritical AlertSeverity = "critical"
)

type SecurityAlert struct {
    Severity    AlertSeverity
    Title       string
    Description string
    UserID      *uuid.UUID
    IPAddress   string
    Metadata    map[string]interface{}
}

type Alerter interface {
    SendAlert(ctx context.Context, alert SecurityAlert) error
}

// Example: Log-based alerting (basic)
type LogAlerter struct {
    logger *logger.Logger
}

func (la *LogAlerter) SendAlert(ctx context.Context, alert SecurityAlert) error {
    la.logger.Error("SECURITY_ALERT",
        "severity", alert.Severity,
        "title", alert.Title,
        "description", alert.Description,
        "user_id", alert.UserID,
        "ip_address", alert.IPAddress,
    )

    // In production: integrate with PagerDuty, Slack, SNS, etc.
    return nil
}

// Trigger alerts for suspicious activities
func (as *AuthSecurity) CheckSuspiciousActivity(ctx context.Context, userID uuid.UUID, ipAddress string) error {
    // Check for multiple IPs in short time
    // Check for impossible travel (two logins from different continents within minutes)
    // Check for brute force patterns

    // If suspicious:
    alert := SecurityAlert{
        Severity:    SeverityCritical,
        Title:       "Suspicious login activity detected",
        Description: fmt.Sprintf("Multiple failed login attempts from IP %s", ipAddress),
        UserID:      &userID,
        IPAddress:   ipAddress,
    }

    return as.alerter.SendAlert(ctx, alert)
}
```

---

## 📝 Implementation Checklist

### High Priority (Implement First)

- [ ] **A01: Access Control**

  - [ ] Add ownership verification to GetTodoByID
  - [ ] Add ownership verification to UpdateTodo
  - [ ] Implement RBAC (roles table + middleware)
  - [ ] Protect admin endpoints

- [ ] **A02: Cryptographic Failures**

  - [ ] Add password strength validation
  - [ ] Configure TLS/HTTPS
  - [ ] Implement secrets rotation

- [ ] **A03: Injection**

  - [ ] Add input validation for search queries
  - [ ] Validate JSON metadata
  - [ ] Add input sanitization

- [ ] **A07: Authentication**
  - [ ] Implement email verification
  - [ ] Add device/session tracking
  - [ ] Implement password reset

### Medium Priority

- [ ] **A04: Insecure Design**

  - [ ] Implement rate limiting
  - [ ] Add account lockout mechanism

- [ ] **A05: Security Misconfiguration**

  - [ ] Add security headers middleware
  - [ ] Sanitize error messages in production
  - [ ] Environment-specific configs

- [ ] **A09: Logging & Monitoring**
  - [ ] Implement audit logging
  - [ ] Add security alerting

### Low Priority (Optional/Advanced)

- [ ] **A06: Vulnerable Components**

  - [ ] Set up Dependabot
  - [ ] Regular dependency audits

- [ ] **A08: Data Integrity**

  - [ ] Implement webhook signatures
  - [ ] Add checksum validation

- [ ] **A10: SSRF**
  - [ ] Validate external URLs (if you add webhooks)

---

## 🔍 Testing Your Security

### Manual Tests

```bash
# Test 1: Try to access another user's todo
curl -H "Authorization: Bearer <user-a-token>" \
     https://api.example.com/todos/<user-b-todo-id>
# Should return 403 Forbidden

# Test 2: Try SQL injection in search
curl -H "Authorization: Bearer <token>" \
     "https://api.example.com/todos/search?q='; DROP TABLE todos;--"
# Should return 400 Bad Request (validation error)

# Test 3: Test rate limiting
for i in {1..101}; do
  curl https://api.example.com/auth/login
done
# Request 101 should return 429 Too Many Requests

# Test 4: Test account lockout
for i in {1..6}; do
  curl -X POST https://api.example.com/auth/login \
       -d '{"email":"test@example.com","password":"wrong"}'
done
# 6th attempt should return "account locked"
```

### Automated Security Scanning

```bash
# Install OWASP ZAP
docker pull owasp/zap2docker-stable

# Run baseline scan
docker run -t owasp/zap2docker-stable \
    zap-baseline.py -t http://localhost:8080

# Check dependencies for vulnerabilities
go list -json -m all | nancy sleuth
```

---

## 📚 Additional Resources

- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
- [Go Security Best Practices](https://github.com/Checkmarx/Go-SCP)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [REST API Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html)

---

**Last Updated:** October 31, 2025  
**Project:** doit - Go REST API  
**Security Standard:** OWASP Top 10 (2021)
