package model

import (
	"context"
	"database/sql"
)

type TicketRepo interface {
	WithTx(tx *sql.Tx) TicketRepo
	InsertEvent(ctx context.Context, params EventData) (EventData, error)
	InsertEventImage(ctx context.Context, params ImageKeyData) error
	DeleteEventImage(ctx context.Context, key string) error
	GetEventByID(ctx context.Context, id int32) (EventData, error)
	GetEventCategory(ctx context.Context, eventID int32) ([]EventCategoryData, error)
	GetEventImages(ctx context.Context, eventID int32) ([]EventImageData, error)
	BrowseEvents(ctx context.Context, filter BrowseFilter) ([]EventData, error)
	UpdateTicketStatus(ctx context.Context, status, seatNum string, eventID int32) (int32, error)
}

type TicketService interface {
	CreateEvent(ctx context.Context, req InsertTicketRequest) (EventData, error)
	DeleteImg(ctx context.Context, key string) error
	UploadImg(ctx context.Context, eventID int32, files []FileData) error
	GetEventDetail(ctx context.Context, id int32) (EventDetailsData, error)
	BrowseEvents(ctx context.Context, filter BrowseFilter) (BrowseResult, error)
	UpdateTicketStatus(ctx context.Context, status, seatNum string, eventID int32) (int32, error)
}

type ImageKeyData struct {
	EventID      int32
	Key          string
	IsPrimary    bool
	DisplayOrder int
}
