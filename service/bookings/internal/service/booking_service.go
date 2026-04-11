// service/bookings/internal/service/booking_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	ticketRPC "ticket-tix/common/gen/ticket/v1"
	"ticket-tix/common/pkg/events"
	"ticket-tix/common/pkg/lock"
	"ticket-tix/service/bookings/internal/model"
	"time"
)

const (
	bookingCreatedTopic = "booking.created"
)

type bookingService struct {
	repo      model.BookingRepo
	ticketSVC ticketRPC.TicketServiceClient
	lock      lock.DistributedLock
	producer  events.Producer
}

func NewBookingService(repo model.BookingRepo, ticketSvc ticketRPC.TicketServiceClient, lock lock.DistributedLock, producer events.Producer) model.BookingService {
	return &bookingService{repo: repo, ticketSVC: ticketSvc, lock: lock, producer: producer}
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

	reserved, err := s.ticketSVC.ReserveTicket(ctx, &ticketRPC.ReserveTicketRequest{
		SeatId:        seatID,
		EventCategory: eventCat,
	})
	if err != nil {
		return fmt.Errorf("reserve ticket: %w", err)
	}

	bookingData, err := s.repo.CreateBooking(ctx, model.CreateBooking{
		EventID:   eventID,
		EventType: eventCat,
		TicketID:  reserved.GetTicketId(),
		UserID:    userID,
		Status:    "PENDING",
	})
	if err != nil {
		_, releaseErr := s.ticketSVC.ReleaseTicket(ctx, &ticketRPC.ReleaseTicketRequest{
			SeatId:        seatID,
			EventCategory: eventCat,
		})
		if releaseErr != nil {
			return fmt.Errorf("failed release ticket: %w", err)
		}
		return fmt.Errorf("create booking: %w", err)
	}
	s.publishBookingCreated(ctx, bookingData, "SEATED-FIXED")
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

	bookingData, err := s.repo.CreateBooking(ctx, model.CreateBooking{
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

	s.publishBookingCreated(ctx, bookingData, "SEATED_FLEXIBLE")
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

	bookingData, err := s.repo.CreateBooking(ctx, model.CreateBooking{
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
		log.Println("incrErr", incrErr)
		return fmt.Errorf("create standing booking: %w", err)
	}

	s.publishBookingCreated(ctx, bookingData, "STANDING")
	return nil
}

func (s *bookingService) seatLockKey(eventCatID int32, seatID string) string {
	return fmt.Sprintf("lock:seat:%d:%s", eventCatID, seatID)
}

func (s *bookingService) publishBookingCreated(ctx context.Context, booking model.CreateBooking, categoryType string) {
	event := events.BookingCreatedEvent{
		BookingID:    booking.ID,
		UserID:       booking.UserID,
		EventID:      booking.EventID,
		EventCatID:   booking.EventType,
		SeatNumber:   booking.SeatNumber,
		CategoryType: categoryType,
		OccurredAt:   time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal booking event: %v", err)
		return
	}

	msg := events.Message{
		Key:   []byte(booking.ID),
		Value: payload,
		Headers: map[string]string{
			"event-type":     "booking.created",
			"content-type":   "application/json",
			"source":         "booking-service",
			"correlation-id": booking.ID,
		},
	}

	if err := s.producer.Publish(ctx, bookingCreatedTopic, msg); err != nil {
		log.Printf("failed to publish booking.created event: %v", err)
	}
}
