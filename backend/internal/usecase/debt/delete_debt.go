package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeleteDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewDeleteDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *DeleteDebtUseCase {
	return &DeleteDebtUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeleteDebtUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": existing.LandlordID.String(), "tenant_id": existing.TenantID.String()},
	})

	return nil
}
