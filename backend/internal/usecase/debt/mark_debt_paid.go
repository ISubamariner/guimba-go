package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type MarkDebtPaidUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewMarkDebtPaidUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *MarkDebtPaidUseCase {
	return &MarkDebtPaidUseCase{repo: repo, auditRepo: auditRepo}
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

	if err := uc.repo.Update(ctx, d); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "MARK_DEBT_PAID",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": d.LandlordID.String(), "tenant_id": d.TenantID.String(), "amount": d.OriginalAmount.Amount.String()},
	})

	return nil
}
