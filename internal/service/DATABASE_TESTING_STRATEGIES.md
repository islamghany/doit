# Database Testing Strategies

This guide covers multiple approaches to test your database-backed services, from unit tests to integration tests.

## Architecture Overview

Your current setup:

- **SQLC Generated Code** → `db.Querier` interface
- **Services** → Use `db.Queries` and `database.Pool`
- **Business Logic** → Lives in service layer

## Testing Pyramid

```
                    /\
                   /  \
                  / E2E\    ← Full integration tests
                 /______\
                /        \
               /Integration\  ← Test with real DB
              /____________\
             /              \
            /   Unit Tests   \  ← Mock database layer
           /__________________\
```

---

## Strategy 1: Mock the Querier Interface (Pure Unit Tests)

**Best for:** Testing business logic in isolation

### Pros & Cons

✅ Fast (no database needed)  
✅ Easy to test edge cases  
✅ Reliable (no flaky tests)  
❌ Doesn't test SQL queries  
❌ More setup code

### Implementation

Create mocks for your `db.Querier` interface:

```bash
# Install mockgen
go install go.uber.org/mock/mockgen@latest

# Generate mocks
mockgen -source=internal/data/db/querier.go -destination=internal/service/mocks/mock_querier.go -package=mocks
```

### Example Test

```go
package service_test

import (
    "context"
    "testing"

    "doit/internal/data/db"
    "doit/internal/model"
    "doit/internal/service"
    "doit/internal/service/mocks"

    "github.com/google/uuid"
    "go.uber.org/mock/gomock"
)

func TestUserService_GetUserByID(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockQuerier := mocks.NewMockQuerier(ctrl)

    // Create service with mock
    svc := service.NewUserServiceWithQuerier(mockQuerier)

    // Setup test data
    userID := uuid.New()
    expectedUser := db.User{
        ID:       userID,
        Email:    "test@example.com",
        Username: "testuser",
        IsActive: true,
    }

    // Setup expectations
    mockQuerier.EXPECT().
        GetUserByID(gomock.Any(), userID).
        Return(expectedUser, nil).
        Times(1)

    // Execute
    user, err := svc.GetUserByID(context.Background(), userID)

    // Assert
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if user.ID != userID {
        t.Errorf("expected user ID %v, got %v", userID, user.ID)
    }
}
```

---

## Strategy 2: Test Containers (Real PostgreSQL)

**Best for:** Integration tests with real database behavior

### Pros & Cons

✅ Tests real SQL queries  
✅ Catches database-specific bugs  
✅ Isolated (each test gets fresh DB)  
✅ No manual setup needed  
❌ Slower than unit tests  
❌ Requires Docker

### Setup

```bash
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

### Implementation

```go
package service_test

import (
    "context"
    "testing"
    "time"

    "doit/internal/config"
    "doit/internal/service"
    "doit/pkg/database"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDatabase(t *testing.T) (*database.Pool, func()) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("postgres"),
        postgres.WithPassword("postgres"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second)),
    )
    if err != nil {
        t.Fatal(err)
    }

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatal(err)
    }

    // Connect to database
    pool, err := database.New(ctx, database.Config{
        Host:            "localhost",
        Port:            5432, // testcontainers handles port mapping
        Database:        "testdb",
        User:            "postgres",
        Password:        "postgres",
        MaxConns:        10,
        MinConns:        2,
        MaxConnLifetime: 5 * time.Minute,
        MaxConnIdleTime: 5 * time.Minute,
        DisableTLS:      true,
        LogLevel:        "error",
    })
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations
    // (You'd use golang-migrate here)

    // Return cleanup function
    cleanup := func() {
        pool.Close()
        postgresContainer.Terminate(ctx)
    }

    return pool, cleanup
}

func TestUserService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    pool, cleanup := setupTestDatabase(t)
    defer cleanup()

    svc := service.NewUserService(pool)
    ctx := context.Background()

    // Test creating a user
    input := model.CreateUserInput{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    }

    user, err := svc.CreateUser(ctx, input)
    if err != nil {
        t.Fatalf("failed to create user: %v", err)
    }

    // Test retrieving the user
    retrieved, err := svc.GetUserByID(ctx, user.ID)
    if err != nil {
        t.Fatalf("failed to get user: %v", err)
    }

    if retrieved.Email != input.Email {
        t.Errorf("expected email %s, got %s", input.Email, retrieved.Email)
    }
}
```

---

## Strategy 3: Transaction Rollback Pattern

**Best for:** Fast integration tests using a real DB

### Pros & Cons

✅ Fast (uses one database)  
✅ Tests real SQL  
✅ Automatic cleanup  
❌ Requires shared test database  
❌ Can't test transaction logic

### Implementation

```go
package service_test

