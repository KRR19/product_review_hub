package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"product_review_hub/internal/config"
	"product_review_hub/internal/rabbitmq"
)

func main() {
	log.Println("Starting Review Watcher service...")

	cfg := config.New()

	// Connect to RabbitMQ
	conn, err := rabbitmq.NewConnection(rabbitmq.Config{
		Host:     cfg.RabbitMQ.Host,
		Port:     cfg.RabbitMQ.Port,
		User:     cfg.RabbitMQ.User,
		Password: cfg.RabbitMQ.Password,
	})
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create consumer
	consumer := rabbitmq.NewConsumer(conn)

	// Start consuming messages
	msgs, err := consumer.Consume()
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Process messages
	go func() {
		for msg := range msgs {
			var event rabbitmq.ReviewEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			// Log the event
			log.Printf("=== Review Event Received ===")
			log.Printf("  Event Type: %s", event.EventType)
			log.Printf("  Timestamp:  %s", event.Timestamp.Format("2006-01-02 15:04:05"))
			log.Printf("  Review ID:  %s", event.Data.ReviewID)
			log.Printf("  Product ID: %s", event.Data.ProductID)
			if event.Data.Rating > 0 {
				log.Printf("  Rating:     %d", event.Data.Rating)
			}
			log.Printf("=============================")
		}
	}()

	log.Println("Review Watcher is running. Press Ctrl+C to exit.")
	<-quit

	log.Println("Shutting down Review Watcher...")
}
