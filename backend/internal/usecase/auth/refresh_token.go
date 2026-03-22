package auth

import (
	"context"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/cache"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// RefreshTokenUseCase handles token refresh.
type RefreshTokenUseCase struct {
	userRepo  repository.UserRepository
	jwt       *auth.JWTManager
	blocklist *cache.TokenBlocklist
}

// NewRefreshTokenUseCase creates a new RefreshTokenUseCase.
func NewRefreshTokenUseCase(userRepo repository.UserRepository, jwt *auth.JWTManager, blocklist *cache.TokenBlocklist) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{userRepo: userRepo, jwt: jwt, blocklist: blocklist}
}

// Execute validates the refresh token, blocks the old one, and issues a new pair.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	claims, err := uc.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, apperror.NewUnauthorized("Invalid or expired refresh token")
	}

	// Check blocklist
	blocked, err := uc.blocklist.IsBlocked(ctx, claims.ID)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}
	if blocked {
		return nil, apperror.NewUnauthorized("Token has been revoked")
	}

	// Block old refresh token (token rotation)
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining > 0 {
		_ = uc.blocklist.Block(ctx, claims.ID, remaining)
	}

	// Verify user still exists and is active
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, apperror.NewUnauthorized("User not found or inactive")
	}

	// Generate new pair
	roleNames := extractRoleNames(user.Roles)
	tokens, err := uc.jwt.GenerateTokenPair(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, apperror.NewInternal(err)
	}

	return tokens, nil
}
