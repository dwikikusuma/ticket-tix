package servcie

import (
	"context"
	"ticket-tix/service/fullfilement/internal/model"
)

type fulfillmentService struct {
	repo model.FulfillmentRepo
}

func NewService(repo model.FulfillmentRepo) model.FulfillmentService {
	return &fulfillmentService{repo: repo}
}

func (s *fulfillmentService) InsertFulfillment(ctx context.Context, booking model.Booking) error {
	return s.repo.CreateFulfillment(ctx, booking)
}

func (s *fulfillmentService) UpdateCancelledFulfillment(ctx context.Context, bookingID string) error {
	return s.repo.CancelFulfillment(ctx, bookingID)
}
