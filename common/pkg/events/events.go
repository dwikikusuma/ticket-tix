package events

import "time"

type BookingCreatedEvent struct {
	BookingID    string    `json:"booking_id"`
	UserID       int32     `json:"user_id"`
	EventID      int32     `json:"event_id"`
	EventCatID   int32     `json:"event_cat_id"`
	SeatNumber   string    `json:"seat_number"`
	CategoryType string    `json:"category_type"`
	OccurredAt   time.Time `json:"occurred_at"`
}
