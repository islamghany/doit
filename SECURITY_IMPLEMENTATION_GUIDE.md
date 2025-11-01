# Security Implementation Quick Start Guide

> **Step-by-step guide to implement OWASP Top 10 security practices in your doit API**

This guide provides a **prioritized, actionable checklist** to secure your application.

---

## 📋 What's Been Done

I've created the following starter code for you:

✅ **Security Middleware:**

- `internal/middlewares/rbac_middleware.go` - Role-Based Access Control
- `internal/middlewares/security_headers.go` - Security headers (HSTS, CSP, etc.)

✅ **Validation Utilities:**

- `pkg/validator/password.go` - Password strength validation
- `pkg/validator/input.go` - Input sanitization and validation

✅ **Authorization Helpers:**

- `internal/service/authorization.go` - Ownership verification functions

✅ **Model Updates:**

- Added `UserRole` to `internal/model/user.go`
- Added `GetUserFromContext()` helper function

---

## 🚀 Implementation Roadmap

### Phase 1: Critical Security (Week 1) 🔴

#### Step 1: Add Roles to Database (30 minutes)

**Create migration:**

```bash
migrate create -ext sql -dir internal/data/migrations -seq add_user_roles
```

**Edit migration files:**

```sql
-- internal/data/migrations/000004_add_user_roles.up.sql

-- Add role column with default value
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index for role-based queries
CREATE INDEX idx_users_role ON users(role);

-- Add check constraint
ALTER TABLE users ADD CONSTRAINT check_user_role
    CHECK (role IN ('user', 'admin', 'moderator'));

-- Make first user an admin (replace with your email)
UPDATE users SET role = 'admin'
WHERE email = 'your-admin-email@example.com'
LIMIT 1;
```

```sql
-- internal/data/migrations/000004_add_user_roles.down.sql

ALTER TABLE users DROP CONSTRAINT IF EXISTS check_user_role;
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;
```

**Run migration:**

```bash
make migrate-up
# or: migrate -path internal/data/migrations -database "${DATABASE_URL}" up
```

#### Step 2: Update SQLC Queries (15 minutes)

**Update `internal/data/queries/users.sql`:**

```sql
-- name: CreateUser :one
INSERT INTO users (
    id, email, username, password_hash, role, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, COALESCE($5, 'user'), NOW(), NOW()
) RETURNING *;

-- name: GetUserByID :one
SELECT id, email, username, role, email_verified, is_active,
       metadata, last_login_at, created_at, updated_at, token_version
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, username, role, email_verified, is_active,
       metadata, last_login_at, created_at, updated_at, token_version
FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT id, email, username, role, email_verified, is_active,
       metadata, last_login_at, created_at, updated_at, token_version
FROM users
WHERE username = $1 LIMIT 1;

-- Add new query for getting user with password (for authentication)
-- name: GetUserWithPasswordByEmail :one
SELECT id, email, username, password_hash, role, email_verified,
       is_active, metadata, last_login_at, created_at, updated_at, token_version
FROM users
WHERE email = $1 LIMIT 1;
```

**Regenerate SQLC code:**

```bash
make sqlc-generate
# or: sqlc generate
```

#### Step 3: Update User Service (30 minutes)

**Update `internal/service/user_service.go`:**

```go
import (
    "doit/pkg/validator"
)

func (s *UserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // 1. Validate password strength (NEW)
    if err := validator.ValidatePasswordWithDefaults(input.Password); err != nil {
        return nil, fmt.Errorf("password validation failed: %w", err)
    }

    // 2. Validate email format (NEW)
    if err := validator.ValidateEmail(input.Email); err != nil {
        return nil, fmt.Errorf("email validation failed: %w", err)
    }

    // 3. Validate username format (NEW)
    if err := validator.ValidateUsername(input.Username); err != nil {
        return nil, fmt.Errorf("username validation failed: %w", err)
    }

    // ... rest of your existing code ...

    user, err := s.querier.CreateUser(ctx, db.CreateUserParams{
        ID:           uuid.New(),
        Email:        input.Email,
        Username:     input.Username,
        PasswordHash: hashedPassword,
        Role:         db.UserRole(model.UserRoleUser), // NEW: default role
    })

    // ... rest of your code ...
}
```

#### Step 4: Fix Todo Service - Add Ownership Checks (45 minutes)

**Update `internal/service/todo_service.go`:**

