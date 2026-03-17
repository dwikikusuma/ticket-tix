// service/bookings/internal/service/booking_service.go
package service

import (
	"context"
	"fmt"
	"log"
	ticketRPC "ticket-tix/common/gen/ticket/v1"
	"ticket-tix/common/pkg/lock"
	"ticket-tix/service/bookings/internal/model"
	"time"
)

type bookingService struct {
	repo      model.BookingRepo
	ticketSVC ticketRPC.TicketServiceClient
	lock      lock.DistributedLock
}

func NewBookingService(repo model.BookingRepo, ticketSvc ticketRPC.TicketServiceClient, lock lock.DistributedLock) model.BookingService {
	return &bookingService{repo: repo, ticketSVC: ticketSvc, lock: lock}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID int32, eventID, eventCat int32, seatID, bookType, categoryType string) error {
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
	seatKey := s.seatLockKey(eventCat, seatID)
	token, accErr := s.lock.Acquire(ctx, seatKey, 30*time.Second)
	if accErr != nil {
		return fmt.Errorf("acquire seat lock: %w", accErr)
	}
	defer s.lock.Release(ctx, seatKey, token)

	_, err := s.ticketSVC.ValidateTicket(ctx, &ticketRPC.ValidateTicketRequest{
		SeatId:        seatID,
		EventId:       eventID,
		EventCategory: eventCat,
	})
	if err != nil {
		return fmt.Errorf("validate ticket: %w", err)
	}

	reserved, err := s.ticketSVC.UpdateTicketStatus(ctx, &ticketRPC.UpdateTicketStatusRequest{
		SeatId:        seatID,
		EventCategory: eventCat,
		Status:        "RESERVED",
	})
	if err != nil {
		return fmt.Errorf("reserve ticket: %w", err)
	}

	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventID,
		EventType: eventCat,
		TicketID:  reserved.GetTicketId(),
		UserID:    userID, // ← no longer hardcoded
		Status:    "PENDING",
	})
	if err != nil {
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
	_, err := s.ticketSVC.DecreaseTicket(ctx, &ticketRPC.DecreaseTicketRequest{
		EventCategoryId: eventCat,
		DecreaseBy:      1,
	})

	if err != nil {
		return fmt.Errorf("decrease standing stock: %w", err)
	}

	_, err = s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventID,
		EventType: eventCat,
		TicketID:  0,
		UserID:    userID,
		Status:    "CONFIRMED",
	})
	if err != nil {
		_, incrErr := s.ticketSVC.IncreaseTicket(ctx, &ticketRPC.IncreaseTicketRequest{
			EventCategoryId: eventCat,
			IncreaseBy:      1,
		})
		log.Println("incrErr", incrErr) // implement dlq for failed compensations in production
		return fmt.Errorf("create standing booking: %w", err)
	}

	return nil
}

func (s *bookingService) seatLockKey(eventCatID int32, seatID string) string {
	return fmt.Sprintf("lock:seat:%d:%s", eventCatID, seatID)
}
