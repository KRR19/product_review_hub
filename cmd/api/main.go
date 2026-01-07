package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product_review_hub/internal/config"
	"product_review_hub/internal/server"
)

func main() {
	cfg := config.New()

	srv := server.New(cfg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
