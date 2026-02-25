package model

import "context"

type BookingRepo interface {
	CreateBooking(ctx context.Context, bookingDetail CreateBooking) (CreateBooking, error)
}

type BookingService interface {
	CreateBooking(ctx context.Context, eventId, eventCat int32) error
}

type CreateBooking struct {
	ID        string
	EventID   int32
	TicketID  int32
	UserID    int32
	EventType int32
	Status    string
}
