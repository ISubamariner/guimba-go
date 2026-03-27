package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Debt represents money owed by a tenant to a landlord.
type Debt struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	LandlordID     uuid.UUID  `json:"landlord_id"`
	PropertyID     *uuid.UUID `json:"property_id,omitempty"`
	DebtType       DebtType   `json:"debt_type"`
	Description    string     `json:"description"`
	OriginalAmount Money      `json:"original_amount"`
	AmountPaid     Money      `json:"amount_paid"`
	DueDate        time.Time  `json:"due_date"`
	Status         DebtStatus `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewDebt creates a new Debt with generated ID, PENDING status, and zero AmountPaid.
func NewDebt(tenantID, landlordID uuid.UUID, propertyID *uuid.UUID, debtType DebtType, description string, originalAmount Money, dueDate time.Time, notes *string) (*Debt, error) {
	d := &Debt{
		ID:             uuid.New(),
		TenantID:       tenantID,
		LandlordID:     landlordID,
		PropertyID:     propertyID,
		DebtType:       debtType,
		Description:    description,
		OriginalAmount: originalAmount,
		AmountPaid:     ZeroMoney(originalAmount.Currency),
		DueDate:        dueDate,
		Status:         DebtStatusPending,
		Notes:          notes,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := d.Validate(); err != nil {
		return nil, err
	}

	return d, nil
}

// Validate checks business rules for a Debt.
func (d *Debt) Validate() error {
	if d.Description == "" {
		return ErrDebtDescriptionRequired
	}
	if len(d.Description) > 500 {
		return ErrDebtDescriptionTooLong
	}
	if d.OriginalAmount.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrDebtAmountRequired
	}
	if !d.DebtType.IsValid() {
		return ErrDebtInvalidType
	}
	if d.DueDate.IsZero() {
		return ErrDebtDueDateRequired
	}
	return nil
}

// RecordPayment adds a payment to the debt.
// Auto-transitions to PARTIAL or PAID based on remaining balance.
func (d *Debt) RecordPayment(amount Money) error {
	if d.Status == DebtStatusPaid {
		return ErrDebtAlreadyPaid
	}
	if d.Status == DebtStatusCancelled {
		return ErrDebtAlreadyCancelled
	}
	if d.OriginalAmount.Currency != amount.Currency {
		return ErrCurrencyMismatch
	}

	balance := d.GetBalance()
	gt, _ := amount.IsGreaterThan(balance)
	if gt {
		return ErrDebtOverpayment
	}

	newPaid, err := d.AmountPaid.Add(amount)
	if err != nil {
		return err
	}
	d.AmountPaid = newPaid

	if d.IsFullyPaid() {
		d.Status = DebtStatusPaid
	} else {
		d.Status = DebtStatusPartial
	}

	d.UpdatedAt = time.Now().UTC()
	return nil
}

// ReversePayment removes a payment amount from AmountPaid.
// Recalculates status: zero paid -> PENDING, else -> PARTIAL.
func (d *Debt) ReversePayment(amount Money) error {
	if d.AmountPaid.Currency != amount.Currency {
		return ErrCurrencyMismatch
	}

	newPaid, err := d.AmountPaid.Subtract(amount)
	if err != nil {
		return err
	}
	d.AmountPaid = newPaid

	if d.AmountPaid.IsZero() {
		d.Status = DebtStatusPending
	} else {
		d.Status = DebtStatusPartial
	}

	d.UpdatedAt = time.Now().UTC()
	return nil
}

// MarkAsOverdue transitions the debt to OVERDUE if currently PENDING or PARTIAL.
func (d *Debt) MarkAsOverdue() {
	if d.Status == DebtStatusPending || d.Status == DebtStatusPartial {
		d.Status = DebtStatusOverdue
		d.UpdatedAt = time.Now().UTC()
	}
}

// Cancel cancels the debt. Cannot cancel a PAID debt.
func (d *Debt) Cancel(reason *string) error {
	if d.Status == DebtStatusPaid {
		return ErrDebtAlreadyPaid
	}
	d.Status = DebtStatusCancelled
	if reason != nil {
		if d.Notes != nil {
			combined := *d.Notes + "; Cancelled: " + *reason
			d.Notes = &combined
		} else {
			note := "Cancelled: " + *reason
			d.Notes = &note
		}
	}
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// GetBalance returns OriginalAmount - AmountPaid.
func (d *Debt) GetBalance() Money {
	result, _ := d.OriginalAmount.Subtract(d.AmountPaid)
	return result
}

// IsFullyPaid returns true when AmountPaid >= OriginalAmount.
func (d *Debt) IsFullyPaid() bool {
	return d.AmountPaid.Amount.GreaterThanOrEqual(d.OriginalAmount.Amount)
}

// IsOverdue returns true when DueDate is past and status is not PAID or CANCELLED.
func (d *Debt) IsOverdue() bool {
	if d.Status == DebtStatusPaid || d.Status == DebtStatusCancelled {
		return false
	}
	return time.Now().UTC().After(d.DueDate)
}
