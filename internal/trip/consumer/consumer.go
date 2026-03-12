package consumer

import (
	"context"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/trip/service"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

// DriverConsumer handles driver-related events for the Trip Service.
type DriverConsumer struct {
	svc      *service.TripService
	consumer *rabbitmq.Consumer
	logger   *slog.Logger
}

func NewDriverConsumer(
	svc *service.TripService,
	consumer *rabbitmq.Consumer,
	logger *slog.Logger,
) *DriverConsumer {
	return &DriverConsumer{
		svc:      svc,
		consumer: consumer,
		logger:   logger,
	}
}

// Start begins consuming DriverAssigned events.
// Blocks until context is cancelled — run in a goroutine.
func (c *DriverConsumer) Start(ctx context.Context) error {
	return c.consumer.Subscribe(
		ctx,
		events.QueueTripDriverAssigned,
		events.RoutingKeyDriverAssigned,
		c.handleDriverAssigned,
	)
}

func (c *DriverConsumer) handleDriverAssigned(ctx context.Context, body []byte) error {
	event, err := rabbitmq.Decode[events.DriverAssigned](body)
	if err != nil {
		c.logger.Error("invalid driver assigned event", "error", err)
		return nil
	}

	c.logger.Info("driver assigned event received",
		"trip_id", event.TripID,
		"driver_id", event.DriverID,
	)

	// Idempotency check happens inside AssignDriver —
	// if the driver is already assigned, it returns an invalid
	// transition error which we log and ack.
	if err := c.svc.AssignDriver(ctx, event.TripID, event.DriverID); err != nil {
		c.logger.Error("failed to assign driver to trip",
			"trip_id", event.TripID,
			"driver_id", event.DriverID,
			"error", err,
		)
		return err
	}

	return nil
}
