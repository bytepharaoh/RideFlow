package ws

type EventType string

const (
	EventDriverAssigned  EventType = "driver_assigned"
	EventTripStarted     EventType = "trip_started"
	EventTripCompleted   EventType = "trip_completed"
	EventTripCancelled   EventType = "trip_cancelled"
	EventPaymentRequired EventType = "payment_required"
)

type Message struct {
	Type    EventType `json:"type"`
	Payload any       `json:"payload"`
}

type DriverAssignedPayload struct {
	TripID     string  `json:"trip_id"`
	DriverID   string  `json:"driver_id"`
	DriverName string  `json:"driver_name"`
	Vehicle    string  `json:"vehicle"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

type TripCompletedPayload struct {
	TripID    string  `json:"trip_id"`
	FinalFare float64 `json:"final_fare"`
}
