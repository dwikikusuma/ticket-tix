package model

import "context"

type TicketRepo interface {
	CreateEvent(ctx context.Context, params EventData) (EventData, error)
}
