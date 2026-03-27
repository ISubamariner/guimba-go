package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transaction represents an immutable financial record (payment or refund).
type Transaction struct {
	ID               uuid.UUID       `json:"id"`
	DebtID           uuid.UUID       `json:"debt_id"`
	TenantID         uuid.UUID       `json:"tenant_id"`
	LandlordID       uuid.UUID       `json:"landlord_id"`
	RecordedByUserID *uuid.UUID      `json:"recorded_by_user_id,omitempty"`
	TransactionType  TransactionType `json:"transaction_type"`
	Amount           Money           `json:"amount"`
	PaymentMethod    PaymentMethod   `json:"payment_method"`
	TransactionDate  time.Time       `json:"transaction_date"`
	Description      string          `json:"description"`
	ReceiptNumber    *string         `json:"receipt_number,omitempty"`
	ReferenceNumber  *string         `json:"reference_number,omitempty"`
	IsVerified       bool            `json:"is_verified"`
	VerifiedByUserID *uuid.UUID      `json:"verified_by_user_id,omitempty"`
	VerifiedAt       *time.Time      `json:"verified_at,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// NewTransaction creates a new Transaction with generated ID.
func NewTransaction(debtID, tenantID, landlordID uuid.UUID, recordedBy *uuid.UUID, txType TransactionType, amount Money, method PaymentMethod, txDate time.Time, description string, receipt, reference *string) (*Transaction, error) {
	tx := &Transaction{
		ID:               uuid.New(),
		DebtID:           debtID,
		TenantID:         tenantID,
		LandlordID:       landlordID,
		RecordedByUserID: recordedBy,
		TransactionType:  txType,
		Amount:           amount,
		PaymentMethod:    method,
		TransactionDate:  txDate,
		Description:      description,
		ReceiptNumber:    receipt,
		ReferenceNumber:  reference,
		IsVerified:       false,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	return tx, nil
}

// Validate checks business rules for a Transaction.
func (tx *Transaction) Validate() error {
	if tx.Amount.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrTransactionAmountRequired
	}
	if !tx.TransactionType.IsValid() {
		return ErrTransactionInvalidType
	}
	if !tx.PaymentMethod.IsValid() {
		return ErrTransactionInvalidPaymentMethod
	}
	if tx.TransactionDate.IsZero() {
		return ErrTransactionDateRequired
	}
	return nil
}

// Verify marks the transaction as verified by a specific user.
func (tx *Transaction) Verify(userID uuid.UUID) error {
	if tx.IsVerified {
		return ErrTransactionAlreadyVerified
	}
	tx.IsVerified = true
	tx.VerifiedByUserID = &userID
	now := time.Now().UTC()
	tx.VerifiedAt = &now
	tx.UpdatedAt = now
	return nil
}
