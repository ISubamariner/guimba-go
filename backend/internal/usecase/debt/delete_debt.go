package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeleteDebtUseCase struct {
	repo repository.DebtRepository
}

func NewDeleteDebtUseCase(repo repository.DebtRepository) *DeleteDebtUseCase {
	return &DeleteDebtUseCase{repo: repo}
}

func (uc *DeleteDebtUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	return uc.repo.Delete(ctx, id)
}
