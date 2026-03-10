// service/bookings/internal/service/booking_service.go
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
	return &bookingService{repo: repo, ticketSVC: ticketSvc}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID int32, eventID, eventCat int32, seatID, bookType, categoryType string) error {
	// Route based on WHAT TYPE OF BOOKING this is.
	// Your old code only checked seatID != "" which is wrong —
	// FLEXIBLE also sends empty seatID but needs completely different handling.
	switch categoryType {
	case "STANDING":
		return s.bookStanding(ctx, userID, eventID, eventCat)
	case "SEATED":
		switch bookType {
		case "FIXED":
			return s.bookSeatedFixed(ctx, userID, eventID, eventCat, seatID)
		case "FLEXIBLE":
			return s.bookSeatedFlexible(ctx, userID, eventID, eventCat)
		}
	}
	return fmt.Errorf("unknown category_type/book_type combination: %s/%s", categoryType, bookType)
}

// SEATED + FIXED: client picks a specific seat
func (s *bookingService) bookSeatedFixed(ctx context.Context, userID, eventID, eventCat int32, seatID string) error {
	// Step 1: validate the seat exists and is AVAILABLE
	// (ticket-service checks status, event ownership, etc.)
	_, err := s.ticketSVC.ValidateTicket(ctx, &ticketRPC.ValidateTicketRequest{
		SeatId:        seatID,
		EventId:       eventID,
		EventCategory: eventCat,
	})
	if err != nil {
		return fmt.Errorf("validate ticket: %w", err)
	}

	// Step 2: reserve the seat — AVAILABLE → RESERVED
	// "RESERVED" means "held while user pays", not "permanently booked"
	reserved, err := s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
		SeatId:        seatID,
		EventCategory: eventCat,
		Status:        "RESERVED",
	})
	if err != nil {
		return fmt.Errorf("reserve ticket: %w", err)
	}

	// Step 3: save the booking record in our DB
	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventID,
		EventType: eventCat,
		TicketID:  reserved.GetTicketId(),
		UserID:    userID, // ← no longer hardcoded
		Status:    "PENDING",
	})
	if err != nil {
		// Something went wrong saving — release the seat we just reserved.
		// This is "best-effort compensation" — if this also fails, the expiry job
		// will clean it up in 15 minutes. That's fine for now.
		s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
			SeatId:        seatID,
			EventCategory: eventCat,
			Status:        "AVAILABLE",
		})
		return fmt.Errorf("create booking: %w", err)
	}
	return nil
}

func (s *bookingService) bookSeatedFlexible(ctx context.Context, userID, eventID, eventCat int32) error {
	resp, err := s.ticketSVC.ReserveSeat(ctx, &ticketRPC.ReserveFlexibleSeatRequest{
		EventCategoryId: eventCat,
		EventId:         eventID,
	})

	if err != nil {
		return fmt.Errorf("reserve flexible seat: %w", err)
	}

	// Save booking — include the assigned seat number so user knows where they're sitting
	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:    eventID,
		EventType:  eventCat,
		TicketID:   resp.GetTicketId(),
		UserID:     userID,
		SeatNumber: resp.GetSeatNumber(),
		Status:     "CONFIRMED",
	})
	if err != nil {
		return fmt.Errorf("create booking after flexible reserve: %w", err)
	}

	return nil
}

// STANDING: no seat, just check there's capacity
func (s *bookingService) bookStanding(ctx context.Context, userID, eventID, eventCat int32) error {
	// Step 1: validate capacity exists (stock check)
	_, err := s.ticketSVC.ValidateTicket(ctx, &ticketRPC.ValidateTicketRequest{
		SeatId:        "", // empty — no seat for standing
		EventId:       eventID,
		EventCategory: eventCat,
	})
	if err != nil {
		return fmt.Errorf("validate standing: %w", err)
	}

	// Step 2: save booking — no ticketID for standing
	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventID,
		EventType: eventCat,
		TicketID:  0, // null — standing tickets have no individual seat row
		UserID:    userID,
		Status:    "CONFIRMED", // standing goes straight to confirmed, no payment gate
	})
	if err != nil {
		return fmt.Errorf("create standing booking: %w", err)
	}

	// NOTE: stock decrement missing here — this is BUG-14, fixed in TIX-004.
	// Right now two concurrent STANDING bookings can both succeed even if stock=1.
	// That's your next ticket after this baseline works.
	return nil
}
