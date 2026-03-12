package ws

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	manager *Manager
	logger  *slog.Logger
}

func NewHandler(manager *Manager, logger *slog.Logger) *Handler {
	return &Handler{
		manager: manager,
		logger:  logger,
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	rawConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	conn := newConnection(rawConn, userID, h.logger)
	h.manager.Register(conn)

	go conn.writePump()
	conn.readPump(func() {
		h.manager.Unregister(conn)
	})
}
