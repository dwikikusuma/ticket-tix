package service

import (
	"ticket-tix/common/pkg/storage"
	"ticket-tix/service/ticket/internal/model"
)

type TicketService struct {
	storage *storage.Storage
	repo    model.TicketRepo
}

func NewTicketService(storage *storage.Storage, repo model.TicketRepo) *TicketService {
	return &TicketService{
		storage: storage,
		repo:    repo,
	}
}
