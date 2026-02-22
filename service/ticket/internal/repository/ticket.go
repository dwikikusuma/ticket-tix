package repository

import (
	"context"
	"database/sql"
	"fmt"
	ticketDB "ticket-tix/service/ticket/internal/infra/postgres"
	"ticket-tix/service/ticket/internal/model"
)

type ticketRepo struct {
	db    *ticketDB.Queries
	rawDB *sql.DB
}

func NewTicketRepo(db *sql.DB) model.TicketRepo {
	return &ticketRepo{
		db:    ticketDB.New(db),
		rawDB: db,
	}
}

func (r *ticketRepo) WithTx(tx *sql.Tx) model.TicketRepo {
	return &ticketRepo{
		db:    r.db.WithTx(tx),
		rawDB: r.rawDB,
	}
}

func (r *ticketRepo) InsertEvent(ctx context.Context, params model.EventData) (model.EventData, error) {
	event, err := r.db.InsertEvent(ctx, ticketDB.InsertEventParams{
		Name: params.Name,
		Description: sql.NullString{
			String: params.Description,
			Valid:  params.Description != "",
		},
		Location:  params.Location,
		StartTime: params.StartTime,
		EndTime:   params.EndTime,
	})
	if err != nil {
		return model.EventData{}, fmt.Errorf("insert event: %w", err)
	}

	return toModel(event), nil
}

func (r *ticketRepo) InsertEventImage(ctx context.Context, params model.ImageKeyData) error {
	_, err := r.db.InsertEventImage(ctx, ticketDB.InsertEventImageParams{
		EventID:  params.EventID,
		ImageKey: params.Key,
		IsPrimary: sql.NullBool{
			Bool:  params.IsPrimary,
			Valid: true,
		},
		DisplayOrder: sql.NullInt32{
			Int32: int32(params.DisplayOrder),
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("insert event image: %w", err)
	}
	return nil
}

func (r *ticketRepo) DeleteEventImage(ctx context.Context, key string) error {
	err := r.db.DeleteEventImage(ctx, key)
	if err != nil {
		return fmt.Errorf("delete event image: %w", err)
	}
	return nil
}

func toModel(e ticketDB.Event) model.EventData {
	return model.EventData{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description.String,
		Location:    e.Location,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
	}
}
