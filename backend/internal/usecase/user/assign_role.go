package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// AssignRoleUseCase handles assigning a role to a user.
type AssignRoleUseCase struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewAssignRoleUseCase creates a new AssignRoleUseCase.
func NewAssignRoleUseCase(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *AssignRoleUseCase {
	return &AssignRoleUseCase{userRepo: userRepo, roleRepo: roleRepo}
}

// Execute assigns a role to a user.
func (uc *AssignRoleUseCase) Execute(ctx context.Context, userID, roleID uuid.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User", userID)
	}

	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return apperror.NewNotFound("Role", roleID)
	}

	return uc.userRepo.AssignRole(ctx, userID, roleID)
}

// RemoveRoleUseCase handles removing a role from a user.
type RemoveRoleUseCase struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

// NewRemoveRoleUseCase creates a new RemoveRoleUseCase.
func NewRemoveRoleUseCase(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{userRepo: userRepo, roleRepo: roleRepo}
}

// Execute removes a role from a user.
func (uc *RemoveRoleUseCase) Execute(ctx context.Context, userID, roleID uuid.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User", userID)
	}

	return uc.userRepo.RemoveRole(ctx, userID, roleID)
}
