package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// MoneyDTO represents a monetary amount for API transport.
type MoneyDTO struct {
	Amount   string `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required,len=3"`
}

// ToMoney converts a MoneyDTO to a domain Money value.
func ToMoney(dto MoneyDTO) (entity.Money, error) {
	amount, err := decimal.NewFromString(dto.Amount)
	if err != nil {
		return entity.Money{}, fmt.Errorf("invalid amount format: %w", err)
	}
	return entity.NewMoney(amount, entity.Currency(dto.Currency))
}

// NewMoneyDTO converts a domain Money value to a MoneyDTO.
func NewMoneyDTO(m entity.Money) MoneyDTO {
	return MoneyDTO{
		Amount:   m.Amount.StringFixed(2),
		Currency: string(m.Currency),
	}
}

// CreateDebtRequest is the request body for creating a debt.
type CreateDebtRequest struct {
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required"`
	PropertyID     *uuid.UUID `json:"property_id" validate:"omitempty"`
	DebtType       string     `json:"debt_type" validate:"required"`
	Description    string     `json:"description" validate:"required,max=500"`
	OriginalAmount MoneyDTO   `json:"original_amount" validate:"required"`
	DueDate        string     `json:"due_date" validate:"required"`
	Notes          *string    `json:"notes" validate:"omitempty"`
}

// UpdateDebtRequest is the request body for updating a debt.
type UpdateDebtRequest struct {
	Description string     `json:"description" validate:"required,max=500"`
	DebtType    string     `json:"debt_type" validate:"required"`
	DueDate     string     `json:"due_date" validate:"required"`
	PropertyID  *uuid.UUID `json:"property_id" validate:"omitempty"`
	Notes       *string    `json:"notes" validate:"omitempty"`
}

// CancelDebtRequest is the request body for cancelling a debt.
type CancelDebtRequest struct {
	Reason *string `json:"reason" validate:"omitempty"`
}

// DebtResponse is the response body for a single debt.
type DebtResponse struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	LandlordID     uuid.UUID  `json:"landlord_id"`
	PropertyID     *uuid.UUID `json:"property_id,omitempty"`
	DebtType       string     `json:"debt_type"`
	Description    string     `json:"description"`
	OriginalAmount MoneyDTO   `json:"original_amount"`
	AmountPaid     MoneyDTO   `json:"amount_paid"`
	Balance        MoneyDTO   `json:"balance"`
	DueDate        string     `json:"due_date"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

// DebtListResponse is the response body for a list of debts.
type DebtListResponse struct {
	Data   []DebtResponse `json:"data"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// NewDebtResponse creates a DebtResponse from a domain Debt entity.
func NewDebtResponse(d *entity.Debt) DebtResponse {
	return DebtResponse{
		ID:             d.ID,
		TenantID:       d.TenantID,
		LandlordID:     d.LandlordID,
		PropertyID:     d.PropertyID,
		DebtType:       string(d.DebtType),
		Description:    d.Description,
		OriginalAmount: NewMoneyDTO(d.OriginalAmount),
		AmountPaid:     NewMoneyDTO(d.AmountPaid),
		Balance:        NewMoneyDTO(d.GetBalance()),
		DueDate:        d.DueDate.Format("2006-01-02"),
		Status:         string(d.Status),
		Notes:          d.Notes,
		CreatedAt:      d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      d.UpdatedAt.Format(time.RFC3339),
	}
}

// NewDebtListResponse creates a DebtListResponse from domain entities.
func NewDebtListResponse(debts []*entity.Debt, total, limit, offset int) DebtListResponse {
	data := make([]DebtResponse, 0, len(debts))
	for _, d := range debts {
		data = append(data, NewDebtResponse(d))
	}
	return DebtListResponse{Data: data, Total: total, Limit: limit, Offset: offset}
}
