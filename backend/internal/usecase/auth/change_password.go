package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// ChangePasswordUseCase handles user password change.
type ChangePasswordUseCase struct {
	userRepo repository.UserRepository
}

// NewChangePasswordUseCase creates a new ChangePasswordUseCase.
func NewChangePasswordUseCase(userRepo repository.UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{userRepo: userRepo}
}

// Execute changes the user's password after validating the current password.
func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFoundMsg("User not found")
	}

	if !auth.CheckPassword(currentPassword, user.HashedPassword) {
		return apperror.NewValidation("Current password is incorrect")
	}

	if auth.CheckPassword(newPassword, user.HashedPassword) {
		return apperror.NewValidation("New password must be different from current password")
	}

	hashed, err := auth.HashPassword(newPassword)
	if err != nil {
		return apperror.NewInternal(err)
	}

	return uc.userRepo.UpdatePassword(ctx, userID, hashed)
}
