package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetDebtUseCase struct {
	repo repository.DebtRepository
}

func NewGetDebtUseCase(repo repository.DebtRepository) *GetDebtUseCase {
	return &GetDebtUseCase{repo: repo}
}

func (uc *GetDebtUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", id)
	}

	// Lazy overdue detection
	if d.IsOverdue() && d.Status != entity.DebtStatusOverdue {
		d.MarkAsOverdue()
		_ = uc.repo.Update(ctx, d)
	}

	return d, nil
}
