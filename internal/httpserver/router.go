package httpserver

import (
	"net/http"
	"task_scheduler/internal/httpserver/handlers"
)

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", handlers.Health)
}
