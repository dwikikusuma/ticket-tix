package service

import (
	"context"
	"fmt"
	"ticket-tix/common/pkg/storage"
	"ticket-tix/service/ticket/internal/model"

	"github.com/google/uuid"
)

type TicketService struct {
	storage *storage.Storage
	repo    model.TicketRepo
}

func NewTicketService(storage *storage.Storage, repo model.TicketRepo) *TicketService {
	return &TicketService{storage: storage, repo: repo}
}

func (s *TicketService) CreateEvent(ctx context.Context, req model.InsertTicketRequest) (model.EventData, error) {
	var imageKeys []model.ImageKeyData

	for _, f := range req.Files {
		key := fmt.Sprintf("events/%s-%s", uuid.New().String(), f.Filename)

		if err := s.storage.UploadImage(ctx, f.Reader, key, f.Size, f.ContentType); err != nil {
			return model.EventData{}, fmt.Errorf("failed to upload image %s: %w", f.Filename, err)
		}

		imageKeys = append(imageKeys, model.ImageKeyData{
			Key:          key,
			IsPrimary:    f.IsPrimary,
			DisplayOrder: f.DisplayOrder,
		})
	}

	event, err := s.repo.CreateEvent(ctx, req.Event, imageKeys)
	if err != nil {
		for _, img := range imageKeys {
			s.storage.Delete(ctx, img.Key)
		}
		return model.EventData{}, err
	}

	return event, nil
}
