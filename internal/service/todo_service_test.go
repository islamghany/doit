package service

import (
	"context"
	"encoding/json"
	"testing"

	"doit/internal/data/db"
	"doit/internal/model"
	"doit/internal/service/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	// testify requires this
	"github.com/stretchr/testify/require"
)

// TestTodoService_CreateTodo tests CreateTodo with mock
func TestTodoService_CreateTodo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	// Setup test data
	userID := uuid.New()
	todoID := uuid.New()

	input := model.CreateTodoInput{
		UserID:      userID,
		Title:       "Test Todo",
		Description: "This is a test todo",
		Priority:    model.TodoPriorityMedium,
		Tags:        []string{"test", "todo"},
		Metadata:    map[string]interface{}{"test": "test"},
	}

	// Setup expectation
	mockQuerier.EXPECT().
		CreateTodo(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateTodoParams) (db.Todo, error) {
			// Verify the params
			if params.UserID != userID {
				t.Errorf("expected user ID %v, got %v", userID, params.UserID)
			}
			return db.Todo{
				ID: todoID,
			}, nil
		}).
		Times(1)

	// Test implementation would use the mock
	svc := NewTodoServiceWithQuerier(mockQuerier)
	todo, err := svc.CreateTodo(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, todo)
	require.Equal(t, todoID, todo.ID)
}

// TestTodoService_GetTodoByID tests GetTodoByID with mock
func TestTodoService_GetTodoByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	description := "This is a test todo"
	// Setup test data
	todoID := uuid.New()
	userID := uuid.New()
	metadata := map[string]interface{}{"test": "test"}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}
	expectedTodo := db.Todo{
		ID:          todoID,
		UserID:      userID,
		Title:       "Test Todo",
		Description: &description,
		Priority:    db.TodoPriorityMedium,
		Tags:        []string{"test", "todo"},
		Metadata:    metadataJSON,
	}

	// Setup expectation
	mockQuerier.EXPECT().
		GetTodoByID(gomock.Any(), todoID).
		Return(expectedTodo, nil).
		Times(1)

	// Test implementation would use the mock
	svc := NewTodoServiceWithQuerier(mockQuerier)
	todo, err := svc.GetTodoByID(context.Background(), todoID, userID)

	require.NoError(t, err)
	require.NotNil(t, todo)
	require.Equal(t, userID, todo.UserID)
	require.Equal(t, todoID, todo.ID)
	require.Equal(t, expectedTodo.Title, todo.Title)
	require.Equal(t, *expectedTodo.Description, todo.Description)
	require.Equal(t, string(expectedTodo.Priority), string(todo.Priority))
	require.Equal(t, expectedTodo.Tags, todo.Tags)
	require.Equal(t, metadata, todo.Metadata)
}

// TestTodoService_ListUserTodos tests ListUserTodos with mock
func TestTodoService_ListUserTodos(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockQuerier := mocks.NewMockQuerier(ctrl)

	// Setup test data
	userID := uuid.New()
	limit := int32(10)
	offset := int32(0)
	description := "This is a test todo"
	metadata := map[string]interface{}{"test": "test"}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}
	expectedTodos := []db.Todo{
		{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       "Test Todo",
			Description: &description,
			Priority:    db.TodoPriorityMedium,
			Tags:        []string{"test", "todo"},
			Metadata:    metadataJSON,
		},
		{
			ID:          uuid.New(),
			UserID:      userID,
			Title:       "Test Todo 2",
			Description: &description,
			Priority:    db.TodoPriorityMedium,
			Tags:        []string{"test", "todo"},
			Metadata:    metadataJSON,
		},
	}

	// Setup expectation
	mockQuerier.EXPECT().
		ListTodosByUser(gomock.Any(), db.ListTodosByUserParams{
			UserID: userID,
			Limit:  limit,
			Offset: offset,
		}).
		Return(expectedTodos, nil).
		Times(1)

	// Test implementation would use the mock
	svc := NewTodoServiceWithQuerier(mockQuerier)
	todos, err := svc.ListUserTodos(context.Background(), userID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, todos)
	require.Equal(t, len(expectedTodos), len(todos))
	for i, todo := range todos {
		require.Equal(t, expectedTodos[i].ID, todo.ID)
		require.Equal(t, expectedTodos[i].UserID, todo.UserID)
		require.Equal(t, expectedTodos[i].Title, todo.Title)
		require.Equal(t, *expectedTodos[i].Description, todo.Description)
		require.Equal(t, string(expectedTodos[i].Priority), string(todo.Priority))
		require.Equal(t, expectedTodos[i].Tags, todo.Tags)
		require.Equal(t, metadata, todo.Metadata)
	}
}
