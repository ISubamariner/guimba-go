package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// GetProfileUseCase handles retrieving the current user's profile.
type GetProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewGetProfileUseCase creates a new GetProfileUseCase.
func NewGetProfileUseCase(userRepo repository.UserRepository) *GetProfileUseCase {
	return &GetProfileUseCase{userRepo: userRepo}
}

// Execute retrieves a user profile by ID.
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NewNotFound("User", userID)
	}
	return user, nil
}
