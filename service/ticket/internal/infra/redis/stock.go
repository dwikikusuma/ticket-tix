package redis

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type StockCounter struct {
	client *redis.Client
}

func NewStockCounter(client *redis.Client) StockCounter {
	return StockCounter{client: client}
}

func stockKey(evntID int32) string {
	return fmt.Sprintf("stock:%d", evntID)
}

func (s *StockCounter) Seed(ctx context.Context, eventID int32, stock int64) error {
	key := stockKey(eventID)
	err := s.client.SetArgs(ctx, key, stock, redis.SetArgs{
		Mode: "NX",
		Get:  false,
	}).Err()

	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("seed stock: %w", err)
	}
	return nil
}

func (s *StockCounter) Decrement(ctx context.Context, eventID int32, val int64) error {
	key := stockKey(eventID)
	remaining, err := s.client.DecrBy(ctx, key, val).Result()

	if err != nil {
		log.Printf("failed to decrement stock for event %d: %v\n", eventID, err)
		return fmt.Errorf("decrement stock (%d): %w", remaining, err)
	}

	if remaining < 0 {
		s.client.IncrBy(ctx, key, val)
		log.Printf("stock insufficient for event %d, rolling back decrement. Remaining stock: %d\n", eventID, remaining)
		return fmt.Errorf("stock not sufficient: (%d)", remaining)
	}

	return nil
}

func (s *StockCounter) Increment(ctx context.Context, eventID int32, val int64) error {
	key := stockKey(eventID)
	_, err := s.client.IncrBy(ctx, key, val).Result()
	if err != nil {
		return fmt.Errorf("increment stock: %w", err)
	}
	return nil
}

func (s *StockCounter) Get(ctx context.Context, eventID int32) (int64, error) {
	key := stockKey(eventID)
	stock, err := s.client.Get(ctx, key).Int64()
	if err != nil {
		return 0, fmt.Errorf("get stock: %w", err)
	}
	return stock, nil
}
