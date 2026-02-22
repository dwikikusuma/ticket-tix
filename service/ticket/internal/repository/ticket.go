package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

func (r *ticketRepo) CreateEvent(ctx context.Context, params model.EventData, imageKeys []model.ImageKeyData) (model.EventData, error) {
	tx, err := r.rawDB.BeginTx(ctx, nil)
	if err != nil {
		return model.EventData{}, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback() // no-op if committed

	qtx := r.db.WithTx(tx)

	// 1. insert event
	event, err := qtx.InsertEvent(ctx, ticketDB.InsertEventParams{
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
		return model.EventData{}, fmt.Errorf("failed to insert event: %w", err)
	}

	// 2. insert all image records
	for _, img := range imageKeys {
		_, err := qtx.InsertEventImage(ctx, ticketDB.InsertEventImageParams{
			EventID:  event.ID,
			ImageKey: img.Key,
			IsPrimary: sql.NullBool{
				Bool:  img.IsPrimary,
				Valid: true,
			},
			DisplayOrder: sql.NullInt32{
				Int32: int32(img.DisplayOrder),
				Valid: true,
			},
		})
		if err != nil {
			return model.EventData{}, fmt.Errorf("failed to insert event image: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return model.EventData{}, fmt.Errorf("failed to commit tx: %w", err)
	}

	log.Printf("event created: id=%d images=%d", event.ID, len(imageKeys))
	return toModel(event), nil
}

func toModel(e ticketDB.Event) model.EventData {
	return model.EventData{
		Name:        e.Name,
		Description: e.Description.String,
		Location:    e.Location,
		StartTime:   e.StartTime,
		EndTime:     e.EndTime,
	}
}
