package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consumes messages from a RabbitMQ queue.
type Consumer struct {
	conn *Connection
}

// NewConsumer creates a new Consumer.
func NewConsumer(conn *Connection) *Consumer {
	return &Consumer{conn: conn}
}

// Consume starts consuming messages from the queue and returns a channel of deliveries.
func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	ch := c.conn.Channel()

	// Declare the queue
	q, err := ch.QueueDeclare(
		QueueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,       // queue name
		RoutingKey,   // routing key
		ExchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Consumer started, waiting for messages on queue: %s", QueueName)

	return msgs, nil
}
