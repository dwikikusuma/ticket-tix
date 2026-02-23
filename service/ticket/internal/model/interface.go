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
}

type TicketService interface {
	CreateEvent(ctx context.Context, req InsertTicketRequest) (EventData, error)
	DeleteImg(ctx context.Context, key string) error
	UploadImg(ctx context.Context, eventID int32, file FileData) error
	GetEventDetail(ctx context.Context, id int32) (EventDetailsData, error)
}

type ImageKeyData struct {
	EventID      int32
	Key          string
	IsPrimary    bool
	DisplayOrder int
}
