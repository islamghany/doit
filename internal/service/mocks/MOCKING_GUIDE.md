# Service Layer Mocking Guide

This guide explains how to use mocks for testing your service layer without a database.

## 🎯 Why Mock?

- ✅ **Fast** - No database setup needed
- ✅ **Isolated** - Test business logic only
- ✅ **Reliable** - No flaky tests
- ✅ **Edge Cases** - Easy to test error scenarios

---

## 🚀 Quick Start

### 1. Generate Mocks

```bash
# Using Makefile (recommended)
make generate-mocks

# Or using script directly
./scripts/generate-mocks.sh

# Or manually
$(go env GOPATH)/bin/mockgen -source=internal/data/db/querier.go \
  -destination=internal/service/mocks/mock_querier.go \
  -package=mocks
```

### 2. Install MockGen (if needed)

```bash
make install-mockgen

# Or directly
go install go.uber.org/mock/mockgen@latest
```

---

## 📚 Basic Usage

### Example Test with Mock

```go
package service

import (
    "context"
    "testing"

    "doit/internal/data/db"
    "doit/internal/service/mocks"

    "github.com/google/uuid"
    "go.uber.org/mock/gomock"
)

func TestUserService_GetUserByID(t *testing.T) {
    // 1. Create controller
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // 2. Create mock
    mockQuerier := mocks.NewMockQuerier(ctrl)

    // 3. Setup test data
    userID := uuid.New()
    expectedUser := db.User{
        ID:       userID,
        Email:    "test@example.com",
        Username: "testuser",
        IsActive: true,
    }

    // 4. Define expectations
    mockQuerier.EXPECT().
        GetUserByID(gomock.Any(), userID).
        Return(expectedUser, nil).
        Times(1)

    // 5. Create service with mock
    // Note: You'll need to refactor UserService to accept Querier interface
    // svc := NewUserServiceWithQuerier(mockQuerier)

    // 6. Run test
    // user, err := svc.GetUserByID(context.Background(), userID)

    // 7. Assert results
    // if err != nil {
    //     t.Fatalf("expected no error, got: %v", err)
    // }
    // if user.ID != userID {
    //     t.Errorf("expected user ID %v, got %v", userID, user.ID)
    // }
}
```

---

## 🔧 Refactoring for Testability

To use mocks effectively, refactor your services to accept the `Querier` interface:

### Before (Hard to Test)

```go
type UserService struct {
    pool    *database.Pool
    queries *db.Queries  // Concrete type
}

func NewUserService(pool *database.Pool) *UserService {
    return &UserService{
        pool:    pool,
        queries: db.New(pool),
    }
}
```

### After (Easy to Test)

```go
type UserService struct {
    pool    *database.Pool
    queries db.Querier  // Interface!
}

func NewUserService(pool *database.Pool) *UserService {
    return &UserService{
        pool:    pool,
        queries: db.New(pool),
    }
}

// Add constructor for testing
func NewUserServiceWithQuerier(querier db.Querier) *UserService {
    return &UserService{
        queries: querier,
    }
}
```

---

## 📖 Common Patterns

### Pattern 1: Simple Method Call

```go
mockQuerier.EXPECT().
    GetUserByID(gomock.Any(), userID).
    Return(expectedUser, nil).
    Times(1)
```

### Pattern 2: Error Handling

```go
mockQuerier.EXPECT().
    GetUserByID(gomock.Any(), userID).
    Return(db.User{}, errors.New("user not found")).
    Times(1)
```

### Pattern 3: Multiple Calls

```go
mockQuerier.EXPECT().
    GetUserByID(gomock.Any(), userID).
    Return(expectedUser, nil).
    Times(3) // Called 3 times
```

### Pattern 4: Any Number of Calls

```go
mockQuerier.EXPECT().
    GetUserByID(gomock.Any(), userID).
    Return(expectedUser, nil).
    AnyTimes()
```

### Pattern 5: Argument Matching

```go
// Match specific argument
mockQuerier.EXPECT().
    CreateUser(gomock.Any(), gomock.Eq(params)).
    Return(user, nil)

// Match any argument
mockQuerier.EXPECT().
    CreateUser(gomock.Any(), gomock.Any()).
    Return(user, nil)

// Custom matcher
mockQuerier.EXPECT().
    CreateUser(gomock.Any(), gomock.AssignableToTypeOf(db.CreateUserParams{})).
    Return(user, nil)
```