```go
// GetTodoByIDWithOwnership - NEW METHOD
func (s *TodoService) GetTodoByIDWithOwnership(ctx context.Context, todoID uuid.UUID, userID uuid.UUID) (*model.Todo, error) {
    todo, err := s.querier.GetTodoByID(ctx, todoID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("todo not found")
        }
        return nil, fmt.Errorf("failed to get todo: %w", err)
    }

    // Verify ownership
    if err := VerifyOwnership(todo.UserID, userID, "todo"); err != nil {
        return nil, err
    }

    return s.toTodoModel(todo), nil
}

// UpdateTodo - UPDATE EXISTING METHOD
func (s *TodoService) UpdateTodo(ctx context.Context, todoID uuid.UUID, userID uuid.UUID, input model.UpdateTodoInput) (*model.Todo, error) {
    // First, verify ownership
    existingTodo, err := s.querier.GetTodoByID(ctx, todoID)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, fmt.Errorf("todo not found")
        }
        return nil, fmt.Errorf("failed to get todo: %w", err)
    }

    if err := VerifyOwnership(existingTodo.UserID, userID, "todo"); err != nil {
        return nil, err
    }

    // ... rest of your existing update logic ...
}
```

#### Step 5: Update Handlers to Use Ownership Checks (30 minutes)

