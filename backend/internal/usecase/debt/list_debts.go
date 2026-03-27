package debt

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListDebtsUseCase struct {
	repo repository.DebtRepository
}

func NewListDebtsUseCase(repo repository.DebtRepository) *ListDebtsUseCase {
	return &ListDebtsUseCase{repo: repo}
}

func (uc *ListDebtsUseCase) Execute(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	debts, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Lazy overdue detection on all results
	for _, d := range debts {
		if d.IsOverdue() && d.Status != entity.DebtStatusOverdue {
			d.MarkAsOverdue()
			_ = uc.repo.Update(ctx, d)
		}
	}

	return debts, total, nil
}
