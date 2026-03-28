package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type VerifyTransactionUseCase struct {
	repo      repository.TransactionRepository
	auditRepo repository.AuditRepository
}

func NewVerifyTransactionUseCase(repo repository.TransactionRepository, auditRepo repository.AuditRepository) *VerifyTransactionUseCase {
	return &VerifyTransactionUseCase{repo: repo, auditRepo: auditRepo}
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

	if err := uc.repo.Update(ctx, tx); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "VERIFY_TRANSACTION",
		ResourceType: "Transaction",
		ResourceID:   id,
		Success:      true,
		Metadata: map[string]any{
			"verified_by_user_id": verifierID.String(),
			"transaction_type":   string(tx.TransactionType),
			"amount":             tx.Amount.Amount.String(),
			"currency":           string(tx.Amount.Currency),
			"landlord_id":        tx.LandlordID.String(),
			"tenant_id":          tx.TenantID.String(),
		},
	})

	return nil
}
