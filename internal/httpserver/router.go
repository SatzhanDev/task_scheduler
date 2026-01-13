package httpserver

import (
	"net/http"
	"task_scheduler/internal/auth"
	"task_scheduler/internal/httpserver/handlers"
	"task_scheduler/internal/task"
	"task_scheduler/internal/user"
)

func registerRoutes(mux *http.ServeMux, taskSvc task.Service, userSvc user.Service, jwtManager *auth.JWTManager) {
	mux.HandleFunc("GET /healthz", handlers.Health)

	authMW := auth.JWTMiddleware(jwtManager)
	taskHandler := handlers.NewTasksHandler(taskSvc)

	mux.Handle("POST /v1/tasks", authMW(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /v1/tasks/{id}", authMW(http.HandlerFunc(taskHandler.Get)))
	mux.Handle("GET /v1/tasks", authMW(http.HandlerFunc(taskHandler.List)))

	authHandler := handlers.NewAuthHandler(userSvc, jwtManager)
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

}
