package model

import "context"

type TicketService interface {
	CreateEvent(ctx context.Context, req InsertTicketRequest) (EventData, error)
}

type TicketRepo interface {
	CreateEvent(ctx context.Context, params EventData, imageKeys []ImageKeyData) (EventData, error)
}

type ImageKeyData struct {
	Key          string
	IsPrimary    bool
	DisplayOrder int
}
