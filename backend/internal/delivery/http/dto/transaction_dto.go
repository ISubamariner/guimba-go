package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// RecordPaymentRequest is the request body for recording a payment transaction.
type RecordPaymentRequest struct {
	DebtID          uuid.UUID `json:"debt_id" validate:"required"`
	TenantID        uuid.UUID `json:"tenant_id" validate:"required"`
	Amount          MoneyDTO  `json:"amount" validate:"required"`
	PaymentMethod   string    `json:"payment_method" validate:"required"`
	TransactionDate string    `json:"transaction_date" validate:"required"`
	Description     string    `json:"description" validate:"required"`
	ReceiptNumber   *string   `json:"receipt_number" validate:"omitempty"`
	ReferenceNumber *string   `json:"reference_number" validate:"omitempty"`
}

// RecordRefundRequest is the request body for recording a refund transaction.
type RecordRefundRequest struct {
	DebtID          uuid.UUID `json:"debt_id" validate:"required"`
	TenantID        uuid.UUID `json:"tenant_id" validate:"required"`
	Amount          MoneyDTO  `json:"amount" validate:"required"`
	PaymentMethod   string    `json:"payment_method" validate:"required"`
	RefundDate      string    `json:"refund_date" validate:"required"`
	Description     string    `json:"description" validate:"required"`
	ReferenceNumber *string   `json:"reference_number" validate:"omitempty"`
}

// TransactionResponse is the response body for a single transaction.
type TransactionResponse struct {
	ID               uuid.UUID  `json:"id"`
	DebtID           uuid.UUID  `json:"debt_id"`
	TenantID         uuid.UUID  `json:"tenant_id"`
	LandlordID       uuid.UUID  `json:"landlord_id"`
	RecordedByUserID *uuid.UUID `json:"recorded_by_user_id,omitempty"`
	TransactionType  string     `json:"transaction_type"`
	Amount           MoneyDTO   `json:"amount"`
	PaymentMethod    string     `json:"payment_method"`
	TransactionDate  string     `json:"transaction_date"`
	Description      string     `json:"description"`
	ReceiptNumber    *string    `json:"receipt_number,omitempty"`
	ReferenceNumber  *string    `json:"reference_number,omitempty"`
	IsVerified       bool       `json:"is_verified"`
	VerifiedByUserID *uuid.UUID `json:"verified_by_user_id,omitempty"`
	VerifiedAt       *string    `json:"verified_at,omitempty"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

// TransactionListResponse is the response body for a list of transactions.
type TransactionListResponse struct {
	Data   []TransactionResponse `json:"data"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// NewTransactionResponse creates a TransactionResponse from a domain Transaction entity.
func NewTransactionResponse(tx *entity.Transaction) TransactionResponse {
	resp := TransactionResponse{
		ID:               tx.ID,
		DebtID:           tx.DebtID,
		TenantID:         tx.TenantID,
		LandlordID:       tx.LandlordID,
		RecordedByUserID: tx.RecordedByUserID,
		TransactionType:  string(tx.TransactionType),
		Amount:           NewMoneyDTO(tx.Amount),
		PaymentMethod:    string(tx.PaymentMethod),
		TransactionDate:  tx.TransactionDate.Format(time.RFC3339),
		Description:      tx.Description,
		ReceiptNumber:    tx.ReceiptNumber,
		ReferenceNumber:  tx.ReferenceNumber,
		IsVerified:       tx.IsVerified,
		VerifiedByUserID: tx.VerifiedByUserID,
		CreatedAt:        tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        tx.UpdatedAt.Format(time.RFC3339),
	}
	if tx.VerifiedAt != nil {
		v := tx.VerifiedAt.Format(time.RFC3339)
		resp.VerifiedAt = &v
	}
	return resp
}

// NewTransactionListResponse creates a TransactionListResponse from domain entities.
func NewTransactionListResponse(txs []*entity.Transaction, total, limit, offset int) TransactionListResponse {
	data := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		data = append(data, NewTransactionResponse(tx))
	}
	return TransactionListResponse{Data: data, Total: total, Limit: limit, Offset: offset}
}
