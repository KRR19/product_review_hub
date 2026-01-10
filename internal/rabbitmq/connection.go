package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// ExchangeName is the name of the exchange for review events.
	ExchangeName = "review.events"
	// QueueName is the name of the queue for the review watcher.
	QueueName = "review.events.watcher"
	// RoutingKey is the routing key for review events.
	RoutingKey = "review.#"
)

// Config holds RabbitMQ connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
}

// Connection wraps an AMQP connection.
type Connection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewConnection creates a new RabbitMQ connection.
func NewConnection(cfg Config) (*Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare the exchange
	err = ch.ExchangeDeclare(
		ExchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("Connected to RabbitMQ at %s:%s", cfg.Host, cfg.Port)

	return &Connection{
		conn:    conn,
		channel: ch,
	}, nil
}

// Channel returns the AMQP channel.
func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

// Close closes the connection and channel.
func (c *Connection) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
