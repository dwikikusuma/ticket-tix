package handler

import (
	"context"
	ticketRPC "ticket-tix/common/gen/ticket/v1"
	"ticket-tix/service/ticket/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RPCHandler struct {
	ticketRPC.UnimplementedTicketServiceServer
	svc model.TicketService
}

func NewRPCHandler(svc model.TicketService) *RPCHandler {
	return &RPCHandler{svc: svc}
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
