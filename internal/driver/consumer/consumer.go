// Package consumer contains all RabbitMQ message consumers
// for the Driver Service.
package consumer

import (
	"context"
	"log/slog"

	"github.com/bytepharoh/rideflow/internal/driver/service"
	"github.com/bytepharoh/rideflow/pkg/messaging/events"
	"github.com/bytepharoh/rideflow/pkg/messaging/rabbitmq"
)

// TripConsumer handles trip-related events for the Driver Service.
type TripConsumer struct {
	svc       *service.DriverService
	consumer  *rabbitmq.Consumer
	publisher *rabbitmq.Publisher
	logger    *slog.Logger
}

func NewTripConsumer(
	svc *service.DriverService,
	consumer *rabbitmq.Consumer,
	publisher *rabbitmq.Publisher,
	logger *slog.Logger,
) *TripConsumer {
	return &TripConsumer{
		svc:       svc,
		consumer:  consumer,
		publisher: publisher,
		logger:    logger,
	}
}

// Start begins consuming TripCreated events.
// Blocks until the context is cancelled — run in a goroutine.
func (c *TripConsumer) Start(ctx context.Context) error {
	return c.consumer.Subscribe(
		ctx,
		events.QueueDriverTripCreated,
		events.RoutingKeyTripCreated,
		c.handleTripCreated,
	)
}

func (c *TripConsumer) handleTripCreated(ctx context.Context, body []byte) error {
	event, err := rabbitmq.Decode[events.TripCreated](body)
	if err != nil {
		// Bad message format — nacking won't help, it will just loop forever.
		// Log and ack to discard. In production, route to a dead-letter queue.
		c.logger.Error("invalid trip created event", "error", err)
		return nil
	}

	c.logger.Info("trip created event received", "trip_id", event.TripID)

	// Find the nearest available driver to the pickup location
	driver, err := c.svc.FindNearestAvailable(ctx, event.OriginLat, event.OriginLng)
	if err != nil {
		c.logger.Warn("no available driver for trip",
			"trip_id", event.TripID,
			"error", err,
		)
		// Return nil to ack — no point retrying if no drivers are available.
		// In production: retry with backoff, notify rider of wait time.
		return nil
	}

	// Assign the driver
	if err := c.svc.AssignTrip(ctx, driver.ID, event.TripID); err != nil {
		c.logger.Error("failed to assign driver",
			"driver_id", driver.ID,
			"trip_id", event.TripID,
			"error", err,
		)
		return err // nack — retry this
	}

	// Publish DriverAssigned event — Trip Service will consume this.
	assigned := events.DriverAssigned{
		TripID:     event.TripID,
		DriverID:   driver.ID,
		RiderID:    event.RiderID,
		DriverName: driver.Name,
		Vehicle:    driver.Vehicle,
		DriverLat:  driver.Location.Latitude,
		DriverLng:  driver.Location.Longitude,
	}

	if err := c.publisher.Publish(ctx, events.RoutingKeyDriverAssigned, assigned); err != nil {
		c.logger.Error("failed to publish driver assigned event",
			"trip_id", event.TripID,
			"error", err,
		)
		return err // nack — retry
	}

	c.logger.Info("driver assigned",
		"trip_id", event.TripID,
		"driver_id", driver.ID,
	)

	return nil
}
