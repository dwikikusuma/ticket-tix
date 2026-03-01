package service

import (
	"context"
	ticketRPC "ticket-tix/common/gen/ticket/v1"
	"ticket-tix/service/bookings/internal/model"
)

type bookingService struct {
	repo      model.BookingRepo
	ticketSVC ticketRPC.TicketServiceClient
}

func NewBookingService(repo model.BookingRepo, ticketSvc ticketRPC.TicketServiceClient) model.BookingService {
	return &bookingService{
		repo:      repo,
		ticketSVC: ticketSvc,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, eventId, eventCat int32, seatID string) error {
	// Status: will set to Book and updated after create payment event
	// TicketID: will call rpc to ticket
	// UserID: will get from auth for now set to dummy value

	isValid, err := s.ticketSVC.ValidateTicket(ctx, &ticketRPC.ValidateTicketRequest{
		SeatId:        seatID,
		EventId:       eventId,
		EventCategory: eventCat,
	})

	if err != nil || !isValid.GetIsValid() {
		return err
	}

	var ticketID int32
	if seatID != "" {
		bookedTicket, err := s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
			SeatId:        seatID,
			EventCategory: eventCat,
			Status:        "Booked",
		})

		if err != nil {
			return err
		}

		ticketID = bookedTicket.GetTicketId()
	}

	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventId,
		EventType: eventCat,
		TicketID:  ticketID,
		UserID:    1,
		Status:    "Book",
	})

	if err != nil {
		return err
	}
	return nil
}
