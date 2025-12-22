package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task_scheduler/internal/task"
	"time"
)

type TasksHandler struct {
	svc task.Service
}

func NewTasksHandler(svc task.Service) *TasksHandler {
	return &TasksHandler{
		svc: svc,
	}
}

type createTaskRequest struct {
	Title string  `json:"title"`
	DueAt *string `json:"due_at"`
}

func (h *TasksHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}
	var dueAt *time.Time

	if req.DueAt != nil {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"due_at must be RFC3339",
			)
			return
		}
		dueAt = &t

	}
	tsk, err := h.svc.Create(req.Title, dueAt)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrInvalidInput):
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			// внутренняя ошибка — клиенту детали не показываем
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}
	WriteJSON(w, http.StatusCreated, tsk)
}

func (h *TasksHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}
	tsk, err := h.svc.Get(int(id))
	if err != nil {
		switch {
		case errors.Is(err, task.ErrNotFound):
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, task.ErrInvalidInput):
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			// внутренняя ошибка — клиенту детали не показываем
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, tsk)
}
