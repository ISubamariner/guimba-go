package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetTransactionUseCase struct {
	repo repository.TransactionRepository
}

func NewGetTransactionUseCase(repo repository.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{repo: repo}
}

func (uc *GetTransactionUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, apperror.NewNotFound("Transaction", id)
	}
	return tx, nil
}
