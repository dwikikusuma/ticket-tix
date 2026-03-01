package service

import (
	"context"
	"fmt"
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
	// Step 1: Validate the ticket/seat is bookable
	isValid, err := s.ticketSVC.ValidateTicket(ctx, &ticketRPC.ValidateTicketRequest{
		SeatId:        seatID,
		EventId:       eventId,
		EventCategory: eventCat,
	})

	// FIX: original code did `if err != nil || !isValid.GetIsValid() { return err }`
	// which silently returns nil when err != nil (swallowed error).
	if err != nil {
		return fmt.Errorf("validate ticket: %w", err)
	}
	if !isValid.GetIsValid() {
		// This branch shouldn't normally be hit because the RPC returns an error
		// when invalid, but guard it explicitly to be safe.
		return fmt.Errorf("ticket is not valid for booking")
	}

	// Step 2: For SEATED tickets, reserve the specific seat via RPC.
	// For STANDING tickets seatID is empty — no seat reservation needed,
	// but available_stock decrement must happen (see NOTE below).
	var ticketID int32
	if seatID != "" {
		// FIX: was passing "Booked" which is rejected by ticket service.
		// Correct status at this stage is "RESERVED" (held during checkout).
		// "SOLD" should only be set after payment is confirmed.
		bookedTicket, err := s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
			SeatId:        seatID,
			EventCategory: eventCat,
			Status:        "RESERVED",
		})
		if err != nil {
			return fmt.Errorf("reserve ticket: %w", err)
		}
		ticketID = bookedTicket.GetTicketId()
	}

	// NOTE: For STANDING tickets, available_stock is not decremented here.
	// This needs a dedicated RPC (e.g. DecrementStock) or the ValidateTicket
	// RPC should handle it atomically (validate + decrement in one DB tx).
	// Without this, concurrent standing ticket bookings will oversell.

	// Step 3: Persist the booking record
	// FIX: original repo call was missing EventID, causing a NOT NULL DB error.
	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventId,
		EventType: eventCat,
		TicketID:  ticketID,
		UserID:    1, // TODO: extract from auth context
		Status:    "PENDING",
	})
	if err != nil {
		// If DB write fails after reserving the seat, attempt to roll back the reservation.
		// This is best-effort — a background job should also expire stale RESERVED tickets.
		if seatID != "" {
			if _, releaseErr := s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
				SeatId:        seatID,
				EventCategory: eventCat,
				Status:        "AVAILABLE",
			}); releaseErr != nil {
				// Log but don't mask the original error; the reserved_until expiry will clean this up
				_ = releaseErr
			}
		}
		return fmt.Errorf("create booking record: %w", err)
	}

	return nil
}
