package redis

import (
	"context"
	"errors"
	"fmt"
	"ticket-tix/common/pkg/jwt"

	"github.com/redis/go-redis/v9"
)

type RefreshToken struct {
	client *redis.Client
}

func NewRefreshToken(client *redis.Client) RefreshToken {
	return RefreshToken{client: client}
}

func (r *RefreshToken) generateRefreshTokenKey(userID int32, token string) string {
	return fmt.Sprintf("user:refresh-token:%d:%s", userID, token)
}

func (r *RefreshToken) generateAccessTokenKey(userID int32, token string) string {
	return fmt.Sprintf("user:access-token:%d:%s", userID, token)
}

func (r *RefreshToken) SaveRefreshToken(ctx context.Context, userID int32, token string) error {
	key := r.generateRefreshTokenKey(userID, token)
	return r.client.Set(ctx, key, true, jwt.RefreshTokenExpiry).Err()
}

func (r *RefreshToken) ValidateRefreshToken(ctx context.Context, userID int32, token string) (bool, error) {
	key := r.generateRefreshTokenKey(userID, token)
	tokenErr := r.client.Get(ctx, key).Err()

	if errors.Is(tokenErr, redis.Nil) {
		return false, nil
	} else if tokenErr != nil {
		return false, tokenErr
	}
	return true, nil
}

func (r *RefreshToken) RevokeUser(ctx context.Context, userID int32) error {
	pattern := fmt.Sprintf("user:refresh-token:%d:*", userID)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (r *RefreshToken) RevokeRefreshToken(ctx context.Context, userID int32, token string) error {
	key := r.generateRefreshTokenKey(userID, token)
	return r.client.Del(ctx, key).Err()
}

func (r *RefreshToken) BlackListAccessToken(ctx context.Context, token string) error {
	key := fmt.Sprintf("blacklist:access:%s", token)
	return r.client.Set(ctx, key, "1", jwt.AccessTokenExpiry).Err()
}
