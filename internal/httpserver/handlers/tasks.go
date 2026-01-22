package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"task_scheduler/internal/auth"
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

type updateTaskRequest struct {
	Title  *string         `json:"title,omitempty"`
	DueAt  json.RawMessage `json:"due_at,omitempty"`
	Status *string         `json:"status.omitempty"`
}

//--------------------------------------------------------------//

func (h *TasksHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

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
	tsk, err := h.svc.Create(userID, req.Title, dueAt)
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
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}
	tsk, err := h.svc.Get(userID, int(id))
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

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

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

	tasks, total, effLimit, err := h.svc.List(userID, limit, offset)
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

func (h *TasksHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}

	if req.Title == nil && req.Status == nil && req.DueAt == nil {
		WriteError(w, http.StatusBadRequest, "EMPTY_PATCH", "no fields to update")
		return
	}

	var dueAt task.OptionalTime
	if req.DueAt == nil {
		// поле due_at НЕ пришло вообще => не трогаем
		dueAt.Set = false
	} else if string(req.DueAt) == "null" {
		// поле пришло, но null => очистить due_at
		dueAt.Set = true
		dueAt.Value = nil
	} else {
		var s string
		if err := json.Unmarshal(req.DueAt, &s); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_DUE_AT", "due_at must be RFC3339 string or null")
			return
		}

		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_DUE_AT", "due_at must be RFC3339 format")
			return
		}

		dueAt.Set = true
		dueAt.Value = &t
	}

	input := task.UpdateTaskInput{
		Title:  req.Title,
		Status: req.Status,
		DueAt:  dueAt,
	}

	updated, err := h.svc.Update(userID, id, input)
	if err != nil {
		switch {
		case errors.Is(err, task.ErrNotFound):
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, task.ErrInvalidInput):
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}

	WriteJSON(w, http.StatusOK, updated)

}

func (h *TasksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}

	if err := h.svc.Delete(userID, id); err != nil {
		switch {
		case errors.Is(err, task.ErrNotFound):
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		case errors.Is(err, task.ErrInvalidInput):
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
