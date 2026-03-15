package jobs

import (
	"context"
	"log"
	"ticket-tix/service/ticket/internal/model"
	"time"
)

type ExpireReservedSeatJob struct {
	repo model.TicketRepo
}

func NewExpireReservedSeatJob(repo model.TicketRepo) *ExpireReservedSeatJob {
	return &ExpireReservedSeatJob{repo: repo}
}

func (e *ExpireReservedSeatJob) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("context canceled, exiting expireReservedSeatJob")
			return
		case <-ticker.C:
			log.Println("starting expireReservedSeatJob")
			e.processExpireSeats(ctx)
		}
	}
}

func (e *ExpireReservedSeatJob) processExpireSeats(ctx context.Context) {
	expiredSeats, err := e.repo.ExpireReservedSeats(ctx)
	if err != nil {
		log.Printf("failed to expire reserved seats: %v", err)
		return
	}

	for _, seat := range expiredSeats {
		log.Printf("expired reserved \nid: %d \nseat: %s \nevent category %d", seat.TicketID, seat.SeatNumber, seat.EventCategoryID)
	}
}
