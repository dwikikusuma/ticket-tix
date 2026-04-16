package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"ticket-tix/common/pkg/events"
	"ticket-tix/service/fullfilement/internal/model"
)

const (
	confirmedStatus = "confirmed"
)

type Handler struct {
	logger  *slog.Logger
	service model.FulfillmentService
}

func NewHandler(logger *slog.Logger, service model.FulfillmentService) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) HandleOrderCreated(ctx context.Context, msg events.Message) error {
	var request events.BookingCreatedEvent
	if err := json.Unmarshal(msg.Value, &request); err != nil {
		h.logger.Error("failed to unmarshal booking created event", "error", err)
		return err
	}
	bookingData := h.toBooking(request)
	err := h.service.InsertFulfillment(ctx, bookingData)
	if err != nil {
		h.logger.Error("failed to insert fulfillment", "error", err)
		return err
	}

	h.logger.Info("successfully processed booking created event", "booking_id", request.BookingID)
	return nil
}

func (h *Handler) toBooking(req events.BookingCreatedEvent) model.Booking {
	return model.Booking{
		BookingID: req.BookingID,
		UserID:    strconv.Itoa(int(req.UserID)),
		Status:    confirmedStatus,
		Category:  req.CategoryType,
		SeatID:    req.SeatNumber,
	}
}
