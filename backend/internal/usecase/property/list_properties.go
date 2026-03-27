package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListPropertiesUseCase struct {
	repo repository.PropertyRepository
}

func NewListPropertiesUseCase(repo repository.PropertyRepository) *ListPropertiesUseCase {
	return &ListPropertiesUseCase{repo: repo}
}

func (uc *ListPropertiesUseCase) Execute(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
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
