package httpserver

import (
	"context"
	"net/http"
	"task_scheduler/internal/task"
	"time"
)

type Server struct {
	s *http.Server
}

func New(addr string, svc task.Service) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux, svc)

	return &Server{
		s: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (srv *Server) Start() error {
	return srv.s.ListenAndServe()
}

func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.s.Shutdown(ctx)
}
