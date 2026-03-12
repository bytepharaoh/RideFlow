package ws

import (
	"io"
	"log/slog"
	"testing"
)

func TestManagerRegisterAndUnregister(t *testing.T) {
	t.Parallel()

	manager := NewManager(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	conn := &Connection{userID: "rider-1", send: make(chan []byte, 1)}

	manager.Register(conn)

	if got := manager.ConnectedUsers(); got != 1 {
		t.Fatalf("ConnectedUsers() = %d, want %d", got, 1)
	}

	manager.Unregister(conn)

	if got := manager.ConnectedUsers(); got != 0 {
		t.Fatalf("ConnectedUsers() = %d, want %d", got, 0)
	}
}

func TestManagerSendFanout(t *testing.T) {
	t.Parallel()

	manager := NewManager(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	connA := &Connection{userID: "rider-1", send: make(chan []byte, 1), logger: manager.logger}
	connB := &Connection{userID: "rider-1", send: make(chan []byte, 1), logger: manager.logger}

	manager.Register(connA)
	manager.Register(connB)

	manager.Send("rider-1", Message{
		Type:    EventDriverAssigned,
		Payload: DriverAssignedPayload{TripID: "trip-1", DriverID: "driver-1"},
	})

	if len(connA.send) != 1 {
		t.Fatalf("connA queued messages = %d, want %d", len(connA.send), 1)
	}
	if len(connB.send) != 1 {
		t.Fatalf("connB queued messages = %d, want %d", len(connB.send), 1)
	}
}

func TestManagerSendMissingUserDoesNothing(t *testing.T) {
	t.Parallel()

	manager := NewManager(slog.New(slog.NewJSONHandler(io.Discard, nil)))
	manager.Send("missing-user", Message{
		Type:    EventTripCompleted,
		Payload: TripCompletedPayload{TripID: "trip-1", FinalFare: 50},
	})

	if got := manager.ConnectedUsers(); got != 0 {
		t.Fatalf("ConnectedUsers() = %d, want %d", got, 0)
	}
}
