package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeletePropertyUseCase struct {
	repo repository.PropertyRepository
}

func NewDeletePropertyUseCase(repo repository.PropertyRepository) *DeletePropertyUseCase {
	return &DeletePropertyUseCase{repo: repo}
}

func (uc *DeletePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	return uc.repo.Delete(ctx, id)
}
