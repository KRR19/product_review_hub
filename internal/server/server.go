package server

import (
	"context"
	"log"
	"net/http"
	"product_review_hub/internal/api"
	"product_review_hub/internal/config"
	"product_review_hub/internal/database"
	"product_review_hub/internal/handler"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {
	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	h := handler.New(db)

	api.HandlerFromMux(h, r)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
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
