package repository

import (
	"context"
	"database/sql"
	"log"
	ticketDB "ticket-tix/service/ticket/internal/infra/postgres"
	"ticket-tix/service/ticket/internal/model"
)

type ticketRepo struct {
	db *ticketDB.Queries
}

func NewTicketRepo(db *sql.DB) model.TicketRepo {
	return &ticketRepo{
		db: ticketDB.New(db),
	}
}

func (r *ticketRepo) CreateEvent(ctx context.Context, params model.EventData) (model.EventData, error) {
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
		log.Println("failed to insert event: ", err)
		return model.EventData{}, err
	}

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
