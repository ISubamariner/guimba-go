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
}

func NewRecordPaymentUseCase(txRepo repository.TransactionRepository, debtRepo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository) *RecordPaymentUseCase {
	return &RecordPaymentUseCase{txRepo: txRepo, debtRepo: debtRepo, userRepo: userRepo, tenantRepo: tenantRepo}
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
	if err := d.RecordPayment(amount); err != nil {
		return nil, err
	}

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

	return tx, nil
}
