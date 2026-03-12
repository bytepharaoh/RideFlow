package consumer

import (
	"context"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/gateway/ws"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

type sender interface {
	Send(userID string, msg ws.Message)
}

type EventConsumer struct {
	manager  sender
	consumer *rabbitmq.Consumer
	logger   *slog.Logger
}

func NewEventConsumer(manager sender, consumer *rabbitmq.Consumer, logger *slog.Logger) *EventConsumer {
	return &EventConsumer{
		manager:  manager,
		consumer: consumer,
		logger:   logger,
	}
}

func (c *EventConsumer) Start(ctx context.Context) {
	go c.subscribeDriverAssigned(ctx)
	go c.subscribeTripCompleted(ctx)

	<-ctx.Done()
}

func (c *EventConsumer) subscribeDriverAssigned(ctx context.Context) {
	if err := c.consumer.Subscribe(
		ctx,
		events.QueueGatewayDriverAssigned,
		events.RoutingKeyDriverAssigned,
		c.handleDriverAssigned,
	); err != nil {
		c.logger.Error("driver assigned consumer stopped", "error", err)
	}
}

func (c *EventConsumer) subscribeTripCompleted(ctx context.Context) {
	if err := c.consumer.Subscribe(
		ctx,
		events.QueueGatewayTripCompleted,
		events.RoutingKeyTripCompleted,
		c.handleTripCompleted,
	); err != nil {
		c.logger.Error("trip completed consumer stopped", "error", err)
	}
}

func (c *EventConsumer) handleDriverAssigned(ctx context.Context, body []byte) error {
	event, err := rabbitmq.Decode[events.DriverAssigned](body)
	if err != nil {
		c.logger.Error("invalid driver assigned event", "error", err)
		return nil
	}

	c.manager.Send(event.RiderID, ws.Message{
		Type: ws.EventDriverAssigned,
		Payload: ws.DriverAssignedPayload{
			TripID:     event.TripID,
			DriverID:   event.DriverID,
			DriverName: event.DriverName,
			Vehicle:    event.Vehicle,
			Lat:        event.DriverLat,
			Lng:        event.DriverLng,
		},
	})

	return nil
}

func (c *EventConsumer) handleTripCompleted(ctx context.Context, body []byte) error {
	event, err := rabbitmq.Decode[events.TripCompleted](body)
	if err != nil {
		c.logger.Error("invalid trip completed event", "error", err)
		return nil
	}

	msg := ws.Message{
		Type: ws.EventTripCompleted,
		Payload: ws.TripCompletedPayload{
			TripID:    event.TripID,
			FinalFare: event.FinalFare,
		},
	}

	c.manager.Send(event.RiderID, msg)
	c.manager.Send(event.DriverID, msg)

	return nil
}
