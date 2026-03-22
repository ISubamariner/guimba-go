package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateUserUseCase handles updating user profile fields.
type UpdateUserUseCase struct {
	repo repository.UserRepository
}

// NewUpdateUserUseCase creates a new UpdateUserUseCase.
func NewUpdateUserUseCase(repo repository.UserRepository) *UpdateUserUseCase {
	return &UpdateUserUseCase{repo: repo}
}

// Execute updates a user's profile fields (not password or roles).
func (uc *UpdateUserUseCase) Execute(ctx context.Context, id uuid.UUID, fullName string, isActive bool) (*entity.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NewNotFound("User", id)
	}

	user.FullName = fullName
	user.IsActive = isActive

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
