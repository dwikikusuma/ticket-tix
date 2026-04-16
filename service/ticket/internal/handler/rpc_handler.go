package handler

import (
	"context"
	"strings"
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

func (h *RPCHandler) ReserveTicket(ctx context.Context, req *ticketRPC.ReserveTicketRequest) (*ticketRPC.ReserveTicketResponse, error) {
	if req.SeatId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "seat_id is required")
	}

	ticketID, err := h.svc.ReserveTicket(ctx, req.SeatId, req.EventCategory)
	if err != nil {
		// map specific errors to correct gRPC codes
		if strings.Contains(err.Error(), "not available") {
			return nil, status.Errorf(codes.FailedPrecondition, "seat is not available: %v", err)
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "seat not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "reserve ticket: %v", err)
	}

	return &ticketRPC.ReserveTicketResponse{
		TicketId:   ticketID,
		SeatNumber: req.SeatId,
	}, nil
}

func (h *RPCHandler) ReleaseTicket(ctx context.Context, req *ticketRPC.ReleaseTicketRequest) (*ticketRPC.ReleaseTicketResponse, error) {
	if req.SeatId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "seat_id is required")
	}

	if err := h.svc.ReleaseTicket(ctx, req.SeatId, req.EventCategory); err != nil {
		if strings.Contains(err.Error(), "not reserved") {
			return nil, status.Errorf(codes.FailedPrecondition, "seat is not reserved: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "release ticket: %v", err)
	}

	return &ticketRPC.ReleaseTicketResponse{}, nil
}

func (h *RPCHandler) ValidateTicket(ctx context.Context, req *ticketRPC.ValidateTicketRequest) (*ticketRPC.ValidateTicketResponse, error) {
	if err := h.svc.ValidateTicketBooking(ctx, req.SeatId, req.EventId, req.EventCategory); err != nil {
		if strings.Contains(err.Error(), "not available") {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}

		if strings.Contains(err.Error(), "required") {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}

		if strings.Contains(err.Error(), "no available capacity") {
			return nil, status.Errorf(codes.ResourceExhausted, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return &ticketRPC.ValidateTicketResponse{IsValid: true}, nil
}
func (h *RPCHandler) ReserveSeat(ctx context.Context, req *ticketRPC.ReserveFlexibleSeatRequest) (*ticketRPC.ReserveFlexibleSeatResponse, error) {
	seatNum, tixID, err := h.svc.ReserveAvailableSeat(ctx, req.GetEventCategoryId())
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "failed to reserve available seat: %v", err)
	}
	return &ticketRPC.ReserveFlexibleSeatResponse{
		TicketId:   tixID,
		SeatNumber: seatNum,
	}, nil
}

func (h *RPCHandler) DecreaseTicket(ctx context.Context, req *ticketRPC.DecreaseTicketRequest) (*ticketRPC.DecreaseTicketResponse, error) {
	eventCat := req.GetEventCategoryId()
	//decreaseBy := req.GetDecreaseBy()
	if err := h.stockCounter.Decrement(ctx, eventCat, 1); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "failed to decrease ticket stock: %v", err)
	}
	return &ticketRPC.DecreaseTicketResponse{}, nil
}

func (h *RPCHandler) IncreaseTicket(ctx context.Context, req *ticketRPC.IncreaseTicketRequest) (*ticketRPC.IncreaseTicketResponse, error) {
	eventCat := req.GetEventCategoryId()
	increaseBy := req.GetIncreaseBy()

	if err := h.stockCounter.Increment(ctx, eventCat, increaseBy); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "failed to increase ticket stock: %v", err)
	}
	return &ticketRPC.IncreaseTicketResponse{}, nil
}
