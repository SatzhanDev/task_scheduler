package httpserver

import (
	"net/http"
	"task_scheduler/internal/httpserver/handlers"
	"task_scheduler/internal/task"
	"task_scheduler/internal/user"
)

func registerRoutes(mux *http.ServeMux, taskSvc task.Service, userSvc user.Service) {
	mux.HandleFunc("GET /healthz", handlers.Health)

	taskH := handlers.NewTasksHandler(taskSvc)
	mux.HandleFunc("POST /v1/tasks", taskH.Create)
	mux.HandleFunc("GET /v1/tasks/{id}", taskH.Get)
	mux.HandleFunc("GET /v1/tasks", taskH.List)

	userH := handlers.NewAuthHandler(userSvc)
	mux.HandleFunc("POST /v1/auth/register", userH.Register)
	mux.HandleFunc("POST /v1/auth/login", userH.Login)

}
