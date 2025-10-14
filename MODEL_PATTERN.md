# 📦 Model Pattern with Fat Services

This document explains how we structure our application using the **Model Pattern** alongside **Fat Services** for clean, maintainable code.

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Layer                           │
│                    (api/v1/*/handler.go)                     │
│                   - Parse requests                           │
│                   - Call services                            │
│                   - Return responses                         │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                      Service Layer                           │
│                  (internal/service/*.go)                     │
│              - Business logic (FAT)                          │
│              - Validation                                    │
│              - Transactions                                  │
│              - DB operations                                 │
│              - Returns: model.* types                        │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ├──────────────────┬─────────────────────────┐
                 ▼                  ▼                         ▼
┌────────────────────────┐ ┌──────────────────┐  ┌──────────────────┐
│   Model Layer          │ │  Data Layer      │  │  Database Layer  │
│ (internal/model/*.go)  │ │ (internal/data/) │  │ (pkg/database/)  │
│                        │ │                  │  │                  │
│ - Domain types         │ │ - Generated code │  │ - Connection     │
│ - Input/Output DTOs    │ │ - SQL queries    │  │ - Transactions   │
│ - Business constants   │ │ - DB models      │  │ - Pooling        │
└────────────────────────┘ └──────────────────┘  └──────────────────┘
```

---

## 📁 Directory Structure

```
internal/
├── model/              # Clean domain models (YOU write these)
│   ├── user.go         # User domain types
│   └── todo.go         # Todo domain types
├── service/            # Fat services (YOU write these)
│   ├── user_service.go # User business logic
│   └── todo_service.go # Todo business logic
└── data/
    ├── db/             # Generated code (sqlc generates)
    │   ├── models.go
    │   ├── users.sql.go
    │   └── todos.sql.go
    ├── queries/        # SQL queries (YOU write these)
    │   ├── users.sql
    │   └── todos.sql
    └── migrations/     # Database migrations (YOU write these)
        ├── 000001_create_users_table.up.sql
        └── 000002_create_todos_table.up.sql
```

---

## 🎯 Why Model Pattern?

### ❌ **Without Model Pattern** (Bad)

```go
// Service returns database types directly
func (s *UserService) GetUser(ctx, id) (db.User, error) {
    return s.queries.GetUserByID(ctx, id)
}

// Handler gets database internals
func (h *Handler) GetUser(w, r) error {
    user, _ := h.service.GetUser(ctx, id)
    // user has PasswordHash exposed! 😱
    // user has internal db.* types! 😱
    return web.RespondOK(w, r, user)
}
```

**Problems:**

- ❌ Database types leak to HTTP layer
- ❌ Password hash exposed to API
- ❌ Cannot change DB without breaking API
- ❌ No clear boundaries

### ✅ **With Model Pattern** (Good)

```go
// Service returns domain models
func (s *UserService) GetUser(ctx, id) (*model.User, error) {
    dbUser, _ := s.queries.GetUserByID(ctx, id)
    return s.toUserModel(dbUser), nil // Convert to model
}

// Handler gets clean domain types
func (h *Handler) GetUser(w, r) error {
    user, _ := h.service.GetUser(ctx, id)
    // user is clean model.User ✅
    // No password hash ✅
    // Clean JSON ✅
    return web.RespondOK(w, r, user)
}
```

**Benefits:**

- ✅ Clean separation of concerns
- ✅ Database changes don't affect API
- ✅ Sensitive data hidden (password hashes)
- ✅ Clear domain boundaries
- ✅ Easy to test

---

## 📝 Model Types Explained

### 1. **Domain Models** (Output Types)

These represent your business entities and are returned from services.

```go
// internal/model/user.go
type User struct {
    ID            uuid.UUID              `json:"id"`
    Email         string                 `json:"email"`
    Username      string                 `json:"username"`
    EmailVerified bool                   `json:"email_verified"`
    IsActive      bool                   `json:"is_active"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
    LastLoginAt   *time.Time             `json:"last_login_at,omitempty"`
    CreatedAt     time.Time              `json:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at"`
    // ✅ NO PasswordHash - sensitive data hidden
    // ✅ Clean JSON tags
    // ✅ Friendly types (map instead of []byte)
}
```

### 2. **Input Types** (DTOs)

These represent data coming into your system.

```go
type CreateUserInput struct {
    Email    string                 `json:"email" validate:"required,email"`
    Username string                 `json:"username" validate:"required,min=3,max=50"`
    Password string                 `json:"password" validate:"required,min=8"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

**Benefits:**

- Validation tags
- Only fields needed for creation
- Clear intent

### 3. **Update Types** (Partial DTOs)

These represent partial updates using pointers for optional fields.

```go
type UpdateUserInput struct {
    Email    *string                `json:"email,omitempty" validate:"omitempty,email"`
    Username *string                `json:"username,omitempty" validate:"omitempty,min=3"`
    IsActive *bool                  `json:"is_active,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

**Why pointers?**

- `nil` = field not provided (don't update)
- `&value` = field provided with value (update it)
- Can distinguish between "not sent" and "sent as empty"

### 4. **Filter Types**

These represent query parameters for filtering/pagination.

```go
type TodoFilter struct {
    UserID    uuid.UUID     `json:"user_id" validate:"required"`
    Status    *TodoStatus   `json:"status,omitempty"`
    Priority  *TodoPriority `json:"priority,omitempty"`
    Tags      []string      `json:"tags,omitempty"`
    DueBefore *time.Time    `json:"due_before,omitempty"`
    Limit     int32         `json:"limit" validate:"min=1,max=100"`
    Offset    int32         `json:"offset" validate:"min=0"`
}
```

### 5. **Aggregate Types**

These represent computed/aggregated data.

```go
type TodoStats struct {
    Total      int64 `json:"total"`
    Pending    int64 `json:"pending"`
    InProgress int64 `json:"in_progress"`
    Completed  int64 `json:"completed"`
    Overdue    int64 `json:"overdue"`
}
```

---

## 🔄 Data Flow Example

### Creating a User

```
┌──────────────┐
│ HTTP Request │ POST /api/v1/users
└──────┬───────┘
       │ {
       │   "email": "user@example.com",
       │   "username": "john",
       │   "password": "secret123"
       │ }
       ▼
┌──────────────────────────────────────────────┐
│ Handler (THIN)                               │
│ api/v1/user/handler.go                       │
├──────────────────────────────────────────────┤
│ 1. Parse JSON → model.CreateUserInput        │
│ 2. Validate input                            │
│ 3. Call service                              │
└──────┬───────────────────────────────────────┘
       │ model.CreateUserInput
       ▼
┌──────────────────────────────────────────────┐
│ Service (FAT)                                │
│ internal/service/user_service.go             │
├──────────────────────────────────────────────┤
│ 1. Business validation                       │
│ 2. Hash password                             │
│ 3. Convert to db.CreateUserParams            │
│ 4. Call queries.CreateUser()                 │
│ 5. Convert db.User → model.User              │
└──────┬───────────────────────────────────────┘
       │ db.User
       ▼
┌──────────────────────────────────────────────┐
│ Data Layer (GENERATED)                       │
│ internal/data/db/users.sql.go                │
├──────────────────────────────────────────────┤
│ 1. Execute SQL INSERT                        │
│ 2. Return db.User                            │
└──────┬───────────────────────────────────────┘
       │ db.User {
       │   ID, Email, Username,
       │   PasswordHash, ...
       │ }
       ▼
┌──────────────────────────────────────────────┐
│ Service Conversion                           │
│ toUserModel(db.User) → model.User            │
├──────────────────────────────────────────────┤
│ • Hide PasswordHash                          │
│ • Convert []byte → map[string]interface{}    │
│ • Handle nullable fields                     │
└──────┬───────────────────────────────────────┘
       │ model.User (clean)
       ▼
┌──────────────────────────────────────────────┐
│ Handler Response                             │
├──────────────────────────────────────────────┤
│ web.RespondCreated(w, r, user)               │
└──────┬───────────────────────────────────────┘
       │ JSON Response
       ▼
{
  "id": "...",
  "email": "user@example.com",
  "username": "john",
  "email_verified": false,
  "is_active": true,
  "created_at": "2025-10-14T..."
}
```

---

## 🔨 Implementation Examples

### Complete Service Example

```go
// internal/service/user_service.go
package service

type UserService struct {
    pool    *database.Pool
    queries *db.Queries
}

func NewUserService(pool *database.Pool) *UserService {
    return &UserService{
        pool:    pool,
        queries: db.New(pool),
    }
}

// Public method: Takes model.* input, returns model.* output
func (s *UserService) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // 1. Validate
    if err := s.validateCreateUserInput(input); err != nil {
        return nil, err
    }

    // 2. Business logic
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    metadataJSON, _ := json.Marshal(input.Metadata)

    // 3. Call database (db.* types)
    dbUser, err := s.queries.CreateUser(ctx, db.CreateUserParams{
        ID:           uuid.New(),
        Email:        input.Email,
        Username:     input.Username,
        PasswordHash: string(hashedPassword),
        Metadata:     metadataJSON,
    })
    if err != nil {
        return nil, err
    }

    // 4. Convert to model (hide internals)
    return s.toUserModel(dbUser), nil
}

// Private converter: db.User → model.User
func (s *UserService) toUserModel(dbUser db.User) *model.User {
    user := &model.User{
        ID:            dbUser.ID,
        Email:         dbUser.Email,
        Username:      dbUser.Username,
        EmailVerified: dbUser.EmailVerified,
        IsActive:      dbUser.IsActive,
        CreatedAt:     dbUser.CreatedAt,
        UpdatedAt:     dbUser.UpdatedAt,
    }

    // Handle nullable fields
    if dbUser.LastLoginAt.Valid {
        user.LastLoginAt = &dbUser.LastLoginAt.Time
    }

    // Convert JSONB from []byte to map
    if len(dbUser.Metadata) > 0 {
        var metadata map[string]interface{}
        if err := json.Unmarshal(dbUser.Metadata, &metadata); err == nil {
            user.Metadata = metadata
        }
    }

    return user
}
```

### Complete Handler Example

```go
// api/v1/user/handler.go
package user

type Handler struct {
    log     *logger.Logger
    service *service.UserService  // Uses model.* types
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) error {
    ctx := r.Context()

    // 1. Parse to model.CreateUserInput
    var input model.CreateUserInput
    if err := web.DecodeJSON(r, &input); err != nil {
        return web.NewRequestError(err, http.StatusBadRequest)
    }

    // 2. Call service (gets model.User back)
    user, err := h.service.CreateUser(ctx, input)
    if err != nil {
        h.log.Error(ctx, "Failed to create user", "error", err)
        return err
    }

    // 3. Return response (model.User → JSON)
    return web.RespondCreated(w, r, user)
}
```

---

## 🎨 Best Practices

### ✅ DO

1. **Keep models in `internal/model/`**

   ```
   internal/model/
   ├── user.go      # User domain types
   ├── todo.go      # Todo domain types
   └── common.go    # Shared types (if needed)
   ```

2. **Use clear naming conventions**

   ```go
   type User struct {}              // Domain model (output)
   type CreateUserInput struct {}   // Create operation
   type UpdateUserInput struct {}   // Update operation
   type UserFilter struct {}        // List/query operations
   type UserStats struct {}         // Aggregates
   ```

3. **Hide sensitive data**

   ```go
   // ❌ BAD - exposes password
   type User struct {
       PasswordHash string
   }

   // ✅ GOOD - no password
   type User struct {
       Email string
       Username string
   }
   ```

4. **Use pointers for optional updates**

   ```go
   type UpdateUserInput struct {
       Email    *string  // nil = don't update
       Username *string  // &"value" = update to value
   }
   ```

5. **Convert at service boundary**
   ```go
   // Service converts: db.User → model.User
   func (s *Service) GetUser(ctx, id) (*model.User, error) {
       dbUser, _ := s.queries.GetUserByID(ctx, id)
       return s.toUserModel(dbUser), nil  // ✅ Convert
   }
   ```

### ❌ DON'T

1. **Don't return db.\* types from services**

   ```go
   // ❌ BAD - leaks database internals
   func (s *Service) GetUser(ctx, id) (db.User, error)

   // ✅ GOOD - returns domain model
   func (s *Service) GetUser(ctx, id) (*model.User, error)
   ```

2. **Don't use db.\* types in handlers**

   ```go
   // ❌ BAD
   func (h *Handler) CreateUser(w, r) error {
       var params db.CreateUserParams  // Database type in handler!
   }

   // ✅ GOOD
   func (h *Handler) CreateUser(w, r) error {
       var input model.CreateUserInput  // Domain type
   }
   ```

3. **Don't expose internal IDs or keys**

   ```go
   // ❌ BAD
   type User struct {
       InternalKey string `json:"internal_key"`
   }

   // ✅ GOOD
   type User struct {
       ID uuid.UUID `json:"id"`
   }
   ```

---

## 🧪 Testing with Models

```go
func TestUserService_CreateUser(t *testing.T) {
    ctx := context.Background()
    pool := setupTestDB(t)
    service := service.NewUserService(pool)

    // Use model types in tests
    input := model.CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    }

    user, err := service.CreateUser(ctx, input)

    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
    assert.NotEmpty(t, user.ID)
    // ✅ No PasswordHash field to accidentally check
}
```

---

## 📊 Comparison Table

| Aspect          | Without Models           | With Models           |
| --------------- | ------------------------ | --------------------- |
| **Type Safety** | db.User everywhere       | model.User in domain  |
| **Security**    | Password hash exposed    | Sensitive data hidden |
| **Testing**     | Database types in tests  | Clean domain types    |
| **Flexibility** | DB changes = API changes | DB independent of API |
| **Clarity**     | Mixed concerns           | Clear boundaries      |
| **Maintenance** | Hard to refactor         | Easy to change        |

---

## 🚀 Quick Start

1. **Define your model**

   ```go
   // internal/model/entity.go
   type Entity struct {
       ID   uuid.UUID `json:"id"`
       Name string    `json:"name"`
   }

   type CreateEntityInput struct {
       Name string `json:"name" validate:"required"`
   }
   ```

2. **Use in service**

   ```go
   func (s *Service) CreateEntity(ctx, input model.CreateEntityInput) (*model.Entity, error) {
       dbEntity, _ := s.queries.CreateEntity(ctx, db.CreateEntityParams{...})
       return s.toEntityModel(dbEntity), nil
   }
   ```

3. **Use in handler**
   ```go
   func (h *Handler) CreateEntity(w, r) error {
       var input model.CreateEntityInput
       web.DecodeJSON(r, &input)
       entity, _ := h.service.CreateEntity(ctx, input)
       return web.RespondCreated(w, r, entity)
   }
   ```

---

## 📚 Summary

The **Model Pattern** provides:

- ✅ **Clean separation** between database and domain
- ✅ **Security** by hiding sensitive data
- ✅ **Flexibility** to change DB without breaking API
- ✅ **Testability** with clear, simple types
- ✅ **Maintainability** through clear boundaries

Combined with **Fat Services**, you get:

- All business logic in one place
- Clean, reusable domain models
- Easy to test and maintain
- Production-ready architecture

---

**Happy modeling! 🎉**
