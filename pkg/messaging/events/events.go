// Package events defines all async events exchanged between services.
// These are the message contracts — both publisher and consumer
// must agree on the structure.
package events

const (
	// Exchange is the single topic exchange all services use.
	Exchange = "rideflow.events"

	// Routing keys — used by publishers and bindings
	RoutingKeyTripCreated    = "trip.created"
	RoutingKeyDriverAssigned = "driver.assigned"
	RoutingKeyTripCompleted  = "trip.completed"

	// Queue names — one per consumer per event type
	QueueDriverTripCreated    = "driver.trip.created"
	QueueTripDriverAssigned   = "trip.driver.assigned"
	QueuePaymentTripCompleted = "payment.trip.completed"
)

// TripCreated is published by the Trip Service when a new trip is created.
// The Driver Service consumes this to find and assign a driver.
type TripCreated struct {
	TripID      string  `json:"trip_id"`
	RiderID     string  `json:"rider_id"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	OriginLat   float64 `json:"origin_lat"`
	OriginLng   float64 `json:"origin_lng"`
}

// DriverAssigned is published by the Driver Service when a driver accepts.
// The Trip Service consumes this to update the trip status.
type DriverAssigned struct {
	TripID   string `json:"trip_id"`
	DriverID string `json:"driver_id"`
}

// TripCompleted is published by the Trip Service when a trip ends.
// The Payment Service consumes this to initiate payment.
type TripCompleted struct {
	TripID    string  `json:"trip_id"`
	RiderID   string  `json:"rider_id"`
	DriverID  string  `json:"driver_id"`
	FinalFare float64 `json:"final_fare"`
}
