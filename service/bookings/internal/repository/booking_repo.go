package repository

import (
	"context"
	"database/sql"
	bookingDB "ticket-tix/service/bookings/internal/infra/postgres"
	"ticket-tix/service/bookings/internal/model"
)

type bookingRepo struct {
	db *bookingDB.Queries
}

func NewBookingRepo(db *sql.DB) model.BookingRepo {
	return &bookingRepo{
		db: bookingDB.New(db),
	}
}

func (r *bookingRepo) CreateBooking(ctx context.Context, bookingDetail model.CreateBooking) (model.CreateBooking, error) {
	var ticketIDNullInt sql.NullInt32
	if bookingDetail.TicketID != 0 {
		ticketIDNullInt = sql.NullInt32{Int32: bookingDetail.TicketID, Valid: true}
	} else {
		ticketIDNullInt = sql.NullInt32{Valid: false}
	}

	// FIX: EventID was missing from CreateBookingParams despite being NOT NULL in the DB schema.
	// This would cause a runtime DB error on every standing-ticket booking and any
	// seated booking where EventID wasn't threaded through correctly.
	createdBooking, err := r.db.CreateBooking(ctx, bookingDB.CreateBookingParams{
		TicketID:        ticketIDNullInt,
		UserID:          bookingDetail.UserID,
		EventCategoryID: bookingDetail.EventType,
		EventID:         bookingDetail.EventID,
		Status:          bookingDetail.Status,
	})
	if err != nil {
		return model.CreateBooking{}, err
	}

	return model.CreateBooking{
		ID:        createdBooking.ID.String(),
		TicketID:  createdBooking.TicketID.Int32,
		UserID:    createdBooking.UserID,
		EventID:   createdBooking.EventID,
		EventType: createdBooking.EventCategoryID,
		Status:    createdBooking.Status,
	}, nil
}
