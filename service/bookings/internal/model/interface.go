package model

import "context"

type BookingRepo interface {
	CreateBooking(ctx context.Context, bookingDetail CreateBooking) (CreateBooking, error)
}

type BookingService interface {
	CreateBooking(ctx context.Context, eventId, eventCat int32, seatID string) error
}

type CreateBooking struct {
	ID        string
	EventID   int32
	TicketID  int32
	UserID    int32
	EventType int32
	Status    string
}

type (
	BookingRequest struct {
		EventID  int32  `json:"event_id" binding:"required"`
		EventCat int32  `json:"event_cat" binding:"required"`
		SeatID   string `json:"seat_id"`
	}
)
