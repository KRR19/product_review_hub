package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"product_review_hub/internal/api"
	"product_review_hub/internal/config"
	"product_review_hub/internal/database"
	"product_review_hub/internal/handler"
	idempotencymw "product_review_hub/internal/middleware"
	"product_review_hub/internal/rabbitmq"
	"product_review_hub/internal/redis"
	"product_review_hub/internal/repository/idempotency"
	"product_review_hub/internal/repository/products"
	"product_review_hub/internal/repository/reviews"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	rabbitConn *rabbitmq.Connection
}

const idempotencyTTL = time.Minute

func New(cfg *config.Config) *Server {
	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	redisClient, err := redis.New(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	// Initialize idempotency store
	idempotencyStore := idempotency.NewRedisStore(redisClient)

	// Initialize RabbitMQ
	rabbitConn, err := rabbitmq.NewConnection(rabbitmq.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
	})
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}

	// Initialize publisher
	publisher := rabbitmq.NewPublisher(rabbitConn)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(idempotencymw.Idempotency(idempotencyStore, idempotencyTTL))

	// Initialize repositories
	productRepo := products.NewRepository(db)
	reviewRepo := reviews.NewRepository(db)

	h := handler.New(db, productRepo, reviewRepo, publisher)

	api.HandlerFromMux(h, r)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		config:     cfg,
		rabbitConn: rabbitConn,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.rabbitConn != nil {
		if err := s.rabbitConn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
	return s.httpServer.Shutdown(ctx)
}
