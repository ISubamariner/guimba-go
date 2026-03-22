package auth

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

// RegisterUseCase handles user registration.
type RegisterUseCase struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
	jwt      *auth.JWTManager
}

// NewRegisterUseCase creates a new RegisterUseCase.
func NewRegisterUseCase(userRepo repository.UserRepository, roleRepo repository.RoleRepository, jwt *auth.JWTManager) *RegisterUseCase {
	return &RegisterUseCase{userRepo: userRepo, roleRepo: roleRepo, jwt: jwt}
}

// Execute registers a new user, assigns the default "viewer" role, and returns a token pair.
func (uc *RegisterUseCase) Execute(ctx context.Context, email, fullName, password string) (*entity.User, *auth.TokenPair, error) {
	// Check if email already taken
	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, apperror.NewConflict("A user with this email already exists")
	}

	// Hash password
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return nil, nil, apperror.NewInternal(err)
	}

	// Create user
	user, err := entity.NewUser(email, fullName, hashed)
	if err != nil {
		return nil, nil, err
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, nil, err
	}

	// Assign default role
	viewerRole, err := uc.roleRepo.GetByName(ctx, "viewer")
	if err == nil && viewerRole != nil {
		_ = uc.userRepo.AssignRole(ctx, user.ID, viewerRole.ID)
		user.Roles = []entity.Role{*viewerRole}
	}

	// Generate tokens
	roleNames := extractRoleNames(user.Roles)
	tokens, err := uc.jwt.GenerateTokenPair(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, nil, apperror.NewInternal(err)
	}

	return user, tokens, nil
}

func extractRoleNames(roles []entity.Role) []string {
	names := make([]string, 0, len(roles))
	for _, r := range roles {
		names = append(names, r.Name)
	}
	return names
}