### Pattern 6: DoAndReturn (Complex Logic)

```go
mockQuerier.EXPECT().
    CreateUser(gomock.Any(), gomock.Any()).
    DoAndReturn(func(ctx context.Context, params db.CreateUserParams) (db.User, error) {
        // Custom logic
        if params.Email == "" {
            return db.User{}, errors.New("email required")
        }

        return db.User{
            ID:       params.ID,
            Email:    params.Email,
            Username: params.Username,
        }, nil
    })
```

### Pattern 7: Call Order

```go
gomock.InOrder(
    mockQuerier.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil),
    mockQuerier.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Return(user, nil),
)
```

---

## 🧪 Complete Test Examples

### Example 1: GetUserByID Success

```go
func TestUserService_GetUserByID_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockQuerier := mocks.NewMockQuerier(ctrl)
    userID := uuid.New()

    mockQuerier.EXPECT().
        GetUserByID(gomock.Any(), userID).
        Return(db.User{
            ID:       userID,
            Email:    "test@example.com",
            Username: "testuser",
            IsActive: true,
        }, nil)

    svc := NewUserServiceWithQuerier(mockQuerier)
    user, err := svc.GetUserByID(context.Background(), userID)

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("expected email test@example.com, got %s", user.Email)
    }
}
```

### Example 2: GetUserByID Not Found

```go
func TestUserService_GetUserByID_NotFound(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockQuerier := mocks.NewMockQuerier(ctrl)
    userID := uuid.New()

    mockQuerier.EXPECT().
        GetUserByID(gomock.Any(), userID).
        Return(db.User{}, errors.New("no rows"))

    svc := NewUserServiceWithQuerier(mockQuerier)
    user, err := svc.GetUserByID(context.Background(), userID)

    if err == nil {
        t.Error("expected error, got nil")
    }
    if user != nil {
        t.Error("expected nil user on error")
    }
}
```

### Example 3: CreateUser Validation

```go
func TestUserService_CreateUser_Validation(t *testing.T) {
    tests := []struct {
        name      string
        input     model.CreateUserInput
        wantError string
    }{
        {
            name:      "missing email",
            input:     model.CreateUserInput{Username: "test", Password: "password123"},
            wantError: "email is required",
        },
        {
            name:      "password too short",
            input:     model.CreateUserInput{Email: "test@test.com", Username: "test", Password: "short"},
            wantError: "password must be at least 8 characters",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := &UserService{} // No mock needed for validation
            err := svc.validateCreateUserInput(tt.input)

            if err == nil {
                t.Error("expected error, got nil")
            }
            if err.Error() != tt.wantError {
                t.Errorf("expected error %q, got %q", tt.wantError, err.Error())
            }
        })
    }
}
```

### Example 4: Table-Driven Tests

```go
func TestUserService_AuthenticateUser(t *testing.T) {
    tests := []struct {
        name      string
        setupMock func(*mocks.MockQuerier)
        input     model.LoginInput
        wantError bool
    }{
        {
            name: "success",
            setupMock: func(m *mocks.MockQuerier) {
                m.EXPECT().
                    GetUserByEmail(gomock.Any(), "test@test.com").
                    Return(db.User{
                        Email:        "test@test.com",
                        PasswordHash: "$2a$10$hashedpassword",
                        IsActive:     true,
                    }, nil)
                m.EXPECT().
                    UpdateUserLastLogin(gomock.Any(), gomock.Any()).
                    Return(nil)
            },
            input:     model.LoginInput{Email: "test@test.com", Password: "password"},
            wantError: false,
        },
        {
            name: "user not found",
            setupMock: func(m *mocks.MockQuerier) {
                m.EXPECT().
                    GetUserByEmail(gomock.Any(), "notfound@test.com").
                    Return(db.User{}, errors.New("no rows"))
            },
            input:     model.LoginInput{Email: "notfound@test.com", Password: "password"},
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            mockQuerier := mocks.NewMockQuerier(ctrl)
            tt.setupMock(mockQuerier)

            svc := NewUserServiceWithQuerier(mockQuerier)
            user, err := svc.AuthenticateUser(context.Background(), tt.input)

            if tt.wantError && err == nil {
                t.Error("expected error, got nil")
            }
            if !tt.wantError && err != nil {
                t.Errorf("expected no error, got: %v", err)
            }
            if !tt.wantError && user == nil {
                t.Error("expected user, got nil")
            }
        })
    }
}
```

