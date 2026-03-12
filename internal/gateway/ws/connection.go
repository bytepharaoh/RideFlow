package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeTimeout   = 10 * time.Second
	pingInterval   = 30 * time.Second
	maxMessageSize = 512
	readTimeout    = pingInterval + writeTimeout
)

type Connection struct {
	conn   *websocket.Conn
	send   chan []byte
	userID string
	logger *slog.Logger
}

func newConnection(conn *websocket.Conn, userID string, logger *slog.Logger) *Connection {
	return &Connection{
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		logger: logger,
	}
}

func (c *Connection) Send(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		c.logger.Error("failed to marshal websocket message", "error", err)
		return
	}

	select {
	case c.send <- data:
	default:
		c.logger.Warn("websocket send buffer full, dropping message", "user_id", c.userID)
	}
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.logger.Error("websocket write failed", "user_id", c.userID, "error", err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) readPump(onClose func()) {
	defer func() {
		onClose()
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				c.logger.Warn("websocket closed unexpectedly", "user_id", c.userID, "error", err)
			}
			return
		}
	}
}
