package events

import (
	"context"
	"log/slog"
	"time"
)

func WithLogging(logger *slog.Logger) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg Message) error {
			start := time.Now()

			err := next(ctx, msg)

			duration := time.Since(start).Milliseconds()

			if err != nil {
				logger.ErrorContext(ctx, "message processing failed",
					"topic", msg.Topic,
					"key", string(msg.Key),
					"duration_ms", duration,
					"error", err.Error(),
				)
			} else {
				logger.InfoContext(ctx, "message processed",
					"topic", msg.Topic,
					"key", string(msg.Key),
					"duration_ms", duration,
				)
			}
			return err
		}
	}
}

func WithRetry(maxRetries int, backoff time.Duration) Middleware {
	return func(next MessageHandler) MessageHandler {
		return func(ctx context.Context, msg Message) error {
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				if attempt > 0 {
					waitDuration := time.Duration(attempt) * backoff

					select {
					case <-time.After(waitDuration):
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				lastErr = next(ctx, msg)
				if lastErr == nil {
					return nil
				}
			}

			return lastErr
		}
	}
}
