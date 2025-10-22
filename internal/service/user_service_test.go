package service

import (
	"context"
	"errors"
	"testing"

	"doit/internal/data/db"
	"doit/internal/model"
	"doit/internal/service/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

// TestUserService_GetUserByID tests GetUserByID with mock
func TestUserService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	// Setup test data
	userID := uuid.New()
	expectedUser := db.User{
		ID:            userID,
		Email:         "test@example.com",
		Username:      "testuser",
		PasswordHash:  "hashedpassword",
		EmailVerified: true,
		IsActive:      true,
		Metadata:      []byte("{}"),
	}

	// Setup expectation
	mockQuerier.EXPECT().
		GetUserByID(gomock.Any(), userID).
		Return(expectedUser, nil).
		Times(1)

	// Create service with mock (we'll need to refactor UserService to accept Querier interface)
	// For now, this demonstrates the pattern

	// Test would continue like:
	svc := NewUserServiceWithQuerier(mockQuerier)
	user, err := svc.GetUserByID(context.Background(), userID)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID != userID {
		t.Errorf("expected user ID %v, got %v", userID, user.ID)
	}
}

// TestUserService_GetUserByID_NotFound tests error handling
func TestUserService_GetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	userID := uuid.New()

	// Setup expectation for not found
	mockQuerier.EXPECT().
		GetUserByID(gomock.Any(), userID).
		Return(db.User{}, errors.New("no rows")).
		Times(1)

	// Test implementation would go here
	// This shows the pattern for testing error cases

	svc := NewUserServiceWithQuerier(mockQuerier)
	user, err := svc.GetUserByID(context.Background(), userID)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if user != nil {
		t.Error("expected nil user on error")
	}
}

// TestUserService_CreateUser tests user creation with validation
func TestUserService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	input := model.CreateUserInput{
		Email:    "newuser@example.com",
		Username: "newuser",
		Password: "password123",
	}

	// Setup expectation
	mockQuerier.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateUserParams) (db.User, error) {
			// Verify the params
			if params.Email != input.Email {
				t.Errorf("expected email %s, got %s", input.Email, params.Email)
			}
			if params.Username != input.Username {
				t.Errorf("expected username %s, got %s", input.Username, params.Username)
			}

			return db.User{
				ID:            params.ID,
				Email:         params.Email,
				Username:      params.Username,
				PasswordHash:  params.PasswordHash,
				EmailVerified: false,
				IsActive:      true,
				Metadata:      params.Metadata,
			}, nil
		}).
		Times(1)

	// Test implementation would use the mock
	svc := NewUserServiceWithQuerier(mockQuerier)
	user, err := svc.CreateUser(context.Background(), input)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.Email != input.Email {
		t.Errorf("expected email %s, got %s", input.Email, user.Email)
	}
	if user.Username != input.Username {
		t.Errorf("expected username %s, got %s", input.Username, user.Username)
	}

	if user.EmailVerified != false {
		t.Errorf("expected email verified %t, got %t", false, user.EmailVerified)
	}
	if user.IsActive != true {
		t.Errorf("expected is active %t, got %t", true, user.IsActive)
	}
}

// TestUserService_CreateUser_ValidationError tests validation
func TestUserService_CreateUser_ValidationError(t *testing.T) {
	tests := []struct {
		name      string
		input     model.CreateUserInput
		wantError string
	}{
		{
			name: "missing email",
			input: model.CreateUserInput{
				Username: "testuser",
				Password: "password123",
			},
			wantError: "email is required",
		},
		{
			name: "missing username",
			input: model.CreateUserInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantError: "username is required",
		},
		{
			name: "password too short",
			input: model.CreateUserInput{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "short",
			},
			wantError: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test doesn't need a mock because it tests pure validation logic
			// You can test the validateCreateUserInput method directly
			svc := NewUserServiceWithQuerier(nil)
			err := svc.validateCreateUserInput(tt.input)

			if err != nil && err.Error() != tt.wantError {
				t.Errorf("expected error %q, got %q", tt.wantError, err.Error())
			}
			if err != nil && err.Error() != tt.wantError {
				t.Errorf("expected error %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}