import (
    "context"
    "testing"

    "doit/internal/data/db"
    "doit/internal/model"
    "doit/internal/service"
    "doit/pkg/database"

    "github.com/jackc/pgx/v5"
)

type TestDB struct {
    pool *database.Pool
    tx   pgx.Tx
    queries *db.Queries
}

func setupTestTx(t *testing.T) (*TestDB, func()) {
    ctx := context.Background()

    // Connect to test database
    pool, err := database.New(ctx, testDBConfig())
    if err != nil {
        t.Fatal(err)
    }

    // Start transaction
    tx, err := pool.Begin(ctx)
    if err != nil {
        t.Fatal(err)
    }

    // Create queries with transaction
    queries := db.New(pool).WithTx(tx)

    testDB := &TestDB{
        pool:    pool,
        tx:      tx,
        queries: queries,
    }

    // Cleanup: rollback transaction
    cleanup := func() {
        tx.Rollback(ctx)
        pool.Close()
    }

    return testDB, cleanup
}

func TestUserService_WithTxRollback(t *testing.T) {
    testDB, cleanup := setupTestTx(t)
    defer cleanup()

    ctx := context.Background()

    // Create user directly using queries
    user, err := testDB.queries.CreateUser(ctx, db.CreateUserParams{
        ID:           uuid.New(),
        Email:        "test@example.com",
        Username:     "testuser",
        PasswordHash: "hashedpassword",
        Metadata:     []byte("{}"),
    })
    if err != nil {
        t.Fatal(err)
    }

    // Test retrieving user
    retrieved, err := testDB.queries.GetUserByID(ctx, user.ID)
    if err != nil {
        t.Fatal(err)
    }

    if retrieved.Email != user.Email {
        t.Errorf("expected %s, got %s", user.Email, retrieved.Email)
    }

    // Automatic rollback on cleanup - database stays clean
}
```

---

## Strategy 4: Table-Driven Tests with Fixtures

**Best for:** Testing multiple scenarios efficiently

### Implementation

```go
package service_test

import (
    "context"
    "testing"

    "doit/internal/model"
    "doit/internal/service"
)

func TestUserService_CreateUser_TableDriven(t *testing.T) {
    tests := []struct {
        name        string
        input       model.CreateUserInput
        wantErr     bool
        errContains string
    }{
        {
            name: "valid user",
            input: model.CreateUserInput{
                Email:    "valid@example.com",
                Username: "validuser",
                Password: "password123",
            },
            wantErr: false,
        },
        {
            name: "missing email",
            input: model.CreateUserInput{
                Username: "validuser",
                Password: "password123",
            },
            wantErr:     true,
            errContains: "email is required",
        },
        {
            name: "password too short",
            input: model.CreateUserInput{
                Email:    "valid@example.com",
                Username: "validuser",
                Password: "short",
            },
            wantErr:     true,
            errContains: "password must be at least 8 characters",
        },
        {
            name: "duplicate email",
            input: model.CreateUserInput{
                Email:    "duplicate@example.com",
                Username: "user1",
                Password: "password123",
            },
            wantErr:     true,
            errContains: "duplicate key",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            pool, cleanup := setupTestDatabase(t)
            defer cleanup()

            svc := service.NewUserService(pool)
            ctx := context.Background()

            // Execute
            user, err := svc.CreateUser(ctx, tt.input)

            // Assert
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if err != nil && tt.errContains != "" {
                    if !strings.Contains(err.Error(), tt.errContains) {
                        t.Errorf("expected error containing %q, got %q",
                            tt.errContains, err.Error())
                    }
                }
            } else {
                if err != nil {
                    t.Errorf("expected no error, got: %v", err)
                }
                if user == nil {
                    t.Error("expected user, got nil")
                }
            }
        })
    }
}
```

---

## Strategy 5: Test Helpers & Fixtures

**Best for:** Reducing test boilerplate

### Implementation

```go
package testutil

import (
    "context"
    "testing"

    "doit/internal/data/db"
    "doit/internal/model"
    "doit/pkg/database"

    "github.com/google/uuid"
)

// TestHelper provides common test utilities
type TestHelper struct {
    Pool    *database.Pool
    Queries *db.Queries
    ctx     context.Context
}

