package rabbitmq

import (
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connection wraps an AMQP connection with reconnection logic.
type Connection struct {
	conn   *amqp.Connection
	url    string
	logger *slog.Logger
}

// NewConnection establishes a connection to RabbitMQ with retries.
// It will attempt to connect up to maxRetries times before giving up.
func NewConnection(url string, logger *slog.Logger) (*Connection, error) {
	c := &Connection{url: url, logger: logger}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Connection) connect() error {
	const maxRetries = 5
	const retryDelay = 3 * time.Second

	var err error
	for i := range maxRetries {
		c.conn, err = amqp.Dial(c.url)
		if err == nil {
			c.logger.Info("connected to rabbitmq")
			return nil
		}

		c.logger.Warn("rabbitmq connection failed, retrying",
			"attempt", i+1,
			"max", maxRetries,
			"error", err,
		)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("rabbitmq: failed to connect after %d attempts: %w", maxRetries, err)
}

// Channel creates a new AMQP channel.
// Each publisher and consumer should use its own channel.
// Channels are not thread-safe — never share a channel between goroutines.
func (c *Connection) Channel() (*amqp.Channel, error) {
	ch, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: open channel: %w", err)
	}
	return ch, nil
}

// Close closes the underlying AMQP connection.
func (c *Connection) Close() error {
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn.Close()
	}
	return nil
}
