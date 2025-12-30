package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	"task_scheduler/internal/task"
	tasksqlite "task_scheduler/internal/task/sqlite"
)

func newTestService(t *testing.T) task.Service {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "tasks.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, db.Ping())
	require.NoError(t, tasksqlite.Migrate(db))

	repo := tasksqlite.New(db)
	return task.NewService(repo)
}

func TestTasksHandler_Create_OK(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	body := []byte(`{"title":"Task 1"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)

	var got task.Task
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Greater(t, got.ID, 0)
	require.Equal(t, "Task 1", got.Title)
	require.Nil(t, got.DueAt)
	require.Equal(t, task.StatusPending, got.Status)
	require.False(t, got.CreatedAt.IsZero())
	require.False(t, got.UpdatedAt.IsZero())
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	body := []byte(`"title":`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, "INVALID_JSON", er.Error.Code)
}

func TestTasksHandler_Create_EmptyTitle(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	body := []byte(`{"title":""}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))

	// у тебя в handler: ErrInvalidInput -> INVALID_REQUEST
	require.Equal(t, "INVALID_REQUEST", er.Error.Code)
}

func TestTasksHandler_Create_InvalidDueAt(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	//  неправильный формат даты
	body := []byte(`{
		"title": "Task with bad due_at",
		"due_at": "2024-12-99"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.Create(rr, req)

	// 1️⃣ проверяем HTTP-статус
	require.Equal(t, http.StatusBadRequest, rr.Code)

	// 2️⃣ проверяем формат ошибки
	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))

	require.Equal(t, "VALIDATION_ERROR", er.Error.Code)
	require.Equal(t, "due_at must be RFC3339", er.Error.Message)
}

func TestTaskHandler_Get_OK(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	createdTask, err := svc.Create("Task for get", nil)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/"+strconv.Itoa(createdTask.ID), nil)
	rr := httptest.NewRecorder()

	// h.Get(rr, req)
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var got task.Task
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, createdTask.ID, got.ID)
	require.Equal(t, "Task for get", got.Title)
	require.Equal(t, task.StatusPending, got.Status)

}

func TestTasksHandler_Get_InvalidID(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/abc", nil)

	rr := httptest.NewRecorder()

	// h.Get(rr, req)
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))

	require.Equal(t, "INVALID_ID", er.Error.Code)
}

func TestTasksHandler_Get_NotFound(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks/{id}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/99999", nil)
	rr := httptest.NewRecorder()

	// h.Get(rr, req)
	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, "NOT_FOUND", er.Error.Code)
}

func TestTaskHandler_List_Empty(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp listTasksResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	require.Len(t, resp.Data, 0)
	require.Equal(t, 0, resp.Meta.Total)
}

func TestTasksHandler_List_Limit(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	for i := 0; i < 12; i++ {
		_, err := svc.Create("Task "+strconv.Itoa(i+1), nil)
		require.NoError(t, err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?limit=5", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp listTasksResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	require.Len(t, resp.Data, 5)
	require.Equal(t, 12, resp.Meta.Total)
	require.Equal(t, 5, resp.Meta.Limit)
	require.Equal(t, 0, resp.Meta.Offset)
}

func TestTasksHandler_List_Offset(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	for i := 0; i < 12; i++ {
		_, err := svc.Create("Task "+strconv.Itoa(i+1), nil)
		require.NoError(t, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?limit=5&offset=5", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp listTasksResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))

	require.Len(t, resp.Data, 5)
	require.Equal(t, 12, resp.Meta.Total)
	require.Equal(t, 5, resp.Meta.Limit)
	require.Equal(t, 5, resp.Meta.Offset)
}

func TestTasksHandler_List_InvalidLimit(t *testing.T) {
	svc := newTestService(t)
	h := NewTasksHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/tasks", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks?limit=abc", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var er ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, "VALIDATION_ERROR", er.Error.Code)
}