---

## 🎨 Best Practices

### 1. One Mock Per Test

```go
func TestSomething(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish() // ← Always defer

    mock := mocks.NewMockQuerier(ctrl)

    // Test code...
}
```

### 2. Be Specific with Expectations

```go
// ❌ Too vague
mockQuerier.EXPECT().CreateUser(gomock.Any(), gomock.Any())

// ✅ More specific
mockQuerier.EXPECT().
    CreateUser(gomock.Any(), gomock.AssignableToTypeOf(db.CreateUserParams{})).
    DoAndReturn(func(ctx context.Context, params db.CreateUserParams) (db.User, error) {
        // Verify specific fields
        if params.Email != "expected@example.com" {
            t.Error("unexpected email")
        }
        return db.User{ID: params.ID}, nil
    })
```

### 3. Test Business Logic, Not Mocks

```go
// ✅ Good - Tests actual business logic
func TestUserService_CreateUser_ValidationLogic(t *testing.T) {
    svc := &UserService{}
    err := svc.validateCreateUserInput(invalidInput)
    // Assert validation worked correctly
}

// ❌ Bad - Just testing that mock works
func TestUserService_CreateUser_JustCallsMock(t *testing.T) {
    mock.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(user, nil)
    svc.CreateUser(context.Background(), input)
    // This doesn't test any actual logic!
}
```

### 4. Use Helper Functions

```go
func setupUserServiceTest(t *testing.T) (*mocks.MockQuerier, *UserService, func()) {
    ctrl := gomock.NewController(t)
    mockQuerier := mocks.NewMockQuerier(ctrl)
    svc := NewUserServiceWithQuerier(mockQuerier)

    cleanup := func() {
        ctrl.Finish()
    }

    return mockQuerier, svc, cleanup
}

func TestSomething(t *testing.T) {
    mockQuerier, svc, cleanup := setupUserServiceTest(t)
    defer cleanup()

    // Test code...
}
```

---

## 🔍 Debugging Tips

### View Mock Expectations

```go
// Set verbose mode to see all mock calls
ctrl := gomock.NewController(t)
defer ctrl.Finish()

// MockGen will print all expectations and calls
```

### Common Errors

**Error:** "missing call(s) to..."

- **Cause:** Expected method wasn't called
- **Fix:** Verify your service actually calls the mocked method

**Error:** "unexpected call to..."

- **Cause:** Method called but no expectation set
- **Fix:** Add `.EXPECT()` for that method or use `.AnyTimes()`

**Error:** "wrong number of returns"

- **Cause:** Mock return values don't match function signature
- **Fix:** Check the return types in your `.Return()`

---

## 📊 Running Tests

```bash
# Run all service tests
go test ./internal/service/... -v

# Run specific test
go test ./internal/service -v -run TestUserService_GetUserByID

# Run with coverage
go test ./internal/service/... -cover

# Generate coverage report
go test ./internal/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 🚀 Next Steps

1. **Refactor Services** - Add constructor that accepts `db.Querier`
2. **Write Tests** - Start with simple get/create operations
3. **Add Coverage** - Aim for 70-80% coverage
4. **Test Edge Cases** - Error handling, validation, etc.

---

## 📚 Resources

- [gomock Documentation](https://github.com/uber-go/mock)
- [Testing in Go](https://go.dev/doc/tutorial/add-a-test)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)

---

## ✅ Summary

You now have:

- ✅ Mock generator script (`scripts/generate-mocks.sh`)
- ✅ Makefile target (`make generate-mocks`)
- ✅ Generated mocks (`internal/service/mocks/mock_querier.go`)
- ✅ Example tests (`internal/service/user_service_test.go`)
- ✅ Complete documentation

**Start testing your services without a database!** 🎉
