package debt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdateDebtUseCase struct {
	repo repository.DebtRepository
}

func NewUpdateDebtUseCase(repo repository.DebtRepository) *UpdateDebtUseCase {
	return &UpdateDebtUseCase{repo: repo}
}

func (uc *UpdateDebtUseCase) Execute(ctx context.Context, id uuid.UUID, updates *entity.Debt) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	// Apply mutable fields only
	existing.Description = updates.Description
	existing.DebtType = updates.DebtType
	existing.DueDate = updates.DueDate
	existing.PropertyID = updates.PropertyID
	existing.Notes = updates.Notes
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, existing)
}
