package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type VerifyTransactionUseCase struct {
	repo repository.TransactionRepository
}

func NewVerifyTransactionUseCase(repo repository.TransactionRepository) *VerifyTransactionUseCase {
	return &VerifyTransactionUseCase{repo: repo}
}

func (uc *VerifyTransactionUseCase) Execute(ctx context.Context, id, verifierID uuid.UUID) error {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tx == nil {
		return apperror.NewNotFound("Transaction", id)
	}

	if err := tx.Verify(verifierID); err != nil {
		return err
	}

	return uc.repo.Update(ctx, tx)
}
