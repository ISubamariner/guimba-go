package auth

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// LoginUseCase handles user authentication.
type LoginUseCase struct {
	userRepo repository.UserRepository
	jwt      *auth.JWTManager
}

// NewLoginUseCase creates a new LoginUseCase.
func NewLoginUseCase(userRepo repository.UserRepository, jwt *auth.JWTManager) *LoginUseCase {
	return &LoginUseCase{userRepo: userRepo, jwt: jwt}
}

// Execute authenticates a user by email and password, returns user and token pair.
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*entity.User, *auth.TokenPair, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, apperror.NewUnauthorized(entity.ErrUserInvalidCredentials.Error())
	}

	if !auth.CheckPassword(password, user.HashedPassword) {
		return nil, nil, apperror.NewUnauthorized(entity.ErrUserInvalidCredentials.Error())
	}

	if !user.IsActive {
		return nil, nil, apperror.NewForbidden(entity.ErrUserNotActive.Error())
	}

	// Update last login
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	roleNames := extractRoleNames(user.Roles)
	tokens, err := uc.jwt.GenerateTokenPair(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, nil, apperror.NewInternal(err)
	}

	return user, tokens, nil
}
