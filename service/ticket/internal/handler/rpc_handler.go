package handler

import (
	"context"
	ticketRPC "ticket-tix/common/gen/ticket/v1"
	"ticket-tix/service/ticket/internal/infra/redis"
	"ticket-tix/service/ticket/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RPCHandler struct {
	ticketRPC.UnimplementedTicketServiceServer
	svc          model.TicketService
	stockCounter redis.StockCounter
}

func NewRPCHandler(svc model.TicketService, stockCounter redis.StockCounter) *RPCHandler {
	return &RPCHandler{
		svc:          svc,
		stockCounter: stockCounter,
	}
}

func (h *RPCHandler) UpdateTicketStatus(ctx context.Context, req *ticketRPC.UpdateTicketStatusRequest) (*ticketRPC.UpdateTicketStatusResponse, error) {
	ticketID, err := h.svc.UpdateTicketStatus(ctx, req.Status, req.SeatId, req.EventCategory)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update ticket status: %v", err)
	}
	return &ticketRPC.UpdateTicketStatusResponse{
		TicketId: ticketID,
		Status:   req.Status,
	}, nil
}

func (h *RPCHandler) ValidateTicket(ctx context.Context, req *ticketRPC.ValidateTicketRequest) (*ticketRPC.ValidateTicketResponse, error) {
	if err := h.svc.ValidateTicketBooking(ctx, req.SeatId, req.EventId, req.EventCategory); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to validate ticket booking: %v", err)
	}
	return &ticketRPC.ValidateTicketResponse{
		IsValid: true,
	}, nil
}

func (h *RPCHandler) ReserveAvailableSeat(ctx context.Context, req *ticketRPC.ReserveFlexibleSeatRequest) (*ticketRPC.ReserveFlexibleSeatResponse, error) {
	seatNum, tixID, err := h.svc.ReserveAvailableSeat(ctx, req.GetEventCategoryId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve available seat: %v", err)
	}
	return &ticketRPC.ReserveFlexibleSeatResponse{
		TicketId:   tixID,
		SeatNumber: seatNum,
	}, nil
}

func (h *RPCHandler) DecreaseTicket(ctx context.Context, req *ticketRPC.DecreaseTicketRequest) (*ticketRPC.DecreaseTicketResponse, error) {
	eventCat := req.GetEventCategoryId()
	decreaseBy := req.GetDecreaseBy()
	if err := h.stockCounter.Decrement(ctx, eventCat, decreaseBy); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decrease ticket stock: %v", err)
	}
	return &ticketRPC.DecreaseTicketResponse{}, nil
}

func (h *RPCHandler) IncreaseTicket(ctx context.Context, req *ticketRPC.IncreaseTicketRequest) (*ticketRPC.IncreaseTicketResponse, error) {
	eventCat := req.GetEventCategoryId()
	increaseBy := req.GetIncreaseBy()

	if err := h.stockCounter.Increment(ctx, eventCat, increaseBy); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to increase ticket stock: %v", err)
	}
	return &ticketRPC.IncreaseTicketResponse{}, nil
}
