package user

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ListUsersUseCase handles listing users with filtering and pagination.
type ListUsersUseCase struct {
	repo repository.UserRepository
}

// NewListUsersUseCase creates a new ListUsersUseCase.
func NewListUsersUseCase(repo repository.UserRepository) *ListUsersUseCase {
	return &ListUsersUseCase{repo: repo}
}

// Execute returns a filtered, paginated list of users and the total count.
func (uc *ListUsersUseCase) Execute(ctx context.Context, filter repository.UserFilter) ([]*entity.User, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.repo.List(ctx, filter)
}
