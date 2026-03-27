package transaction

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListTransactionsUseCase struct {
	repo repository.TransactionRepository
}

func NewListTransactionsUseCase(repo repository.TransactionRepository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{repo: repo}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return uc.repo.List(ctx, filter)
}
