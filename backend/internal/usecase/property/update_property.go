package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdatePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewUpdatePropertyUseCase(repo repository.PropertyRepository) *UpdatePropertyUseCase {
	return &UpdatePropertyUseCase{repo: repo}
}

func (uc *UpdatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID, property *entity.Property) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	property.ID = id
	property.CreatedAt = existing.CreatedAt
	property.OwnerID = existing.OwnerID

	if err := property.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, property)
}
