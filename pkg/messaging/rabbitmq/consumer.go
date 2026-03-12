package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HandlerFunc is the function signature for message handlers.
// Return an error to nack (reject) the message.
// Return nil to ack (confirm) the message.
type HandlerFunc func(ctx context.Context, body []byte) error

// Consumer consumes messages from a RabbitMQ queue.
type Consumer struct {
	channel  *amqp.Channel
	exchange string
	logger   *slog.Logger
}

// NewConsumer creates a Consumer and declares the exchange.
func NewConsumer(conn *Connection, exchange string, logger *slog.Logger) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

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

	// Prefetch of 1 means the consumer processes one message at a time.
	// Without this RabbitMQ floods the consumer with all queued messages
	// at once, which can overwhelm the service under load.
	if err := ch.Qos(1, 0, false); err != nil {
		return nil, fmt.Errorf("set qos: %w", err)
	}

	return &Consumer{
		channel:  ch,
		exchange: exchange,
		logger:   logger,
	}, nil
}

// Subscribe binds a queue to the exchange with the given routing key
// and starts consuming messages. Calls handler for each message.
//
// This method blocks — run it in a goroutine.
// It stops when the context is cancelled.
func (c *Consumer) Subscribe(
	ctx context.Context,
	queue, routingKey string,
	handler HandlerFunc,
) error {
	// Declare the queue — idempotent, safe to call on every startup.
	// durable: true → queue survives RabbitMQ restarts
	// autoDelete: false → queue is not deleted when consumer disconnects
	// exclusive: false → other consumers can use this queue
	if _, err := c.channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare queue %q: %w", queue, err)
	}

	// Bind the queue to the exchange with the routing key.
	// Messages published with this routing key will arrive in this queue.
	if err := c.channel.QueueBind(
		queue,
		routingKey,
		c.exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind queue %q: %w", queue, err)
	}

	msgs, err := c.channel.Consume(
		queue,
		"",    // consumer tag — empty = auto generated
		false, // autoAck: false → we manually ack after processing
		false, // exclusive: false → multiple consumers allowed
		false, // noLocal: false
		false, // noWait: false → wait for server confirmation
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", queue, err)
	}

	c.logger.Info("subscribed to queue",
		"queue", queue,
		"routing_key", routingKey,
	)

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed for queue %q", queue)
			}

			if err := handler(ctx, msg.Body); err != nil {
				c.logger.Error("message handler failed",
					"queue", queue,
					"error", err,
				)
				// nack with requeue: true → message goes back to queue for retry
				_ = msg.Nack(false, true)
				continue
			}

			// ack: message processed successfully, remove from queue
			_ = msg.Ack(false)
		}
	}
}

// Close closes the consumer's channel.
func (c *Consumer) Close() error {
	return c.channel.Close()
}

// Decode is a helper to unmarshal a message body into a typed struct.
func Decode[T any](body []byte) (T, error) {
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		return v, fmt.Errorf("decode message: %w", err)
	}
	return v, nil
}
