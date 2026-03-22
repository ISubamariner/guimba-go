package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client and verifies connectivity.
func NewRedisClient(ctx context.Context, addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
		PoolSize: 25,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("pinging Redis: %w", err)
	}

	slog.Info("connected to Redis", "addr", addr)
	return client, nil
}

// CloseRedis gracefully closes the Redis client.
func CloseRedis(client *redis.Client) {
	if client != nil {
		if err := client.Close(); err != nil {
			slog.Error("failed to close Redis", "error", err)
			return
		}
		slog.Info("Redis connection closed")
	}
}
