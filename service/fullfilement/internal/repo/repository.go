package repo

import (
	"context"
	"fmt"
	"ticket-tix/service/fullfilement/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CancelledStatus = "cancelled"
)

type mongoFulfillmentRepository struct {
	collection *mongo.Collection
}

func NewRepository(collection *mongo.Collection) model.FulfillmentRepo {
	return &mongoFulfillmentRepository{
		collection: collection,
	}
}

func (m *mongoFulfillmentRepository) CreateFulfillment(ctx context.Context, fulfillment model.Booking) error {
	_, err := m.collection.InsertOne(ctx, fulfillment)
	if err != nil {
		fmt.Println("failed to insert fulfillment:", err)
		return err
	}
	return nil
}

func (m *mongoFulfillmentRepository) CancelFulfillment(ctx context.Context, bookingID string) error {
	filter := bson.M{
		"booking_id": bookingID,
	}

	update := bson.M{
		"$set": bson.M{
			"status": CancelledStatus,
		},
	}

	result, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("failed to cancel fulfillment:", err)
		return err
	}

	fmt.Println("cancelled fulfillment count:", result.ModifiedCount)
	return nil
}
