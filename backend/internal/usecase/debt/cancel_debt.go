package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CancelDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewCancelDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *CancelDebtUseCase {
	return &CancelDebtUseCase{repo: repo, auditRepo: auditRepo}
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

	if err := uc.repo.Update(ctx, d); err != nil {
		return err
	}

	metadata := map[string]any{"landlord_id": d.LandlordID.String(), "tenant_id": d.TenantID.String()}
	if reason != nil {
		metadata["reason"] = *reason
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CANCEL_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     metadata,
	})

	return nil
}
