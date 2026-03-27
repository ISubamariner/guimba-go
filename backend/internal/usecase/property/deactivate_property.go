package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
