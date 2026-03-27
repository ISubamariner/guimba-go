package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type MarkDebtPaidUseCase struct {
	repo repository.DebtRepository
}

func NewMarkDebtPaidUseCase(repo repository.DebtRepository) *MarkDebtPaidUseCase {
	return &MarkDebtPaidUseCase{repo: repo}
}

func (uc *MarkDebtPaidUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return apperror.NewNotFound("Debt", id)
	}

	balance := d.GetBalance()
	if err := d.RecordPayment(balance); err != nil {
		return err
	}

	return uc.repo.Update(ctx, d)
}
