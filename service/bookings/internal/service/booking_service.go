package service

import (
	"context"
	"ticket-tix/service/bookings/internal/model"
)

type bookingService struct {
	repo model.BookingRepo
}

func newBookingService(repo model.BookingRepo) model.BookingService {
	return &bookingService{
		repo: repo,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, eventId, eventCat int32) error {
	// Status: will set to Book and updated after create payment event
	// TicketID: will call rpc to ticket
	// UserID: will get from auth for now set to dummy value

	return nil
}
