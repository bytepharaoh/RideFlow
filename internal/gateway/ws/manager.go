package ws

import (
	"log/slog"
	"sync"
)

type Manager struct {
	connections map[string]map[*Connection]struct{}
	mu          sync.RWMutex
	logger      *slog.Logger
}

func NewManager(logger *slog.Logger) *Manager {
	return &Manager{
		connections: make(map[string]map[*Connection]struct{}),
		logger:      logger,
	}
}

func (m *Manager) Register(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[conn.userID] == nil {
		m.connections[conn.userID] = make(map[*Connection]struct{})
	}

	m.connections[conn.userID][conn] = struct{}{}

	m.logger.Info("websocket client connected",
		"user_id", conn.userID,
		"total_connections", len(m.connections[conn.userID]),
	)
}

func (m *Manager) Unregister(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conns, ok := m.connections[conn.userID]
	if !ok {
		return
	}

	if _, exists := conns[conn]; !exists {
		return
	}

	delete(conns, conn)
	close(conn.send)

	if len(conns) == 0 {
		delete(m.connections, conn.userID)
	}

	m.logger.Info("websocket client disconnected", "user_id", conn.userID)
}

func (m *Manager) Send(userID string, msg Message) {
	m.mu.RLock()
	conns := m.connections[userID]
	m.mu.RUnlock()

	if len(conns) == 0 {
		m.logger.Debug("no websocket connections for user", "user_id", userID)
		return
	}

	for conn := range conns {
		conn.Send(msg)
	}
}

func (m *Manager) ConnectedUsers() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.connections)
}
