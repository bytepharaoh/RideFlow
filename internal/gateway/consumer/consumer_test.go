package consumer

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/bytepharoh/rideflow/internal/gateway/ws"
)

type sentMessage struct {
	userID string
	msg    ws.Message
}

type fakeSender struct {
	sent []sentMessage
}

func (f *fakeSender) Send(userID string, msg ws.Message) {
	f.sent = append(f.sent, sentMessage{userID: userID, msg: msg})
}

func TestHandleDriverAssignedSendsToRider(t *testing.T) {
	t.Parallel()

	manager := &fakeSender{}
	consumer := NewEventConsumer(manager, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	body := []byte(`{"trip_id":"trip-1","driver_id":"driver-1","rider_id":"rider-1"}`)
	if err := consumer.handleDriverAssigned(context.Background(), body); err != nil {
		t.Fatalf("handleDriverAssigned() error = %v", err)
	}

	if len(manager.sent) != 1 {
		t.Fatalf("sent messages = %d, want %d", len(manager.sent), 1)
	}
	if manager.sent[0].userID != "rider-1" {
		t.Fatalf("sent to user = %q, want %q", manager.sent[0].userID, "rider-1")
	}
	if manager.sent[0].msg.Type != ws.EventDriverAssigned {
		t.Fatalf("message type = %q, want %q", manager.sent[0].msg.Type, ws.EventDriverAssigned)
	}
}

func TestHandleDriverAssignedIgnoresInvalidJSON(t *testing.T) {
	t.Parallel()

	manager := &fakeSender{}
	consumer := NewEventConsumer(manager, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if err := consumer.handleDriverAssigned(context.Background(), []byte(`not-json`)); err != nil {
		t.Fatalf("handleDriverAssigned() error = %v, want nil", err)
	}
	if len(manager.sent) != 0 {
		t.Fatalf("sent messages = %d, want %d", len(manager.sent), 0)
	}
}

func TestHandleTripCompletedSendsToRiderAndDriver(t *testing.T) {
	t.Parallel()

	manager := &fakeSender{}
	consumer := NewEventConsumer(manager, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	body := []byte(`{"trip_id":"trip-1","rider_id":"rider-1","driver_id":"driver-1","final_fare":42.5}`)
	if err := consumer.handleTripCompleted(context.Background(), body); err != nil {
		t.Fatalf("handleTripCompleted() error = %v", err)
	}

	if len(manager.sent) != 2 {
		t.Fatalf("sent messages = %d, want %d", len(manager.sent), 2)
	}
	if manager.sent[0].msg.Type != ws.EventTripCompleted || manager.sent[1].msg.Type != ws.EventTripCompleted {
		t.Fatal("expected trip completed event for both recipients")
	}
}

func TestHandleTripCompletedIgnoresInvalidJSON(t *testing.T) {
	t.Parallel()

	manager := &fakeSender{}
	consumer := NewEventConsumer(manager, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if err := consumer.handleTripCompleted(context.Background(), []byte(`bad-json`)); err != nil {
		t.Fatalf("handleTripCompleted() error = %v, want nil", err)
	}
	if len(manager.sent) != 0 {
		t.Fatalf("sent messages = %d, want %d", len(manager.sent), 0)
	}
}
