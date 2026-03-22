package mocks

import (
	"context"
	"time"
)

// TokenBlocklistMock is a test mock for cache.TokenBlocklist.
type TokenBlocklistMock struct {
	BlockFn     func(ctx context.Context, jti string, expiration time.Duration) error
	IsBlockedFn func(ctx context.Context, jti string) (bool, error)
}

func (m *TokenBlocklistMock) Block(ctx context.Context, jti string, expiration time.Duration) error {
	if m.BlockFn != nil {
		return m.BlockFn(ctx, jti, expiration)
	}
	return nil
}

func (m *TokenBlocklistMock) IsBlocked(ctx context.Context, jti string) (bool, error) {
	if m.IsBlockedFn != nil {
		return m.IsBlockedFn(ctx, jti)
	}
	return false, nil
}
