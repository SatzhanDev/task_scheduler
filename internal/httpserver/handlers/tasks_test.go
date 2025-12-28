package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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
