package lock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type redLock struct {
	clients []*redis.Client
	quorum  int
}

func NewRedLock(clients ...*redis.Client) DistributedLock {
	return &redLock{
		clients: clients,
		quorum:  len(clients)/2 + 1,
	}
}

func (r *redLock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := uuid.New().String()
	startTime := time.Now()

	acquired := 0
	for _, client := range r.clients {
		err := client.SetArgs(ctx, key, token, redis.SetArgs{
			TTL:  ttl,
			Mode: "NX",
		}).Err()

		if errors.Is(err, redis.Nil) {
			// key already exists, skip to next client
			continue
		}

		if err != nil {
			// this Node failed to acquire lock, skip to next client
			continue
		}

		acquired++
	}

	elapsed := time.Since(startTime)
	drift := time.Duration(float64(ttl)*0.01) + 2*time.Millisecond
	validity := ttl - elapsed - drift

	if acquired >= r.quorum && validity > 0 {
		return token, nil
	}
	return "", errors.New("failed to acquire lock")
}

func (r *redLock) Release(ctx context.Context, key string, token string) error {
	r.releaseAll(ctx, key, token)
	return nil
}

func (r *redLock) releaseAll(ctx context.Context, key string, token string) {
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)

	for _, client := range r.clients {
		script.Run(ctx, client, []string{key}, token)
	}
}
