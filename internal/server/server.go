package server

import (
	"context"
	"net/http"
	"product_review_hub/internal/config"
	"product_review_hub/internal/handler"
	"time"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()

	h := handler.New()
	h.RegisterRoutes(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config: cfg,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
