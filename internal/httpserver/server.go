package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	s *http.Server
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	registerRoutes(mux)

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
