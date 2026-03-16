package lock

import (
	"context"
	"fmt"
	"time"
)

type DistributedLock interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (token string, err error)
	Release(ctx context.Context, key string, token string) error
}

var ErrLockNotAcquired = fmt.Errorf("lock not acquired")
