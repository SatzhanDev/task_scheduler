package httpserver

import (
	"net/http"
	"task_scheduler/internal/httpserver/handlers"
	"task_scheduler/internal/task"
)

func registerRoutes(mux *http.ServeMux, svc task.Service) {
	mux.HandleFunc("GET /healthz", handlers.Health)

	h := handlers.NewTasksHandler(svc)

	mux.HandleFunc("POST /v1/tasks", h.Create)
	mux.HandleFunc("GET /v1/tasks/{id}", h.Get)
	mux.HandleFunc("GET /v1/tasks", h.List)

}
