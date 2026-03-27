package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo     repository.PropertyRepository
	debtRepo repository.DebtRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository, debtRepo repository.DebtRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo, debtRepo: debtRepo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	hasDebts, err := uc.debtRepo.HasActiveDebtsForProperty(ctx, id)
	if err != nil {
		return err
	}
	if hasDebts {
		return apperror.NewConflict(entity.ErrPropertyHasActiveDebts.Error())
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
