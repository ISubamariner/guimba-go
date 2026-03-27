package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type RecordRefundUseCase struct {
	txRepo     repository.TransactionRepository
	debtRepo   repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
}

func NewRecordRefundUseCase(txRepo repository.TransactionRepository, debtRepo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository) *RecordRefundUseCase {
	return &RecordRefundUseCase{txRepo: txRepo, debtRepo: debtRepo, userRepo: userRepo, tenantRepo: tenantRepo}
}

func (uc *RecordRefundUseCase) Execute(ctx context.Context, debtID, tenantID uuid.UUID, recordedBy *uuid.UUID, amount entity.Money, method entity.PaymentMethod, refundDate time.Time, description string, reference *string) (*entity.Transaction, error) {
	// Validate debt exists
	d, err := uc.debtRepo.GetByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", debtID)
	}

	// Reverse payment on debt (validates amount <= paid, currency)
	if err := d.ReversePayment(amount); err != nil {
		return nil, err
	}

	// Create refund transaction record
	tx, err := entity.NewTransaction(debtID, d.TenantID, d.LandlordID, recordedBy, entity.TransactionTypeRefund, amount, method, refundDate, description, nil, reference)
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
