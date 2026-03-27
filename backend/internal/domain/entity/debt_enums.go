package entity

// DebtStatus represents the lifecycle state of a debt.
type DebtStatus string

const (
	DebtStatusPending   DebtStatus = "PENDING"
	DebtStatusPartial   DebtStatus = "PARTIAL"
	DebtStatusPaid      DebtStatus = "PAID"
	DebtStatusOverdue   DebtStatus = "OVERDUE"
	DebtStatusCancelled DebtStatus = "CANCELLED"
)

func (s DebtStatus) IsValid() bool {
	switch s {
	case DebtStatusPending, DebtStatusPartial, DebtStatusPaid, DebtStatusOverdue, DebtStatusCancelled:
		return true
	}
	return false
}

// DebtType represents the category of a debt.
type DebtType string

const (
	DebtTypeRent        DebtType = "RENT"
	DebtTypeUtilities   DebtType = "UTILITIES"
	DebtTypeMaintenance DebtType = "MAINTENANCE"
	DebtTypePenalty     DebtType = "PENALTY"
	DebtTypeOther       DebtType = "OTHER"
)

func (t DebtType) IsValid() bool {
	switch t {
	case DebtTypeRent, DebtTypeUtilities, DebtTypeMaintenance, DebtTypePenalty, DebtTypeOther:
		return true
	}
	return false
}

// TransactionType represents the kind of financial transaction.
type TransactionType string

const (
	TransactionTypePayment    TransactionType = "PAYMENT"
	TransactionTypeRefund     TransactionType = "REFUND"
	TransactionTypePenalty    TransactionType = "PENALTY"
	TransactionTypeAdjustment TransactionType = "ADJUSTMENT"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypePayment, TransactionTypeRefund, TransactionTypePenalty, TransactionTypeAdjustment:
		return true
	}
	return false
}

// PaymentMethod represents how a payment was made.
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "CASH"
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodMobileMoney  PaymentMethod = "MOBILE_MONEY"
	PaymentMethodCheck        PaymentMethod = "CHECK"
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodOther        PaymentMethod = "OTHER"
)

func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCash, PaymentMethodBankTransfer, PaymentMethodMobileMoney,
		PaymentMethodCheck, PaymentMethodCreditCard, PaymentMethodOther:
		return true
	}
	return false
}
