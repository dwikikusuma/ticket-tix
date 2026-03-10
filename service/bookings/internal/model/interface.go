package model

import "context"

type BookingRepo interface {
	CreateBooking(ctx context.Context, bookingDetail CreateBooking) (CreateBooking, error)
}

type BookingService interface {
	CreateBooking(ctx context.Context, userID int32, eventID, eventCat int32, seatID, bookType, categoryType string) error
}

type CreateBooking struct {
	ID        string
	EventID   int32
	TicketID  int32
	UserID    int32
	EventType int32
	Status    string
}

type BookingRequest struct {
	EventID      int32  `json:"event_id"      binding:"required"`
	EventCat     int32  `json:"event_cat"     binding:"required"`
	SeatID       string `json:"seat_id"`                          // empty for FLEXIBLE and STANDING
	BookType     string `json:"book_type"     binding:"required"` // "FIXED" or "FLEXIBLE"
	CategoryType string `json:"category_type" binding:"required"` // "SEATED" or "STANDING"
}
