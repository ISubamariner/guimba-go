package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlocklist manages blocked JWT tokens in Redis.
type TokenBlocklist struct {
	client *redis.Client
}

// NewTokenBlocklist creates a new Redis-backed token blocklist.
func NewTokenBlocklist(client *redis.Client) *TokenBlocklist {
	return &TokenBlocklist{client: client}
}

// Block adds a token JTI to the blocklist with a TTL matching the token's remaining lifetime.
func (b *TokenBlocklist) Block(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("blocklist:%s", jti)
	return b.client.Set(ctx, key, "1", expiration).Err()
}

// IsBlocked checks whether a token JTI has been blocklisted.
func (b *TokenBlocklist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blocklist:%s", jti)
	result, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
