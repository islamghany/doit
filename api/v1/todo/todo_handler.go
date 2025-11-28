package todo

import (
	"net/http"

	"doit/internal/model"
	"doit/internal/service"
	"doit/internal/web"
	"doit/pkg/logger"
)

type Handler struct {
	log         *logger.Logger
	todoService *service.TodoService
}

func NewHandler(log *logger.Logger, todoService *service.TodoService) *Handler {
	return &Handler{log: log, todoService: todoService}
}

// CreateTodo godoc
// @Summary      Create a new todo
// @Description  Create a new todo with the given title, description, priority, tags, metadata, and due date
// @Tags         Todo
// @Accept       json
// @Produce      json
// @Param        todo body CreateTodoInput true "Todo to create"
// @Success      201 {object} model.Todo "Todo created successfully"
// @Failure      400 {object} web.Error "Invalid request body"
// @Failure      401 {object} web.Error "Unauthorized"
// @Failure      500 {object} web.Error "Internal server error"
// @Router       /todos [post]
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) error {
	var input CreateTodoInput
	if err := web.Decode(w, r, &input); err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}

	user, err := model.GetUserFromContext(r.Context())
	if err != nil {
		return web.NewError(err, http.StatusUnauthorized)
	}

	todo, err := h.todoService.CreateTodo(r.Context(), model.CreateTodoInput{
		UserID:      user.ID,
		Title:       input.Title,
		Description: input.Description,
		Priority:    model.TodoPriority(input.Priority),
		Tags:        input.Tags,
		Metadata:    input.Metadata,
		DueDate:     input.DueDate,
	})
	if err != nil {
		return web.NewError(err, http.StatusInternalServerError)
	}

	return web.RespondCreated(w, r, todo)
}

// GetTodo godoc
// @Summary      Get a todo by ID
// @Description  Get a todo by ID with the given ID
// @Tags         Todo
// @Accept       json
// @Produce      json
// @Param        id path string true "Todo ID"
// @Success      200 {object} model.Todo "Todo retrieved successfully"
// @Failure      400 {object} web.Error "Invalid request body"
// @Failure      401 {object} web.Error "Unauthorized"
// @Failure      404 {object} web.Error "Todo not found"
// @Failure      500 {object} web.Error "Internal server error"
// @Router       /todos/{id} [get]
func (h *Handler) GetTodo(w http.ResponseWriter, r *http.Request) error {
	todoID, err := web.GetParamUUID(r, "id")
	if err != nil {
		return web.NewError(err, http.StatusBadRequest)
	}
	user, err := model.GetUserFromContext(r.Context())
	if err != nil {
		return web.NewError(err, http.StatusUnauthorized)
	}

	todo, err := h.todoService.GetTodoByID(r.Context(), todoID, user.ID)
	if err != nil {
		return web.NewError(err, http.StatusNotFound)
	}

	return web.RespondOK(w, r, todo)
}

func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) error {
	return nil
}