**Update `api/v1/todo/todo_handler.go` (you'll need to create/update this):**

```go
package todo

import (
    "net/http"

    "doit/internal/model"
    "doit/internal/service"
    "doit/internal/web"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type Handler struct {
    service *service.TodoService
}

func NewHandler(service *service.TodoService) *Handler {
    return &Handler{service: service}
}

func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()

    // Get user from context
    user, err := model.GetUserFromContext(ctx)
    if err != nil {
        return web.NewError(err, http.StatusUnauthorized)
    }

    // Parse todo ID
    todoID, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        return web.NewError(err, http.StatusBadRequest)
    }

    // Get todo with ownership verification
    todo, err := h.service.GetTodoByIDWithOwnership(ctx, todoID, user.ID)
    if err != nil {
        return web.NewError(err, http.StatusForbidden)
    }

    return web.RespondOK(w, r, todo)
}

func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()

    // Get user from context
    user, err := model.GetUserFromContext(ctx)
    if err != nil {
        return web.NewError(err, http.StatusUnauthorized)
    }

    // Parse todo ID
    todoID, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        return web.NewError(err, http.StatusBadRequest)
    }

    // Decode request
    var input model.UpdateTodoInput
    if err := web.Decode(w, r, &input); err != nil {
        return web.NewError(err, http.StatusBadRequest)
    }

    // Update todo with ownership verification
    todo, err := h.service.UpdateTodo(ctx, todoID, user.ID, input)
    if err != nil {
        return web.NewError(err, http.StatusForbidden)
    }

    return web.RespondOK(w, r, todo)
}
```

#### Step 6: Apply Security Headers (10 minutes)

**Update `api/server.go`:**

```go
import (
    "doit/internal/middlewares"
)

func (s *Server) setupMiddlewares() {
    // Apply security headers to all routes
    s.router.Use(middlewares.SecurityHeaders())

    // ... your other middlewares ...
}
```

#### Step 7: Test Your Implementation

```bash
# 1. Create two test users
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user1@example.com",
    "username": "user1",
    "password": "SecurePassword123!"
  }'

curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user2@example.com",
    "username": "user2",
    "password": "SecurePassword456!"
  }'

# 2. Login as user1 and create a todo
USER1_TOKEN="<token from login response>"

curl -X POST http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "User 1 Todo",
    "description": "This belongs to user 1"
  }'

# Note the todo ID from response
TODO_ID="<id from response>"

# 3. Try to access user1's todo as user2 (should fail)
USER2_TOKEN="<token from user2 login>"

curl -X GET http://localhost:8080/api/v1/todos/$TODO_ID \
  -H "Authorization: Bearer $USER2_TOKEN"

# Expected: 403 Forbidden

# 4. Test password strength
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "weak"
  }'

# Expected: 400 Bad Request with password validation error
```

---

### Phase 2: Authentication Enhancements (Week 2) 🟡

#### Step 1: Add Rate Limiting

**Install Redis (if not already):**

```bash
# Using Docker
docker run -d -p 6379:6379 redis:alpine

# Or add to docker-compose.yml
```

**Create rate limiter (file already exists in guide, create it):**

See `OWASP_TOP_10_GUIDE.md` → A04 → Rate Limiting implementation

**Apply to routes:**

```go
// api/v1/routes.go

rateLimiter := middlewares.NewRateLimiter(redisClient)

// Strict rate limiting on auth endpoints
r.Group(func(r chi.Router) {
    r.Use(rateLimiter.RateLimitMiddleware(5, 15*time.Minute))
    r.Post("/auth/login", authHandler.Login)
    r.Post("/auth/register", authHandler.Register)
})
```

#### Step 2: Implement Account Lockout

See `OWASP_TOP_10_GUIDE.md` → A04 → Account Lockout implementation

#### Step 3: Add Email Verification

See `OWASP_TOP_10_GUIDE.md` → A07 → Email Verification implementation

---

### Phase 3: Advanced Security (Week 3) 🟢

#### Step 1: Add Audit Logging

See `OWASP_TOP_10_GUIDE.md` → A09 → Audit Logging implementation

#### Step 2: Configure HTTPS/TLS

See `OWASP_TOP_10_GUIDE.md` → A02 → TLS Configuration

#### Step 3: Implement Security Monitoring

See `OWASP_TOP_10_GUIDE.md` → A09 → Security Alerting

---

## 🧪 Security Testing Checklist

### Manual Security Tests

- [ ] **Access Control**

  - [ ] User A cannot access User B's todos
  - [ ] Regular user cannot access admin endpoints
  - [ ] Unauthenticated user gets 401 on protected routes

- [ ] **Password Security**

  - [ ] Weak passwords are rejected
  - [ ] Common passwords are rejected
  - [ ] Passwords are hashed in database

- [ ] **Injection Prevention**

  - [ ] SQL injection attempts are blocked
  - [ ] XSS attempts are sanitized
  - [ ] Search queries validate input

- [ ] **Rate Limiting**

  - [ ] Login attempts are rate limited
  - [ ] API requests are rate limited
  - [ ] Correct headers are returned

- [ ] **Security Headers**
  - [ ] HSTS header is present
  - [ ] CSP header is present
  - [ ] X-Frame-Options is set to DENY

### Automated Security Scanning

```bash
# 1. Run OWASP ZAP scan
docker run -t owasp/zap2docker-stable \
    zap-baseline.py -t http://localhost:8080

# 2. Check dependencies for vulnerabilities
go install github.com/sonatype-nexus-community/nancy@latest
go list -json -m all | nancy sleuth

# 3. Run gosec (Go security checker)
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...

# 4. Check for secrets in code
pip install trufflehog
trufflehog filesystem . --only-verified
```

---

## 📊 Progress Tracking

### Phase 1: Critical Security ✅

- [ ] Database: Add roles column
- [ ] SQLC: Regenerate queries with role
- [ ] Service: Add password validation
- [ ] Service: Add ownership verification to GetTodoByID
- [ ] Service: Add ownership verification to UpdateTodo
- [ ] Handler: Update todo handlers with ownership checks
- [ ] Middleware: Apply security headers
- [ ] Testing: Verify access control works

### Phase 2: Authentication Enhancements

- [ ] Rate limiting on auth endpoints
- [ ] Account lockout after failed attempts
- [ ] Email verification
- [ ] Password reset functionality

### Phase 3: Advanced Security

- [ ] Audit logging
- [ ] HTTPS/TLS configuration
- [ ] Security monitoring and alerting
- [ ] Regular security scanning

---

## 🚨 Common Pitfalls to Avoid

1. **Don't bypass security in development**

   - Keep security checks active in dev environment
   - Use realistic test data

2. **Don't trust client input**

   - Always validate on server side
   - Sanitize all user input

3. **Don't log sensitive data**

   - Never log passwords, tokens, or API keys
   - Use `[REDACTED]` for sensitive fields in logs

4. **Don't use weak secrets**

   - Generate strong JWT secrets (32+ characters)
   - Rotate secrets regularly

5. **Don't ignore errors**
   - Always handle security-related errors
   - Log suspicious activity

---

## 📚 Additional Resources

### Security Learning

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Cheat Sheet](https://github.com/Checkmarx/Go-SCP)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)

### Tools

- [OWASP ZAP](https://www.zaproxy.org/) - Security scanner
- [nancy](https://github.com/sonatype-nexus-community/nancy) - Dependency checker
- [gosec](https://github.com/securego/gosec) - Go security checker
- [TruffleHog](https://github.com/trufflesecurity/trufflehog) - Secret scanner

### Standards

- [NIST SP 800-63B](https://pages.nist.gov/800-63-3/) - Digital Identity Guidelines
- [CWE Top 25](https://cwe.mitre.org/top25/) - Most Dangerous Software Weaknesses
- [ASVS](https://owasp.org/www-project-application-security-verification-standard/) - Application Security Verification Standard

---

## 🎯 Success Criteria

By completing this guide, you will have:

✅ **Critical vulnerabilities fixed**

- Access control on all resources
- Password security enforced
- Input validation and sanitization
- Security headers configured

✅ **Authentication hardened**

- Rate limiting implemented
- Account lockout configured
- Email verification added

✅ **Monitoring in place**

- Audit logging implemented
- Security alerts configured
- Regular security scans automated

✅ **Production-ready security posture**

- HTTPS/TLS enabled
- Secrets properly managed
- Security best practices followed

---

**Next Steps:**

1. Start with Phase 1, Step 1 (Add roles to database)
2. Work through each step systematically
3. Test after each change
4. Document any deviations or issues
5. Update your `LEARNING_ROADMAP.md` as you complete tasks

**Questions?** Refer to the detailed `OWASP_TOP_10_GUIDE.md` for implementation details.

---

**Last Updated:** October 31, 2025  
**Version:** 1.0  
**Project:** doit - Go REST API
