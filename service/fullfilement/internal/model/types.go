package model

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Booking struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	BookingID string             `bson:"booking_id"`
	UserID    string             `bson:"user_id"`
	SeatID    string             `bson:"seat_id"`
	Category  string             `bson:"category"`
	Status    string             `bson:"status"`
}

type FulfillmentRepo interface {
	CreateFulfillment(ctx context.Context, fulfillment Booking) error
	CancelFulfillment(ctx context.Context, bookingID string) error
}

type FulfillmentService interface {
	InsertFulfillment(ctx context.Context, booking Booking) error
	UpdateCancelledFulfillment(ctx context.Context, bookingID string) error
}
