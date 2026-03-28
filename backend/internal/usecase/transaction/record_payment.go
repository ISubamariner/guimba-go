package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type RecordPaymentUseCase struct {
	txRepo     repository.TransactionRepository
	debtRepo   repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	auditRepo  repository.AuditRepository
}

func NewRecordPaymentUseCase(txRepo repository.TransactionRepository, debtRepo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, auditRepo repository.AuditRepository) *RecordPaymentUseCase {
	return &RecordPaymentUseCase{txRepo: txRepo, debtRepo: debtRepo, userRepo: userRepo, tenantRepo: tenantRepo, auditRepo: auditRepo}
}

func (uc *RecordPaymentUseCase) Execute(ctx context.Context, debtID, tenantID uuid.UUID, recordedBy *uuid.UUID, amount entity.Money, method entity.PaymentMethod, txDate time.Time, description string, receipt, reference *string) (*entity.Transaction, error) {
	// Validate debt exists
	d, err := uc.debtRepo.GetByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", debtID)
	}

	// Check duplicate reference number
	if reference != nil {
		exists, err := uc.txRepo.ExistsByReferenceNumber(ctx, debtID, *reference)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, entity.ErrTransactionDuplicateReference
		}
	}

	// Record payment on debt (validates overpayment, currency, status)
	balanceBefore := d.GetBalance()
	if err := d.RecordPayment(amount); err != nil {
		return nil, err
	}
	balanceAfter := d.GetBalance()

	// Create transaction record
	tx, err := entity.NewTransaction(debtID, d.TenantID, d.LandlordID, recordedBy, entity.TransactionTypePayment, amount, method, txDate, description, receipt, reference)
	if err != nil {
		return nil, err
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if err := uc.debtRepo.Update(ctx, d); err != nil {
		return nil, err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "APPLY_PAYMENT",
		ResourceType: "Transaction",
		ResourceID:   tx.ID,
		Success:      true,
		Metadata: map[string]any{
			"landlord_id":    d.LandlordID.String(),
			"tenant_id":     d.TenantID.String(),
			"payment_amount": amount.Amount.String(),
			"currency":      string(amount.Currency),
			"balance_before": balanceBefore.Amount.String(),
			"balance_after":  balanceAfter.Amount.String(),
			"debt_type":     string(d.DebtType),
		},
	})

	return tx, nil
}
