package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	ticketDB "ticket-tix/service/ticket/internal/infra/postgres"
	"ticket-tix/service/ticket/internal/model"
	"time"
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

func (r *ticketRepo) GetEventByID(ctx context.Context, id int32) (model.EventData, error) {
	event, err := r.db.GetEventDetails(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.EventData{}, fmt.Errorf("event not found")
		}
		return model.EventData{}, fmt.Errorf("failed get event by id: %w", err)
	}
	return toModel(event), nil
}

func (r *ticketRepo) GetEventCategory(ctx context.Context, eventID int32) ([]model.EventCategoryData, error) {
	evenCat, err := r.db.GetEventCategories(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event category not found")
		}
		return nil, fmt.Errorf("failed get event category: %w", err)
	}

	var res []model.EventCategoryData
	for _, ec := range evenCat {
		res = append(res, toModelEventCategory(ec))
	}
	return res, nil
}

func (r *ticketRepo) GetEventImages(ctx context.Context, eventID int32) ([]model.EventImageData, error) {
	images, err := r.db.GetEventImages(ctx, eventID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("event images not found")
		}
		return nil, fmt.Errorf("failed get event images: %w", err)
	}

	var res []model.EventImageData
	for _, img := range images {
		res = append(res, model.EventImageData{
			EventID:      img.EventID,
			Key:          img.ImageKey,
			IsPrimary:    img.IsPrimary.Bool,
			DisplayOrder: int(img.DisplayOrder.Int32),
		})
	}
	return res, nil
}
func (r *ticketRepo) BrowseEvents(ctx context.Context, filter model.BrowseFilter) ([]model.EventData, error) {
	var cursorTime time.Time
	var cursorID int32

	if filter.Cursor != nil {
		cursorTime = filter.Cursor.StartTime
		cursorID = filter.Cursor.ID
	}

	rows, err := r.db.BrowseEvents(ctx, ticketDB.BrowseEventsParams{
		EventName:  filter.EventName,
		Location:   filter.Location,
		StartDate:  filter.StartDate,
		EndDate:    filter.EndDate,
		CursorTime: cursorTime,
		CursorID:   cursorID,
		PageSize:   int32(filter.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("browse events: %w", err)
	}

	var events []model.EventData
	for _, row := range rows {
		events = append(events, model.EventData{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description.String,
			Location:    row.Location,
			StartTime:   row.StartTime,
			EndTime:     row.EndTime,
			ImageURL:    row.ImageKey,
		})
	}

	return events, nil
}

func (r *ticketRepo) UpdateTicketStatus(ctx context.Context, status, seatNum string, eventID int32) (int32, error) {
	bookTime := time.Now().Add(15 * time.Minute)
	ticketID, err := r.db.UpdateTicketStatus(ctx, ticketDB.UpdateTicketStatusParams{
		Status:          sql.NullString{String: status, Valid: true},
		SeatNumber:      sql.NullString{String: seatNum, Valid: true},
		EventCategoryID: eventID,
		ReservedUntil:   sql.NullTime{Time: bookTime, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("update ticket status: %w", err)
	}
	return ticketID, nil

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

func toModelEventCategory(ec ticketDB.EventCategory) model.EventCategoryData {
	price, _ := strconv.ParseFloat(ec.Price, 64)
	return model.EventCategoryData{
		EventID:           ec.EventID,
		CategoryID:        ec.ID,
		CategoryType:      ec.CategoryType.String,
		Price:             price,
		BookType:          ec.BookType,
		TotalCapacity:     ec.TotalCapacity,
		AvailableCapacity: ec.AvailableStock,
	}
}