func NewTestHelper(t *testing.T) (*TestHelper, func()) {
    ctx := context.Background()
    pool, cleanup := setupTestDatabase(t)

    return &TestHelper{
        Pool:    pool,
        Queries: db.New(pool),
        ctx:     ctx,
    }, cleanup
}

// CreateTestUser creates a user for testing
func (h *TestHelper) CreateTestUser(t *testing.T, email string) *model.User {
    user, err := h.Queries.CreateUser(h.ctx, db.CreateUserParams{
        ID:           uuid.New(),
        Email:        email,
        Username:     "test_" + email,
        PasswordHash: "$2a$10$hashedpassword",
        Metadata:     []byte("{}"),
    })
    if err != nil {
        t.Fatalf("failed to create test user: %v", err)
    }

    return &model.User{
        ID:       user.ID,
        Email:    user.Email,
        Username: user.Username,
    }
}

// CreateTestTodo creates a todo for testing
func (h *TestHelper) CreateTestTodo(t *testing.T, userID uuid.UUID, title string) *model.Todo {
    todo, err := h.Queries.CreateTodo(h.ctx, db.CreateTodoParams{
        ID:       uuid.New(),
        UserID:   userID,
        Title:    title,
        Status:   db.TodoStatusPending,
        Priority: db.TodoPriorityMedium,
        Tags:     []string{},
        Metadata: []byte("{}"),
    })
    if err != nil {
        t.Fatalf("failed to create test todo: %v", err)
    }

    return &model.Todo{
        ID:     todo.ID,
        UserID: todo.UserID,
        Title:  todo.Title,
    }
}

// Usage in tests:
func TestTodoService_GetTodoByID(t *testing.T) {
    helper, cleanup := testutil.NewTestHelper(t)
    defer cleanup()

    // Setup test data
    user := helper.CreateTestUser(t, "test@example.com")
    todo := helper.CreateTestTodo(t, user.ID, "Test Todo")

    // Test
    svc := service.NewTodoService(helper.Pool)
    retrieved, err := svc.GetTodoByID(context.Background(), todo.ID)

    // Assert
    if err != nil {
        t.Fatal(err)
    }
    if retrieved.Title != "Test Todo" {
        t.Errorf("expected 'Test Todo', got %s", retrieved.Title)
    }
}
```

---

## Recommended Approach: Hybrid Strategy

Combine strategies for best results:

```
Unit Tests (Mock)           → Fast feedback for business logic
  ├─ Validation logic
  ├─ Data transformation
  └─ Error handling

Integration Tests (TestContainers) → Test SQL & database interactions
  ├─ CRUD operations
  ├─ Transactions
  └─ Complex queries

E2E Tests (Real DB)         → Test full workflows
  └─ Critical user journeys
```

### Project Structure

```
internal/service/
  ├─ user_service.go
  ├─ user_service_test.go          # Unit tests with mocks
  ├─ user_service_integration_test.go  # Integration tests
  ├─ mocks/
  │   └─ mock_querier.go
  └─ testutil/
      ├─ helpers.go
      ├─ fixtures.go
      └─ testcontainers.go
```

---

## Running Tests

```bash
# Run only unit tests (fast)
go test -short ./...

# Run all tests including integration
go test ./...

# Run specific test
go test -run TestUserService_GetUserByID ./internal/service

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

---

## Best Practices

### 1. Use Build Tags

Separate unit and integration tests:

```go
//go:build integration
// +build integration

package service_test

func TestIntegration(t *testing.T) {
    // ...
}
```

Run with: `go test -tags=integration ./...`

### 2. Parallel Tests

```go
func TestUserService_Parallel(t *testing.T) {
    t.Parallel() // Run in parallel

    pool, cleanup := setupTestDatabase(t)
    defer cleanup()

    // Test code...
}
```

### 3. Test Cleanup

Always defer cleanup:

```go
func TestSomething(t *testing.T) {
    pool, cleanup := setupTestDatabase(t)
    defer cleanup() // ← Always defer

    // Test code...
}
```

### 4. Assert Helpers

```go
func assertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper() // Mark as helper function
    if expected != actual {
        t.Errorf("expected %v, got %v", expected, actual)
    }
}
```

### 5. Context Timeout

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Test code with ctx...
}
```

---

## Next Steps

1. Choose your primary strategy (I recommend: Mocks + TestContainers)
2. Set up test infrastructure
3. Write tests for critical paths first
4. Add integration tests for complex queries
5. Aim for 70-80% coverage

Would you like me to implement any of these strategies for your codebase?
