package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CancelDebtUseCase struct {
	repo repository.DebtRepository
}

func NewCancelDebtUseCase(repo repository.DebtRepository) *CancelDebtUseCase {
	return &CancelDebtUseCase{repo: repo}
}

func (uc *CancelDebtUseCase) Execute(ctx context.Context, id uuid.UUID, reason *string) error {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return apperror.NewNotFound("Debt", id)
	}

	if err := d.Cancel(reason); err != nil {
		return err
	}

	return uc.repo.Update(ctx, d)
}
