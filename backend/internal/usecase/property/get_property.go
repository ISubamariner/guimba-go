package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetPropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewGetPropertyUseCase(repo repository.PropertyRepository) *GetPropertyUseCase {
	return &GetPropertyUseCase{repo: repo}
}

func (uc *GetPropertyUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	property, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if property == nil {
		return nil, apperror.NewNotFound("Property", id)
	}
	return property, nil
}
