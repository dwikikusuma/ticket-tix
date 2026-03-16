package lock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SetNXLock struct {
	client *redis.Client
}

func NewSetNXLock(client *redis.Client) DistributedLock {
	return &SetNXLock{client: client}
}

func (s *SetNXLock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := uuid.New().String()

	err := s.client.SetArgs(ctx, key, token, redis.SetArgs{
		TTL:  ttl,
		Mode: "NX",
	}).Err()

	if errors.Is(err, redis.Nil) {
		return "", ErrLockNotAcquired
	}
	if err != nil {
		return "", fmt.Errorf("acquire lock: %w", err)
	}

	return token, nil
}
func (s *SetNXLock) Release(ctx context.Context, key string, token string) error {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, s.client, []string{key}, token).Result()
	if err != nil {
		log.Println("failed to release lock:", err)
		return err
	}

	if result == 0 {
		return fmt.Errorf("lock already released or expired")
	}
	return nil
}
