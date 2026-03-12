package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher publishes messages to a RabbitMQ topic exchange.
type Publisher struct {
	channel  *amqp.Channel
	exchange string
}

// NewPublisher creates a Publisher and declares the exchange.
// Declaring the exchange is idempotent — safe to call multiple times.
// If the exchange already exists with the same settings, nothing changes.
func NewPublisher(conn *Connection, exchange string) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declare the topic exchange.
	// durable: true  → exchange survives RabbitMQ restarts
	// autoDelete: false → exchange is not deleted when last consumer leaves
	// internal: false → external publishers can use this exchange
	// noWait: false  → wait for server confirmation
	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("declare exchange %q: %w", exchange, err)
	}

	return &Publisher{channel: ch, exchange: exchange}, nil
}

// Publish serializes the payload as JSON and publishes it with the given routing key.
// The routing key determines which queues receive the message.
func (p *Publisher) Publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		p.exchange,
		routingKey,
		false, // mandatory: false → don't return message if no queue matches
		false, // immediate: false → don't require immediate consumer
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // survive RabbitMQ restart
		},
	)
}

// Close closes the publisher's channel.
func (p *Publisher) Close() error {
	return p.channel.Close()
}
