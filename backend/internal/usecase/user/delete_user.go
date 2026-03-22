package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteUserUseCase handles soft-deleting a user.
type DeleteUserUseCase struct {
	repo repository.UserRepository
}

// NewDeleteUserUseCase creates a new DeleteUserUseCase.
func NewDeleteUserUseCase(repo repository.UserRepository) *DeleteUserUseCase {
	return &DeleteUserUseCase{repo: repo}
}

// Execute soft-deletes a user by ID.
func (uc *DeleteUserUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User", id)
	}
	return uc.repo.Delete(ctx, id)
}
