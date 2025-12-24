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

//---------------------------------------------------------------//

type listTasksMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type listTasksResponse struct {
	Data []task.Task   `json:"data"`
	Meta listTasksMeta `json:"meta"`
}

//---------------------------------------------------------------//

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

func (h *TasksHandler) List(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	limit := 0
	offset := 0

	if limitStr := q.Get("limit"); limitStr != "" {
		numLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"limit must be a number",
			)
			return
		}
		limit = numLimit
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		numOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"offset must be a number",
			)
			return
		}
		offset = numOffset
	}

	tasks, total, effLimit, err := h.svc.List(limit, offset)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, listTasksResponse{
		Data: tasks,
		Meta: listTasksMeta{
			Total:  total,
			Limit:  effLimit,
			Offset: offset,
		},
	})
}
