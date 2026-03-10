package repository

import (
	"strconv"
	ticketDB "ticket-tix/service/ticket/internal/infra/postgres"
	"ticket-tix/service/ticket/internal/model"
)

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
