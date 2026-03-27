# Debts & Transactions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Debts & Transactions financial module — Money value object, Debt/Transaction entities with state machine, 12 use cases, 2 migrations, full HTTP API, and ~78 unit tests.

**Architecture:** Clean Architecture faithful port. Money value object provides currency-safe decimal arithmetic. Debts track amounts owed with status state machine (PENDING->PARTIAL->PAID/OVERDUE/CANCELLED). Transactions are immutable payment/refund records. Service-layer orchestration in RecordPayment/RecordRefund use cases coordinates both entities. Lazy overdue detection on read.

**Tech Stack:** Go 1.26+, Chi v5, pgx v5, shopspring/decimal (new dependency), google/uuid

**Spec:** `docs/superpowers/specs/2026-03-27-debts-module-design.md`

---

## File Structure

### New Files (37)

| # | Path | Responsibility |
|---|------|----------------|
| 1 | `backend/internal/domain/entity/money.go` | Money value object with currency-safe decimal arithmetic |
| 2 | `backend/internal/domain/entity/debt_enums.go` | DebtStatus, DebtType, TransactionType, PaymentMethod enums with IsValid() |
| 3 | `backend/internal/domain/entity/debt.go` | Debt entity with state machine (RecordPayment, ReversePayment, Cancel, MarkAsOverdue) |
| 4 | `backend/internal/domain/entity/transaction.go` | Transaction entity (immutable, verification workflow) |
| 5 | `backend/internal/domain/repository/debt_repository.go` | DebtRepository interface + DebtFilter |
| 6 | `backend/internal/domain/repository/transaction_repository.go` | TransactionRepository interface + TransactionFilter |
| 7 | `backend/internal/usecase/debt/create_debt.go` | CreateDebt use case |
| 8 | `backend/internal/usecase/debt/get_debt.go` | GetDebt use case (with lazy overdue) |
| 9 | `backend/internal/usecase/debt/list_debts.go` | ListDebts use case (with lazy overdue) |
| 10 | `backend/internal/usecase/debt/update_debt.go` | UpdateDebt use case |
| 11 | `backend/internal/usecase/debt/cancel_debt.go` | CancelDebt use case |
| 12 | `backend/internal/usecase/debt/mark_debt_paid.go` | MarkDebtPaid use case |
| 13 | `backend/internal/usecase/debt/delete_debt.go` | DeleteDebt use case (soft delete) |
| 14 | `backend/internal/usecase/transaction/record_payment.go` | RecordPayment use case (orchestrates Transaction + Debt) |
| 15 | `backend/internal/usecase/transaction/record_refund.go` | RecordRefund use case (orchestrates Transaction + Debt) |
| 16 | `backend/internal/usecase/transaction/get_transaction.go` | GetTransaction use case |
| 17 | `backend/internal/usecase/transaction/list_transactions.go` | ListTransactions use case |
| 18 | `backend/internal/usecase/transaction/verify_transaction.go` | VerifyTransaction use case |
| 19 | `backend/internal/delivery/http/dto/debt_dto.go` | MoneyDTO, CreateDebtRequest, UpdateDebtRequest, CancelDebtRequest, DebtResponse, DebtListResponse |
| 20 | `backend/internal/delivery/http/dto/transaction_dto.go` | RecordPaymentRequest, RecordRefundRequest, TransactionResponse, TransactionListResponse |
| 21 | `backend/internal/delivery/http/handler/debt_handler.go` | DebtHandler with 7 endpoints + Swagger |
| 22 | `backend/internal/delivery/http/handler/transaction_handler.go` | TransactionHandler with 5 endpoints + Swagger |
| 23 | `backend/internal/infrastructure/persistence/pg/debt_repo_pg.go` | PostgreSQL DebtRepository implementation |
| 24 | `backend/internal/infrastructure/persistence/pg/transaction_repo_pg.go` | PostgreSQL TransactionRepository implementation |
| 25 | `backend/migrations/000009_create_debts.up.sql` | Debts table migration |
| 26 | `backend/migrations/000009_create_debts.down.sql` | Debts table rollback |
| 27 | `backend/migrations/000010_create_transactions.up.sql` | Transactions table migration |
| 28 | `backend/migrations/000010_create_transactions.down.sql` | Transactions table rollback |
| 29 | `backend/tests/mocks/debt_repository_mock.go` | Manual mock for DebtRepository |
| 30 | `backend/tests/mocks/transaction_repository_mock.go` | Manual mock for TransactionRepository |
| 31 | `backend/tests/unit/money_test.go` | ~10 Money VO tests |
| 32 | `backend/tests/unit/debt_entity_test.go` | ~12 Debt entity tests |
| 33 | `backend/tests/unit/transaction_entity_test.go` | ~6 Transaction entity tests |
| 34 | `backend/tests/unit/debt_usecase_test.go` | ~18 Debt use case tests |
| 35 | `backend/tests/unit/transaction_usecase_test.go` | ~14 Transaction use case tests |
| 36 | `backend/tests/unit/debt_handler_test.go` | ~10 Debt handler tests |
| 37 | `backend/tests/unit/transaction_handler_test.go` | ~8 Transaction handler tests |

### Modified Files (4)

| # | Path | Change |
|---|------|--------|
| 1 | `backend/internal/domain/entity/errors.go` | Add ~20 debt/transaction/money domain errors |
| 2 | `backend/internal/usecase/property/deactivate_property.go` | Inject DebtRepository, add HasActiveDebtsForProperty check |
| 3 | `backend/internal/delivery/http/router/router.go` | Add Debt + Transaction handler fields and routes |
| 4 | `backend/cmd/server/main.go` | Wire debt + transaction modules, update DeactivatePropertyUseCase |

---

## Task 1: Money Value Object + Enums + Domain Errors

**Files:**
- Create: `backend/internal/domain/entity/money.go`
- Create: `backend/internal/domain/entity/debt_enums.go`
- Modify: `backend/internal/domain/entity/errors.go`
- Test: `backend/tests/unit/money_test.go`

- [ ] **Step 1: Add shopspring/decimal dependency**

```bash
cd backend && go get github.com/shopspring/decimal
```

Expected: `go.mod` and `go.sum` updated with `github.com/shopspring/decimal`.

- [ ] **Step 2: Add domain errors to errors.go**

Append to `backend/internal/domain/entity/errors.go`:

```go
// Domain errors for Money value object.
var (
	ErrCurrencyMismatch   = errors.New("currency mismatch: operations require same currency")
	ErrNegativeAmount     = errors.New("amount must not be negative")
	ErrInvalidCurrency    = errors.New("invalid currency code")
	ErrInsufficientAmount = errors.New("insufficient amount for operation")
)

// Domain errors for Debt entity.
var (
	ErrDebtDescriptionRequired      = errors.New("debt description is required")
	ErrDebtDescriptionTooLong       = errors.New("debt description must be 500 characters or less")
	ErrDebtAmountRequired           = errors.New("debt amount must be greater than zero")
	ErrDebtInvalidType              = errors.New("invalid debt type")
	ErrDebtDueDateRequired          = errors.New("debt due date is required")
	ErrDebtAlreadyPaid              = errors.New("debt is already fully paid")
	ErrDebtAlreadyCancelled         = errors.New("debt is already cancelled")
	ErrDebtOverpayment              = errors.New("payment amount exceeds remaining balance")
	ErrDebtInvalidStateTransition   = errors.New("invalid debt state transition")
	ErrDebtNotPayable               = errors.New("debt is not in a payable state")
	ErrPropertyHasActiveDebts       = errors.New("cannot deactivate property with active debts")
)

// Domain errors for enum validation.
var (
	ErrInvalidDebtStatus        = errors.New("invalid debt status")
	ErrInvalidDebtType          = errors.New("invalid debt type value")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrInvalidPaymentMethod     = errors.New("invalid payment method")
)

// Domain errors for Transaction entity.
var (
	ErrTransactionAmountRequired       = errors.New("transaction amount must be greater than zero")
	ErrTransactionInvalidType          = errors.New("invalid transaction type")
	ErrTransactionInvalidPaymentMethod = errors.New("invalid payment method")
	ErrTransactionDateRequired         = errors.New("transaction date is required")
	ErrTransactionAlreadyVerified      = errors.New("transaction is already verified")
	ErrTransactionDuplicateReference   = errors.New("duplicate reference number for this debt")
)
```

- [ ] **Step 3: Write Money value object tests**

Create `backend/tests/unit/money_test.go`:

```go
package unit

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

func TestNewMoney_ValidAmount(t *testing.T) {
	m, err := entity.NewMoney(decimal.NewFromFloat(100.50), entity.CurrencyPHP)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !m.Amount.Equal(decimal.NewFromFloat(100.50)) {
		t.Errorf("expected 100.50, got %s", m.Amount.String())
	}
	if m.Currency != entity.CurrencyPHP {
		t.Errorf("expected PHP, got %s", m.Currency)
	}
}

func TestNewMoney_ZeroAmount(t *testing.T) {
	m, err := entity.NewMoney(decimal.Zero, entity.CurrencyUSD)
	if err != nil {
		t.Fatalf("expected no error for zero amount, got %v", err)
	}
	if !m.IsZero() {
		t.Error("expected IsZero to be true")
	}
}

func TestNewMoney_NegativeAmount(t *testing.T) {
	_, err := entity.NewMoney(decimal.NewFromFloat(-10), entity.CurrencyPHP)
	if err != entity.ErrNegativeAmount {
		t.Fatalf("expected ErrNegativeAmount, got %v", err)
	}
}

func TestNewMoney_InvalidCurrency(t *testing.T) {
	_, err := entity.NewMoney(decimal.NewFromFloat(100), entity.Currency("XYZ"))
	if err != entity.ErrInvalidCurrency {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
}

func TestMoney_RoundsTwoDecimalPlaces(t *testing.T) {
	m, _ := entity.NewMoney(decimal.NewFromFloat(10.555), entity.CurrencyPHP)
	expected := decimal.NewFromFloat(10.56)
	if !m.Amount.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.String(), m.Amount.String())
	}
}

func TestMoney_Add_SameCurrency(t *testing.T) {
	a, _ := entity.NewMoney(decimal.NewFromFloat(50), entity.CurrencyPHP)
	b, _ := entity.NewMoney(decimal.NewFromFloat(30.50), entity.CurrencyPHP)
	result, err := a.Add(b)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Amount.Equal(decimal.NewFromFloat(80.50)) {
		t.Errorf("expected 80.50, got %s", result.Amount.String())
	}
}

func TestMoney_Add_DifferentCurrency(t *testing.T) {
	a, _ := entity.NewMoney(decimal.NewFromFloat(50), entity.CurrencyPHP)
	b, _ := entity.NewMoney(decimal.NewFromFloat(30), entity.CurrencyUSD)
	_, err := a.Add(b)
	if err != entity.ErrCurrencyMismatch {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
}

func TestMoney_Subtract_SameCurrency(t *testing.T) {
	a, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	b, _ := entity.NewMoney(decimal.NewFromFloat(40), entity.CurrencyPHP)
	result, err := a.Subtract(b)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Amount.Equal(decimal.NewFromFloat(60)) {
		t.Errorf("expected 60, got %s", result.Amount.String())
	}
}

func TestMoney_Subtract_InsufficientAmount(t *testing.T) {
	a, _ := entity.NewMoney(decimal.NewFromFloat(10), entity.CurrencyPHP)
	b, _ := entity.NewMoney(decimal.NewFromFloat(20), entity.CurrencyPHP)
	_, err := a.Subtract(b)
	if err != entity.ErrInsufficientAmount {
		t.Fatalf("expected ErrInsufficientAmount, got %v", err)
	}
}

func TestMoney_IsGreaterThan(t *testing.T) {
	a, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	b, _ := entity.NewMoney(decimal.NewFromFloat(50), entity.CurrencyPHP)
	gt, err := a.IsGreaterThan(b)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !gt {
		t.Error("expected 100 > 50")
	}
}
```

- [ ] **Step 4: Run tests to verify they fail**

```bash
cd backend && go test ./tests/unit/ -run TestNewMoney -v
cd backend && go test ./tests/unit/ -run TestMoney_ -v
```

Expected: Compilation errors — `entity.NewMoney`, `entity.Money`, `entity.Currency*` not found.

- [ ] **Step 5: Implement Money value object**

Create `backend/internal/domain/entity/money.go`:

```go
package entity

import "github.com/shopspring/decimal"

// Currency represents an ISO 4217 currency code.
type Currency string

const (
	CurrencyPHP Currency = "PHP"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyKES Currency = "KES"
	CurrencyTZS Currency = "TZS"
	CurrencyUGX Currency = "UGX"
)

// ValidCurrencies lists all supported currencies.
var ValidCurrencies = map[Currency]bool{
	CurrencyPHP: true,
	CurrencyUSD: true,
	CurrencyEUR: true,
	CurrencyGBP: true,
	CurrencyKES: true,
	CurrencyTZS: true,
	CurrencyUGX: true,
}

// Money represents a monetary amount with currency.
type Money struct {
	Amount   decimal.Decimal
	Currency Currency
}

// NewMoney creates a Money value, validating amount >= 0 and currency is supported.
func NewMoney(amount decimal.Decimal, currency Currency) (Money, error) {
	if amount.IsNegative() {
		return Money{}, ErrNegativeAmount
	}
	if !ValidCurrencies[currency] {
		return Money{}, ErrInvalidCurrency
	}
	return Money{
		Amount:   amount.Round(2),
		Currency: currency,
	}, nil
}

// ZeroMoney creates a zero-amount Money for the given currency.
func ZeroMoney(currency Currency) Money {
	return Money{
		Amount:   decimal.Zero,
		Currency: currency,
	}
}

// Add returns the sum of two Money values. Both must share the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount.Add(other.Amount).Round(2),
		Currency: m.Currency,
	}, nil
}

// Subtract returns the difference. Result must be >= 0.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	result := m.Amount.Sub(other.Amount).Round(2)
	if result.IsNegative() {
		return Money{}, ErrInsufficientAmount
	}
	return Money{
		Amount:   result,
		Currency: m.Currency,
	}, nil
}

// Multiply returns the product of Money and a factor.
func (m Money) Multiply(factor decimal.Decimal) Money {
	return Money{
		Amount:   m.Amount.Mul(factor).Round(2),
		Currency: m.Currency,
	}
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount.IsZero()
}

// IsGreaterThan returns true if m > other. Both must share the same currency.
func (m Money) IsGreaterThan(other Money) (bool, error) {
	if m.Currency != other.Currency {
		return false, ErrCurrencyMismatch
	}
	return m.Amount.GreaterThan(other.Amount), nil
}
```

- [ ] **Step 6: Implement enums**

Create `backend/internal/domain/entity/debt_enums.go`:

```go
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

// IsValid checks if the debt status is a recognized value.
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

// IsValid checks if the debt type is a recognized value.
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

// IsValid checks if the transaction type is a recognized value.
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

// IsValid checks if the payment method is a recognized value.
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCash, PaymentMethodBankTransfer, PaymentMethodMobileMoney,
		PaymentMethodCheck, PaymentMethodCreditCard, PaymentMethodOther:
		return true
	}
	return false
}
```

- [ ] **Step 7: Run Money tests to verify they pass**

```bash
cd backend && go test ./tests/unit/ -run "TestNewMoney|TestMoney_" -v
```

Expected: All 10 tests PASS.

- [ ] **Step 8: Run full test suite to verify no regressions**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: All 152 existing tests + 10 new = 162 PASS.

- [ ] **Step 9: Commit**

```bash
cd backend && git add internal/domain/entity/money.go internal/domain/entity/debt_enums.go internal/domain/entity/errors.go go.mod go.sum
cd backend && git add ../backend/tests/unit/money_test.go
git commit -m "feat: add Money value object, debt/transaction enums, and domain errors"
```

---

## Task 2: Debt Entity + Tests

**Files:**
- Create: `backend/internal/domain/entity/debt.go`
- Test: `backend/tests/unit/debt_entity_test.go`

- [ ] **Step 1: Write Debt entity tests**

Create `backend/tests/unit/debt_entity_test.go`:

```go
package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

func TestNewDebt_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	debt, err := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Monthly rent", amount, time.Now().Add(30*24*time.Hour), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.Status != entity.DebtStatusPending {
		t.Errorf("expected PENDING, got %s", debt.Status)
	}
	if !debt.AmountPaid.IsZero() {
		t.Error("expected AmountPaid to be zero")
	}
	if debt.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
}

func TestNewDebt_WithPropertyID(t *testing.T) {
	propID := uuid.New()
	amount, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	debt, err := entity.NewDebt(uuid.New(), uuid.New(), &propID, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.PropertyID == nil || *debt.PropertyID != propID {
		t.Error("expected PropertyID to be set")
	}
}

func TestNewDebt_DescriptionRequired(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "", amount, time.Now().Add(24*time.Hour), nil)
	if err != entity.ErrDebtDescriptionRequired {
		t.Fatalf("expected ErrDebtDescriptionRequired, got %v", err)
	}
}

func TestNewDebt_DescriptionTooLong(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	longDesc := make([]byte, 501)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	_, err := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, string(longDesc), amount, time.Now().Add(24*time.Hour), nil)
	if err != entity.ErrDebtDescriptionTooLong {
		t.Fatalf("expected ErrDebtDescriptionTooLong, got %v", err)
	}
}

func TestNewDebt_ZeroAmountFails(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.Zero, entity.CurrencyPHP)
	_, err := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)
	if err != entity.ErrDebtAmountRequired {
		t.Fatalf("expected ErrDebtAmountRequired, got %v", err)
	}
}

func TestNewDebt_DueDateRequired(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Time{}, nil)
	if err != entity.ErrDebtDueDateRequired {
		t.Fatalf("expected ErrDebtDueDateRequired, got %v", err)
	}
}

func TestDebt_RecordPayment_Partial(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	payment, _ := entity.NewMoney(decimal.NewFromFloat(300), entity.CurrencyPHP)
	err := debt.RecordPayment(payment)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.Status != entity.DebtStatusPartial {
		t.Errorf("expected PARTIAL, got %s", debt.Status)
	}
	if !debt.AmountPaid.Amount.Equal(decimal.NewFromFloat(300)) {
		t.Errorf("expected 300 paid, got %s", debt.AmountPaid.Amount.String())
	}
}

func TestDebt_RecordPayment_FullyPaid(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	payment, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	err := debt.RecordPayment(payment)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.Status != entity.DebtStatusPaid {
		t.Errorf("expected PAID, got %s", debt.Status)
	}
}

func TestDebt_RecordPayment_Overpayment(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	overpay, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	err := debt.RecordPayment(overpay)
	if err != entity.ErrDebtOverpayment {
		t.Fatalf("expected ErrDebtOverpayment, got %v", err)
	}
}

func TestDebt_RecordPayment_AlreadyPaid(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = debt.RecordPayment(payment)

	extra, _ := entity.NewMoney(decimal.NewFromFloat(10), entity.CurrencyPHP)
	err := debt.RecordPayment(extra)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

func TestDebt_ReversePayment(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	payment, _ := entity.NewMoney(decimal.NewFromFloat(600), entity.CurrencyPHP)
	_ = debt.RecordPayment(payment) // PARTIAL

	refund, _ := entity.NewMoney(decimal.NewFromFloat(600), entity.CurrencyPHP)
	err := debt.ReversePayment(refund)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.Status != entity.DebtStatusPending {
		t.Errorf("expected PENDING after full reversal, got %s", debt.Status)
	}
}

func TestDebt_Cancel_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	reason := "Landlord forgave"
	err := debt.Cancel(&reason)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if debt.Status != entity.DebtStatusCancelled {
		t.Errorf("expected CANCELLED, got %s", debt.Status)
	}
}

func TestDebt_Cancel_AlreadyPaid(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = debt.RecordPayment(payment) // PAID

	err := debt.Cancel(nil)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

func TestDebt_GetBalance(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(24*time.Hour), nil)

	payment, _ := entity.NewMoney(decimal.NewFromFloat(250), entity.CurrencyPHP)
	_ = debt.RecordPayment(payment)

	balance := debt.GetBalance()
	if !balance.Amount.Equal(decimal.NewFromFloat(750)) {
		t.Errorf("expected 750, got %s", balance.Amount.String())
	}
}

func TestDebt_IsOverdue(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	debt, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(-24*time.Hour), nil)
	if !debt.IsOverdue() {
		t.Error("expected debt with past due date to be overdue")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./tests/unit/ -run "TestNewDebt|TestDebt_" -v
```

Expected: Compilation errors — `entity.NewDebt`, `entity.Debt` not found.

- [ ] **Step 3: Implement Debt entity**

Create `backend/internal/domain/entity/debt.go`:

```go
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
```

- [ ] **Step 4: Run Debt entity tests to verify they pass**

```bash
cd backend && go test ./tests/unit/ -run "TestNewDebt|TestDebt_" -v
```

Expected: All 12 tests PASS.

- [ ] **Step 5: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: 162 + 12 = ~174 PASS.

- [ ] **Step 6: Commit**

```bash
cd backend && git add internal/domain/entity/debt.go
cd backend && git add ../backend/tests/unit/debt_entity_test.go
git commit -m "feat: add Debt entity with state machine and validation"
```

---

## Task 3: Transaction Entity + Tests

**Files:**
- Create: `backend/internal/domain/entity/transaction.go`
- Test: `backend/tests/unit/transaction_entity_test.go`

- [ ] **Step 1: Write Transaction entity tests**

Create `backend/tests/unit/transaction_entity_test.go`:

```go
package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

func TestNewTransaction_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	userID := uuid.New()
	tx, err := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), &userID,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Now(), "Payment for rent", nil, nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}
	if tx.IsVerified {
		t.Error("expected IsVerified to be false")
	}
}

func TestNewTransaction_NilRecordedByUserID(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, err := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), nil,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Now(), "System payment", nil, nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.RecordedByUserID != nil {
		t.Error("expected RecordedByUserID to be nil")
	}
}

func TestNewTransaction_ZeroAmountFails(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.Zero, entity.CurrencyPHP)
	_, err := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), nil,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Now(), "Payment", nil, nil,
	)
	if err != entity.ErrTransactionAmountRequired {
		t.Fatalf("expected ErrTransactionAmountRequired, got %v", err)
	}
}

func TestNewTransaction_DateRequired(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), nil,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Time{}, "Payment", nil, nil,
	)
	if err != entity.ErrTransactionDateRequired {
		t.Fatalf("expected ErrTransactionDateRequired, got %v", err)
	}
}

func TestTransaction_Verify_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), nil,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Now(), "Payment", nil, nil,
	)

	verifierID := uuid.New()
	err := tx.Verify(verifierID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tx.IsVerified {
		t.Error("expected IsVerified to be true")
	}
	if tx.VerifiedByUserID == nil || *tx.VerifiedByUserID != verifierID {
		t.Error("expected VerifiedByUserID to match")
	}
	if tx.VerifiedAt == nil {
		t.Error("expected VerifiedAt to be set")
	}
}

func TestTransaction_Verify_AlreadyVerified(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(
		uuid.New(), uuid.New(), uuid.New(), nil,
		entity.TransactionTypePayment, amount, entity.PaymentMethodCash,
		time.Now(), "Payment", nil, nil,
	)
	_ = tx.Verify(uuid.New())

	err := tx.Verify(uuid.New())
	if err != entity.ErrTransactionAlreadyVerified {
		t.Fatalf("expected ErrTransactionAlreadyVerified, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./tests/unit/ -run "TestNewTransaction|TestTransaction_" -v
```

Expected: Compilation errors — `entity.NewTransaction` not found.

- [ ] **Step 3: Implement Transaction entity**

Create `backend/internal/domain/entity/transaction.go`:

```go
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
```

- [ ] **Step 4: Run Transaction entity tests to verify they pass**

```bash
cd backend && go test ./tests/unit/ -run "TestNewTransaction|TestTransaction_" -v
```

Expected: All 6 tests PASS.

- [ ] **Step 5: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: ~180 PASS.

- [ ] **Step 6: Commit**

```bash
cd backend && git add internal/domain/entity/transaction.go
cd backend && git add ../backend/tests/unit/transaction_entity_test.go
git commit -m "feat: add Transaction entity with verification workflow"
```

---

## Task 4: Repository Interfaces + Mocks + Database Migrations

**Files:**
- Create: `backend/internal/domain/repository/debt_repository.go`
- Create: `backend/internal/domain/repository/transaction_repository.go`
- Create: `backend/tests/mocks/debt_repository_mock.go`
- Create: `backend/tests/mocks/transaction_repository_mock.go`
- Create: `backend/migrations/000009_create_debts.up.sql`
- Create: `backend/migrations/000009_create_debts.down.sql`
- Create: `backend/migrations/000010_create_transactions.up.sql`
- Create: `backend/migrations/000010_create_transactions.down.sql`

- [ ] **Step 1: Create DebtRepository interface**

Create `backend/internal/domain/repository/debt_repository.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// DebtFilter holds optional filters for listing debts.
type DebtFilter struct {
	TenantID   *uuid.UUID
	LandlordID *uuid.UUID
	PropertyID *uuid.UUID
	Status     *entity.DebtStatus
	DebtType   *entity.DebtType
	IsOverdue  *bool
	Search     *string
	Limit      int
	Offset     int
}

// DebtRepository defines the interface for debt persistence operations.
type DebtRepository interface {
	Create(ctx context.Context, debt *entity.Debt) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error)
	List(ctx context.Context, filter DebtFilter) ([]*entity.Debt, int, error)
	Update(ctx context.Context, debt *entity.Debt) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error)
}
```

- [ ] **Step 2: Create TransactionRepository interface**

Create `backend/internal/domain/repository/transaction_repository.go`:

```go
package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// TransactionFilter holds optional filters for listing transactions.
type TransactionFilter struct {
	DebtID     *uuid.UUID
	TenantID   *uuid.UUID
	LandlordID *uuid.UUID
	Type       *entity.TransactionType
	IsVerified *bool
	Limit      int
	Offset     int
}

// TransactionRepository defines the interface for transaction persistence operations.
type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	List(ctx context.Context, filter TransactionFilter) ([]*entity.Transaction, int, error)
	Update(ctx context.Context, tx *entity.Transaction) error
	ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error)
}
```

- [ ] **Step 3: Create DebtRepository mock**

Create `backend/tests/mocks/debt_repository_mock.go`:

```go
package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// DebtRepositoryMock is a test mock for repository.DebtRepository.
type DebtRepositoryMock struct {
	CreateFn                     func(ctx context.Context, debt *entity.Debt) error
	GetByIDFn                    func(ctx context.Context, id uuid.UUID) (*entity.Debt, error)
	ListFn                       func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error)
	UpdateFn                     func(ctx context.Context, debt *entity.Debt) error
	DeleteFn                     func(ctx context.Context, id uuid.UUID) error
	HasActiveDebtsForPropertyFn  func(ctx context.Context, propertyID uuid.UUID) (bool, error)
}

func (m *DebtRepositoryMock) Create(ctx context.Context, debt *entity.Debt) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, debt)
	}
	return nil
}

func (m *DebtRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *DebtRepositoryMock) List(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *DebtRepositoryMock) Update(ctx context.Context, debt *entity.Debt) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, debt)
	}
	return nil
}

func (m *DebtRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

func (m *DebtRepositoryMock) HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error) {
	if m.HasActiveDebtsForPropertyFn != nil {
		return m.HasActiveDebtsForPropertyFn(ctx, propertyID)
	}
	return false, nil
}
```

- [ ] **Step 4: Create TransactionRepository mock**

Create `backend/tests/mocks/transaction_repository_mock.go`:

```go
package mocks

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// TransactionRepositoryMock is a test mock for repository.TransactionRepository.
type TransactionRepositoryMock struct {
	CreateFn                 func(ctx context.Context, tx *entity.Transaction) error
	GetByIDFn                func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	ListFn                   func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error)
	UpdateFn                 func(ctx context.Context, tx *entity.Transaction) error
	ExistsByReferenceNumberFn func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error)
}

func (m *TransactionRepositoryMock) Create(ctx context.Context, tx *entity.Transaction) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, tx)
	}
	return nil
}

func (m *TransactionRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *TransactionRepositoryMock) List(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *TransactionRepositoryMock) Update(ctx context.Context, tx *entity.Transaction) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, tx)
	}
	return nil
}

func (m *TransactionRepositoryMock) ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
	if m.ExistsByReferenceNumberFn != nil {
		return m.ExistsByReferenceNumberFn(ctx, debtID, refNum)
	}
	return false, nil
}
```

- [ ] **Step 5: Create debts migration (up)**

Create `backend/migrations/000009_create_debts.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS debts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    property_id UUID REFERENCES properties(id) ON DELETE SET NULL,
    debt_type VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    original_amount DECIMAL(14,2) NOT NULL,
    original_currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    amount_paid DECIMAL(14,2) NOT NULL DEFAULT 0,
    amount_paid_currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    due_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_debts_tenant_id ON debts(tenant_id);
CREATE INDEX idx_debts_landlord_id ON debts(landlord_id);
CREATE INDEX idx_debts_property_id ON debts(property_id);
CREATE INDEX idx_debts_status ON debts(status);
CREATE INDEX idx_debts_due_date_active ON debts(due_date) WHERE status NOT IN ('PAID', 'CANCELLED') AND deleted_at IS NULL;
CREATE INDEX idx_debts_deleted_at ON debts(deleted_at);

-- Auto-update updated_at trigger
CREATE TRIGGER set_debts_updated_at
    BEFORE UPDATE ON debts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

- [ ] **Step 6: Create debts migration (down)**

Create `backend/migrations/000009_create_debts.down.sql`:

```sql
DROP TRIGGER IF EXISTS set_debts_updated_at ON debts;
DROP TABLE IF EXISTS debts;
```

- [ ] **Step 7: Create transactions migration (up)**

Create `backend/migrations/000010_create_transactions.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    recorded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(14,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'PHP',
    payment_method VARCHAR(20) NOT NULL,
    transaction_date DATE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    receipt_number VARCHAR(100),
    reference_number VARCHAR(100),
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verified_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique partial index: reference_number per debt (when not null)
CREATE UNIQUE INDEX idx_transactions_debt_reference ON transactions(debt_id, reference_number) WHERE reference_number IS NOT NULL;

-- Indexes
CREATE INDEX idx_transactions_debt_id ON transactions(debt_id);
CREATE INDEX idx_transactions_tenant_id ON transactions(tenant_id);
CREATE INDEX idx_transactions_landlord_id ON transactions(landlord_id);
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

-- Auto-update updated_at trigger
CREATE TRIGGER set_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

- [ ] **Step 8: Create transactions migration (down)**

Create `backend/migrations/000010_create_transactions.down.sql`:

```sql
DROP TRIGGER IF EXISTS set_transactions_updated_at ON transactions;
DROP TABLE IF EXISTS transactions;
```

- [ ] **Step 9: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: Clean compilation.

- [ ] **Step 10: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: All ~180 tests PASS (no new tests this task, but mocks must compile).

- [ ] **Step 11: Commit**

```bash
cd backend && git add internal/domain/repository/debt_repository.go internal/domain/repository/transaction_repository.go
cd backend && git add ../backend/tests/mocks/debt_repository_mock.go ../backend/tests/mocks/transaction_repository_mock.go
cd backend && git add migrations/000009_create_debts.up.sql migrations/000009_create_debts.down.sql
cd backend && git add migrations/000010_create_transactions.up.sql migrations/000010_create_transactions.down.sql
git commit -m "feat: add debt/transaction repository interfaces, mocks, and migrations"
```

---

## Task 5: Debt Use Cases + Tests

**Files:**
- Create: `backend/internal/usecase/debt/create_debt.go`
- Create: `backend/internal/usecase/debt/get_debt.go`
- Create: `backend/internal/usecase/debt/list_debts.go`
- Create: `backend/internal/usecase/debt/update_debt.go`
- Create: `backend/internal/usecase/debt/cancel_debt.go`
- Create: `backend/internal/usecase/debt/mark_debt_paid.go`
- Create: `backend/internal/usecase/debt/delete_debt.go`
- Modify: `backend/internal/usecase/property/deactivate_property.go`
- Test: `backend/tests/unit/debt_usecase_test.go`

- [ ] **Step 1: Write Debt use case tests**

Create `backend/tests/unit/debt_usecase_test.go`:

```go
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	debt "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	property "github.com/ISubamariner/guimba-go/backend/internal/usecase/property"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- helpers ---

func newTestDebt(tenantID, landlordID uuid.UUID, amount float64) *entity.Debt {
	m, _ := entity.NewMoney(decimal.NewFromFloat(amount), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, nil, entity.DebtTypeRent, "Test rent", m, time.Now().Add(30*24*time.Hour), nil)
	return d
}

// --- CreateDebt ---

func TestCreateDebt_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	d := newTestDebt(tenantID, landlordID, 1000)
	err := uc.Execute(context.Background(), d)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateDebt_TenantNotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return nil, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	d := newTestDebt(uuid.New(), uuid.New(), 1000)
	err := uc.Execute(context.Background(), d)
	if err == nil {
		t.Fatal("expected error when tenant not found")
	}
}

func TestCreateDebt_WithProperty_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	propID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, OwnerID: landlordID, IsActive: true}, nil
		},
	}

	uc := debt.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	m, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, &propID, entity.DebtTypeRent, "Rent", m, time.Now().Add(30*24*time.Hour), nil)
	err := uc.Execute(context.Background(), d)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- GetDebt ---

func TestGetDebt_Success(t *testing.T) {
	debtID := uuid.New()
	d := newTestDebt(uuid.New(), uuid.New(), 500)
	d.ID = debtID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	result, err := uc.Execute(context.Background(), debtID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != debtID {
		t.Error("expected ID to match")
	}
}

func TestGetDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestGetDebt_LazyOverdue(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 500)
	d.DueDate = time.Now().Add(-48 * time.Hour)
	d.Status = entity.DebtStatusPending

	var updatedDebt *entity.Debt
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error {
			updatedDebt = debt
			return nil
		},
	}
	uc := debt.NewGetDebtUseCase(debtRepo)
	result, err := uc.Execute(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != entity.DebtStatusOverdue {
		t.Errorf("expected OVERDUE, got %s", result.Status)
	}
	if updatedDebt == nil {
		t.Error("expected update to be called for lazy overdue")
	}
}

// --- ListDebts ---

func TestListDebts_Success(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			d := newTestDebt(uuid.New(), uuid.New(), 100)
			return []*entity.Debt{d}, 1, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	debts, total, err := uc.Execute(context.Background(), repository.DebtFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(debts) != 1 {
		t.Errorf("expected 1 debt, got %d", len(debts))
	}
}

func TestListDebts_DefaultLimit(t *testing.T) {
	var captured repository.DebtFilter
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	_, _, _ = uc.Execute(context.Background(), repository.DebtFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

func TestListDebts_MaxLimit(t *testing.T) {
	var captured repository.DebtFilter
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := debt.NewListDebtsUseCase(debtRepo)
	_, _, _ = uc.Execute(context.Background(), repository.DebtFilter{Limit: 999})
	if captured.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", captured.Limit)
	}
}

// --- UpdateDebt ---

func TestUpdateDebt_Success(t *testing.T) {
	debtID := uuid.New()
	existing := newTestDebt(uuid.New(), uuid.New(), 1000)
	existing.ID = debtID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	uc := debt.NewUpdateDebtUseCase(debtRepo)

	updates := &entity.Debt{Description: "Updated rent description", DebtType: entity.DebtTypeRent, DueDate: time.Now().Add(60 * 24 * time.Hour)}
	err := uc.Execute(context.Background(), debtID, updates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewUpdateDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), uuid.New(), &entity.Debt{Description: "x", DebtType: entity.DebtTypeRent, DueDate: time.Now().Add(24 * time.Hour)})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- CancelDebt ---

func TestCancelDebt_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 500)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			if d.Status != entity.DebtStatusCancelled {
				t.Errorf("expected CANCELLED, got %s", d.Status)
			}
			return nil
		},
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	reason := "No longer valid"
	err := uc.Execute(context.Background(), existing.ID, &reason)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCancelDebt_AlreadyPaid(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = existing.RecordPayment(payment) // PAID

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID, nil)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

func TestCancelDebt_AlreadyCancelled(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	reason := "first cancel"
	_ = existing.Cancel(&reason) // CANCELLED

	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	uc := debt.NewCancelDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID, nil)
	// Cancel on already-cancelled is idempotent (no error), status stays CANCELLED
	if err != nil {
		t.Fatalf("expected no error on re-cancel, got %v", err)
	}
}

// --- MarkDebtPaid ---

func TestMarkDebtPaid_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 500)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			if d.Status != entity.DebtStatusPaid {
				t.Errorf("expected PAID, got %s", d.Status)
			}
			return nil
		},
	}
	uc := debt.NewMarkDebtPaidUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// --- DeleteDebt ---

func TestDeleteDebt_Success(t *testing.T) {
	existing := newTestDebt(uuid.New(), uuid.New(), 100)
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return existing, nil
		},
		DeleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	uc := debt.NewDeleteDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteDebt_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	uc := debt.NewDeleteDebtUseCase(debtRepo)
	err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- Property Deactivation with Debt Check ---

func TestDeactivateProperty_BlockedByActiveDebts(t *testing.T) {
	propID := uuid.New()
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		HasActiveDebtsForPropertyFn: func(ctx context.Context, propertyID uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(propRepo, debtRepo)
	err := uc.Execute(context.Background(), propID)
	if err == nil {
		t.Fatal("expected error when property has active debts")
	}
}

func TestDeactivateProperty_AllowedWhenNoActiveDebts(t *testing.T) {
	propID := uuid.New()
	propRepo := &mocks.PropertyRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
			return &entity.Property{ID: propID, Name: "Farm", PropertyCode: "F-001", SizeInSqm: 500, IsActive: true}, nil
		},
		UpdateFn: func(ctx context.Context, p *entity.Property) error { return nil },
	}
	debtRepo := &mocks.DebtRepositoryMock{
		HasActiveDebtsForPropertyFn: func(ctx context.Context, propertyID uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	uc := property.NewDeactivatePropertyUseCase(propRepo, debtRepo)
	err := uc.Execute(context.Background(), propID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./tests/unit/ -run "TestCreateDebt|TestGetDebt|TestListDebts|TestUpdateDebt|TestCancelDebt|TestMarkDebtPaid|TestDeleteDebt|TestDeactivateProperty_Blocked|TestDeactivateProperty_Allowed" -v
```

Expected: Compilation errors — use case constructors not found.

- [ ] **Step 3: Implement CreateDebt use case**

Create `backend/internal/usecase/debt/create_debt.go`:

```go
package debt

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreateDebtUseCase struct {
	repo       repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	propRepo   repository.PropertyRepository
}

func NewCreateDebtUseCase(repo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, propRepo repository.PropertyRepository) *CreateDebtUseCase {
	return &CreateDebtUseCase{repo: repo, userRepo: userRepo, tenantRepo: tenantRepo, propRepo: propRepo}
}

func (uc *CreateDebtUseCase) Execute(ctx context.Context, d *entity.Debt) error {
	if err := d.Validate(); err != nil {
		return err
	}

	// Validate landlord exists
	landlord, err := uc.userRepo.GetByID(ctx, d.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", d.LandlordID)
	}

	// Validate tenant exists and belongs to landlord
	tenant, err := uc.tenantRepo.GetByID(ctx, d.TenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return apperror.NewNotFound("Tenant", d.TenantID)
	}
	if tenant.LandlordID != d.LandlordID {
		return apperror.NewForbidden("tenant does not belong to this landlord")
	}

	// Validate property if provided
	if d.PropertyID != nil {
		prop, err := uc.propRepo.GetByID(ctx, *d.PropertyID)
		if err != nil {
			return err
		}
		if prop == nil {
			return apperror.NewNotFound("Property", *d.PropertyID)
		}
		if prop.OwnerID != d.LandlordID {
			return apperror.NewForbidden("property does not belong to this landlord")
		}
	}

	return uc.repo.Create(ctx, d)
}
```

- [ ] **Step 4: Implement GetDebt use case (with lazy overdue)**

Create `backend/internal/usecase/debt/get_debt.go`:

```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetDebtUseCase struct {
	repo repository.DebtRepository
}

func NewGetDebtUseCase(repo repository.DebtRepository) *GetDebtUseCase {
	return &GetDebtUseCase{repo: repo}
}

func (uc *GetDebtUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", id)
	}

	// Lazy overdue detection
	if d.IsOverdue() && d.Status != entity.DebtStatusOverdue {
		d.MarkAsOverdue()
		_ = uc.repo.Update(ctx, d)
	}

	return d, nil
}
```

- [ ] **Step 5: Implement ListDebts use case**

Create `backend/internal/usecase/debt/list_debts.go`:

```go
package debt

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListDebtsUseCase struct {
	repo repository.DebtRepository
}

func NewListDebtsUseCase(repo repository.DebtRepository) *ListDebtsUseCase {
	return &ListDebtsUseCase{repo: repo}
}

func (uc *ListDebtsUseCase) Execute(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	debts, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Lazy overdue detection on all results
	for _, d := range debts {
		if d.IsOverdue() && d.Status != entity.DebtStatusOverdue {
			d.MarkAsOverdue()
			_ = uc.repo.Update(ctx, d)
		}
	}

	return debts, total, nil
}
```

- [ ] **Step 6: Implement UpdateDebt use case**

Create `backend/internal/usecase/debt/update_debt.go`:

```go
package debt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdateDebtUseCase struct {
	repo repository.DebtRepository
}

func NewUpdateDebtUseCase(repo repository.DebtRepository) *UpdateDebtUseCase {
	return &UpdateDebtUseCase{repo: repo}
}

func (uc *UpdateDebtUseCase) Execute(ctx context.Context, id uuid.UUID, updates *entity.Debt) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	// Apply mutable fields only
	existing.Description = updates.Description
	existing.DebtType = updates.DebtType
	existing.DueDate = updates.DueDate
	existing.PropertyID = updates.PropertyID
	existing.Notes = updates.Notes
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, existing)
}
```

- [ ] **Step 7: Implement CancelDebt use case**

Create `backend/internal/usecase/debt/cancel_debt.go`:

```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CancelDebtUseCase struct {
	repo repository.DebtRepository
}

func NewCancelDebtUseCase(repo repository.DebtRepository) *CancelDebtUseCase {
	return &CancelDebtUseCase{repo: repo}
}

func (uc *CancelDebtUseCase) Execute(ctx context.Context, id uuid.UUID, reason *string) error {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return apperror.NewNotFound("Debt", id)
	}

	if err := d.Cancel(reason); err != nil {
		return err
	}

	return uc.repo.Update(ctx, d)
}
```

- [ ] **Step 8: Implement MarkDebtPaid use case**

Create `backend/internal/usecase/debt/mark_debt_paid.go`:

```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type MarkDebtPaidUseCase struct {
	repo repository.DebtRepository
}

func NewMarkDebtPaidUseCase(repo repository.DebtRepository) *MarkDebtPaidUseCase {
	return &MarkDebtPaidUseCase{repo: repo}
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

	return uc.repo.Update(ctx, d)
}
```

- [ ] **Step 9: Implement DeleteDebt use case**

Create `backend/internal/usecase/debt/delete_debt.go`:

```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeleteDebtUseCase struct {
	repo repository.DebtRepository
}

func NewDeleteDebtUseCase(repo repository.DebtRepository) *DeleteDebtUseCase {
	return &DeleteDebtUseCase{repo: repo}
}

func (uc *DeleteDebtUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	return uc.repo.Delete(ctx, id)
}
```

- [ ] **Step 10: Modify DeactivatePropertyUseCase to inject DebtRepository**

Modify `backend/internal/usecase/property/deactivate_property.go` — change constructor to accept DebtRepository:

```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo     repository.PropertyRepository
	debtRepo repository.DebtRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository, debtRepo repository.DebtRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo, debtRepo: debtRepo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	hasDebts, err := uc.debtRepo.HasActiveDebtsForProperty(ctx, id)
	if err != nil {
		return err
	}
	if hasDebts {
		return apperror.NewConflict(entity.ErrPropertyHasActiveDebts.Error())
	}

	existing.IsActive = false
	return uc.repo.Update(ctx, existing)
}
```

**Note:** This changes the constructor signature. Update these callers:

**Update `backend/tests/unit/property_usecase_test.go`** — change both existing deactivate tests:

```go
// TestDeactivateProperty_Success (line ~207): change constructor call
uc := property.NewDeactivatePropertyUseCase(repo, &mocks.DebtRepositoryMock{})

// TestDeactivateProperty_NotFound (line ~220): change constructor call
uc := property.NewDeactivatePropertyUseCase(repo, &mocks.DebtRepositoryMock{})
```

**Update `backend/tests/unit/property_handler_test.go`** — change `newPropertyHandler` helper:

```go
func newPropertyHandler(repo *mocks.PropertyRepositoryMock, userRepo *mocks.UserRepositoryMock) *handler.PropertyHandler {
	createUC := propertyuc.NewCreatePropertyUseCase(repo, userRepo)
	getUC := propertyuc.NewGetPropertyUseCase(repo)
	listUC := propertyuc.NewListPropertiesUseCase(repo)
	updateUC := propertyuc.NewUpdatePropertyUseCase(repo)
	deactivateUC := propertyuc.NewDeactivatePropertyUseCase(repo, &mocks.DebtRepositoryMock{})  // <-- add DebtRepositoryMock
	deleteUC := propertyuc.NewDeletePropertyUseCase(repo)
	return handler.NewPropertyHandler(createUC, getUC, listUC, updateUC, deactivateUC, deleteUC)
}
```

**`backend/cmd/server/main.go`** — update done in Task 9 (debtRepo declared before property wiring block).

- [ ] **Step 11: Run tests**

```bash
cd backend && go test ./tests/unit/ -v -count=1
```

Expected: All existing tests + ~18 new debt use case tests PASS.

- [ ] **Step 12: Commit**

```bash
cd backend && git add internal/usecase/debt/ internal/usecase/property/deactivate_property.go
cd backend && git add ../backend/tests/unit/debt_usecase_test.go ../backend/tests/unit/property_usecase_test.go ../backend/tests/unit/property_handler_test.go
git commit -m "feat: add debt use cases with lazy overdue, cancel, mark paid, and property deactivation constraint"
```

---

## Task 6: Transaction Use Cases + Tests

**Files:**
- Create: `backend/internal/usecase/transaction/record_payment.go`
- Create: `backend/internal/usecase/transaction/record_refund.go`
- Create: `backend/internal/usecase/transaction/get_transaction.go`
- Create: `backend/internal/usecase/transaction/list_transactions.go`
- Create: `backend/internal/usecase/transaction/verify_transaction.go`
- Test: `backend/tests/unit/transaction_usecase_test.go`

- [ ] **Step 1: Write Transaction use case tests**

Create `backend/tests/unit/transaction_usecase_test.go`:

```go
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	transaction "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

// --- RecordPayment ---

func TestRecordPayment_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	d := newTestDebt(tenantID, landlordID, 1000)

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
		ExistsByReferenceNumberFn: func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
			return false, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)

	amount, _ := entity.NewMoney(decimal.NewFromFloat(300), entity.CurrencyPHP)
	recorderID := uuid.New()
	tx, err := uc.Execute(context.Background(), d.ID, tenantID, &recorderID, amount, entity.PaymentMethodCash, time.Now(), "Partial payment", nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatal("expected transaction to be returned")
	}
	if d.Status != entity.DebtStatusPartial {
		t.Errorf("expected debt status PARTIAL, got %s", d.Status)
	}
}

func TestRecordPayment_DebtNotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return nil, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Payment", nil, nil)
	if err == nil {
		t.Fatal("expected error when debt not found")
	}
}

func TestRecordPayment_Overpayment(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 100)

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	amount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Over", nil, nil)
	if err != entity.ErrDebtOverpayment {
		t.Fatalf("expected ErrDebtOverpayment, got %v", err)
	}
}

func TestRecordPayment_DuplicateReference(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000)

	txRepo := &mocks.TransactionRepositoryMock{
		ExistsByReferenceNumberFn: func(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
			return true, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	ref := "REF-001"
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Payment", nil, &ref)
	if err != entity.ErrTransactionDuplicateReference {
		t.Fatalf("expected ErrTransactionDuplicateReference, got %v", err)
	}
}

func TestRecordPayment_AlreadyPaid(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 100)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = d.RecordPayment(payment) // PAID

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	amount, _ := entity.NewMoney(decimal.NewFromFloat(50), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, amount, entity.PaymentMethodCash, time.Now(), "Late", nil, nil)
	if err != entity.ErrDebtAlreadyPaid {
		t.Fatalf("expected ErrDebtAlreadyPaid, got %v", err)
	}
}

// --- RecordRefund ---

func TestRecordRefund_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	d := newTestDebt(tenantID, landlordID, 1000)
	// Pre-pay 500
	payment, _ := entity.NewMoney(decimal.NewFromFloat(500), entity.CurrencyPHP)
	_ = d.RecordPayment(payment)

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
		UpdateFn: func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	recorderID := uuid.New()
	tx, err := uc.Execute(context.Background(), d.ID, tenantID, &recorderID, refundAmount, entity.PaymentMethodCash, time.Now(), "Partial refund", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatal("expected transaction returned")
	}
	if d.Status != entity.DebtStatusPartial {
		t.Errorf("expected PARTIAL after partial refund of partial payment, got %s", d.Status)
	}
}

func TestRecordRefund_ZeroPayments(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000) // no payments made

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, refundAmount, entity.PaymentMethodCash, time.Now(), "Refund", nil)
	if err != entity.ErrInsufficientAmount {
		t.Fatalf("expected ErrInsufficientAmount, got %v", err)
	}
}

func TestRecordRefund_ExceedsAmountPaid(t *testing.T) {
	d := newTestDebt(uuid.New(), uuid.New(), 1000)
	payment, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	_ = d.RecordPayment(payment)

	txRepo := &mocks.TransactionRepositoryMock{}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	userRepo := &mocks.UserRepositoryMock{}
	tenantRepo := &mocks.TenantRepositoryMock{}

	uc := transaction.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	refundAmount, _ := entity.NewMoney(decimal.NewFromFloat(200), entity.CurrencyPHP)
	_, err := uc.Execute(context.Background(), d.ID, uuid.New(), nil, refundAmount, entity.PaymentMethodCash, time.Now(), "Too much", nil)
	if err != entity.ErrInsufficientAmount {
		t.Fatalf("expected ErrInsufficientAmount, got %v", err)
	}
}

// --- GetTransaction ---

func TestGetTransaction_Success(t *testing.T) {
	txID := uuid.New()
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	expected, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	expected.ID = txID

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return expected, nil
		},
	}
	uc := transaction.NewGetTransactionUseCase(txRepo)
	result, err := uc.Execute(context.Background(), txID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != txID {
		t.Error("expected ID to match")
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	uc := transaction.NewGetTransactionUseCase(txRepo)
	_, err := uc.Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}

// --- ListTransactions ---

func TestListTransactions_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx1, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			return []*entity.Transaction{tx1}, 1, nil
		},
	}
	uc := transaction.NewListTransactionsUseCase(txRepo)
	txs, total, err := uc.Execute(context.Background(), repository.TransactionFilter{Limit: 20})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 1 || len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

func TestListTransactions_DefaultLimit(t *testing.T) {
	var captured repository.TransactionFilter
	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			captured = filter
			return nil, 0, nil
		},
	}
	uc := transaction.NewListTransactionsUseCase(txRepo)
	_, _, _ = uc.Execute(context.Background(), repository.TransactionFilter{Limit: 0})
	if captured.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", captured.Limit)
	}
}

// --- VerifyTransaction ---

func TestVerifyTransaction_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return tx, nil
		},
		UpdateFn: func(ctx context.Context, t *entity.Transaction) error { return nil },
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo)
	verifierID := uuid.New()
	err := uc.Execute(context.Background(), tx.ID, verifierID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !tx.IsVerified {
		t.Error("expected transaction to be verified")
	}
}

func TestVerifyTransaction_AlreadyVerified(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	_ = tx.Verify(uuid.New())

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return tx, nil
		},
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo)
	err := uc.Execute(context.Background(), tx.ID, uuid.New())
	if err != entity.ErrTransactionAlreadyVerified {
		t.Fatalf("expected ErrTransactionAlreadyVerified, got %v", err)
	}
}

func TestVerifyTransaction_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
			return nil, nil
		},
	}
	uc := transaction.NewVerifyTransactionUseCase(txRepo)
	err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd backend && go test ./tests/unit/ -run "TestRecordPayment|TestRecordRefund|TestGetTransaction|TestListTransactions|TestVerifyTransaction" -v
```

Expected: Compilation errors — use case constructors not found.

- [ ] **Step 3: Implement RecordPayment use case**

Create `backend/internal/usecase/transaction/record_payment.go`:

```go
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
```

- [ ] **Step 4: Implement RecordRefund use case**

Create `backend/internal/usecase/transaction/record_refund.go`:

```go
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
```

- [ ] **Step 5: Implement GetTransaction use case**

Create `backend/internal/usecase/transaction/get_transaction.go`:

```go
package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type GetTransactionUseCase struct {
	repo repository.TransactionRepository
}

func NewGetTransactionUseCase(repo repository.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{repo: repo}
}

func (uc *GetTransactionUseCase) Execute(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, apperror.NewNotFound("Transaction", id)
	}
	return tx, nil
}
```

- [ ] **Step 6: Implement ListTransactions use case**

Create `backend/internal/usecase/transaction/list_transactions.go`:

```go
package transaction

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListTransactionsUseCase struct {
	repo repository.TransactionRepository
}

func NewListTransactionsUseCase(repo repository.TransactionRepository) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{repo: repo}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return uc.repo.List(ctx, filter)
}
```

- [ ] **Step 7: Implement VerifyTransaction use case**

Create `backend/internal/usecase/transaction/verify_transaction.go`:

```go
package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type VerifyTransactionUseCase struct {
	repo repository.TransactionRepository
}

func NewVerifyTransactionUseCase(repo repository.TransactionRepository) *VerifyTransactionUseCase {
	return &VerifyTransactionUseCase{repo: repo}
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

	return uc.repo.Update(ctx, tx)
}
```

- [ ] **Step 8: Run tests**

```bash
cd backend && go test ./tests/unit/ -v -count=1
```

Expected: All existing + ~14 new transaction use case tests PASS.

- [ ] **Step 9: Commit**

```bash
cd backend && git add internal/usecase/transaction/
cd backend && git add ../backend/tests/unit/transaction_usecase_test.go
git commit -m "feat: add transaction use cases with payment/refund orchestration and verification"
```

---

## Task 7: DTOs + Debt Handler + Tests

**Files:**
- Create: `backend/internal/delivery/http/dto/debt_dto.go`
- Create: `backend/internal/delivery/http/dto/transaction_dto.go`
- Create: `backend/internal/delivery/http/handler/debt_handler.go`
- Test: `backend/tests/unit/debt_handler_test.go`

- [ ] **Step 1: Create DTOs**

Create `backend/internal/delivery/http/dto/debt_dto.go`:

```go
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
```

Create `backend/internal/delivery/http/dto/transaction_dto.go`:

```go
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
```

- [ ] **Step 2: Implement DebtHandler**

Create `backend/internal/delivery/http/handler/debt_handler.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

type DebtHandler struct {
	createUC   *debtuc.CreateDebtUseCase
	getUC      *debtuc.GetDebtUseCase
	listUC     *debtuc.ListDebtsUseCase
	updateUC   *debtuc.UpdateDebtUseCase
	cancelUC   *debtuc.CancelDebtUseCase
	markPaidUC *debtuc.MarkDebtPaidUseCase
	deleteUC   *debtuc.DeleteDebtUseCase
}

func NewDebtHandler(
	createUC *debtuc.CreateDebtUseCase,
	getUC *debtuc.GetDebtUseCase,
	listUC *debtuc.ListDebtsUseCase,
	updateUC *debtuc.UpdateDebtUseCase,
	cancelUC *debtuc.CancelDebtUseCase,
	markPaidUC *debtuc.MarkDebtPaidUseCase,
	deleteUC *debtuc.DeleteDebtUseCase,
) *DebtHandler {
	return &DebtHandler{
		createUC: createUC, getUC: getUC, listUC: listUC,
		updateUC: updateUC, cancelUC: cancelUC, markPaidUC: markPaidUC,
		deleteUC: deleteUC,
	}
}

// Create godoc
// @Summary      Create a debt
// @Description  Creates a new debt for a tenant
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateDebtRequest  true  "Debt data"
// @Success      201   {object}  dto.DebtResponse
// @Failure      400,404,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts [post]
func (h *DebtHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.OriginalAmount)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid due_date format (use RFC3339 or YYYY-MM-DD)"))
			return
		}
	}

	d, err := entity.NewDebt(req.TenantID, landlordID, req.PropertyID, entity.DebtType(req.DebtType), req.Description, amount, dueDate, req.Notes)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	if err := h.createUC.Execute(r.Context(), d); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewDebtResponse(d))
}

// Get godoc
// @Summary      Get a debt
// @Description  Retrieves a debt by ID
// @Tags         debts
// @Produce      json
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      200  {object}  dto.DebtResponse
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [get]
func (h *DebtHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	d, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(d))
}

// List godoc
// @Summary      List debts
// @Description  Returns a paginated list of debts with optional filtering
// @Tags         debts
// @Produce      json
// @Param        tenant_id   query  string  false  "Filter by tenant ID"
// @Param        property_id query  string  false  "Filter by property ID"
// @Param        status      query  string  false  "Filter by status"
// @Param        debt_type   query  string  false  "Filter by debt type"
// @Param        search      query  string  false  "Search description"
// @Param        limit       query  int     false  "Page size (default 20, max 100)"
// @Param        offset      query  int     false  "Offset"
// @Success      200         {object}  dto.DebtListResponse
// @Router       /api/v1/debts [get]
func (h *DebtHandler) List(w http.ResponseWriter, r *http.Request) {
	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))
	filter := repository.DebtFilter{
		LandlordID: &landlordID,
		Limit:      20,
		Offset:     0,
	}

	if s := r.URL.Query().Get("tenant_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant_id filter"))
			return
		}
		filter.TenantID = &id
	}
	if s := r.URL.Query().Get("property_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid property_id filter"))
			return
		}
		filter.PropertyID = &id
	}
	if s := r.URL.Query().Get("status"); s != "" {
		status := entity.DebtStatus(s)
		filter.Status = &status
	}
	if s := r.URL.Query().Get("debt_type"); s != "" {
		dt := entity.DebtType(s)
		filter.DebtType = &dt
	}
	if s := r.URL.Query().Get("search"); s != "" {
		filter.Search = &s
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Limit = v
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Offset = v
		}
	}

	debts, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtListResponse(debts, total, filter.Limit, filter.Offset))
}

// Update godoc
// @Summary      Update a debt
// @Description  Updates mutable fields of an existing debt
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Debt ID (UUID)"
// @Param        body  body      dto.UpdateDebtRequest  true  "Updated debt data"
// @Success      200   {object}  dto.DebtResponse
// @Failure      400,404,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [put]
func (h *DebtHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	var req dto.UpdateDebtRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid due_date format"))
			return
		}
	}

	updates := &entity.Debt{
		Description: req.Description,
		DebtType:    entity.DebtType(req.DebtType),
		DueDate:     dueDate,
		PropertyID:  req.PropertyID,
		Notes:       req.Notes,
	}

	if err := h.updateUC.Execute(r.Context(), id, updates); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// Cancel godoc
// @Summary      Cancel a debt
// @Description  Cancels a debt with an optional reason
// @Tags         debts
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Debt ID (UUID)"
// @Param        body  body      dto.CancelDebtRequest  false "Cancel reason"
// @Success      200   {object}  dto.DebtResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id}/cancel [put]
func (h *DebtHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	var req dto.CancelDebtRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // optional body

	if err := h.cancelUC.Execute(r.Context(), id, req.Reason); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// MarkPaid godoc
// @Summary      Mark debt as paid
// @Description  Forces a debt to PAID status by paying the remaining balance
// @Tags         debts
// @Produce      json
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      200  {object}  dto.DebtResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id}/pay [put]
func (h *DebtHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	if err := h.markPaidUC.Execute(r.Context(), id); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	updated, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewDebtResponse(updated))
}

// Delete godoc
// @Summary      Delete a debt
// @Description  Soft-deletes a debt by ID
// @Tags         debts
// @Param        id   path      string  true  "Debt ID (UUID)"
// @Success      204
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/debts/{id} [delete]
func (h *DebtHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid debt ID format"))
		return
	}

	if err := h.deleteUC.Execute(r.Context(), id); err != nil {
		handleDebtDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleDebtDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	switch {
	case errors.Is(err, entity.ErrDebtDescriptionRequired),
		errors.Is(err, entity.ErrDebtDescriptionTooLong),
		errors.Is(err, entity.ErrDebtAmountRequired),
		errors.Is(err, entity.ErrDebtInvalidType),
		errors.Is(err, entity.ErrDebtDueDateRequired),
		errors.Is(err, entity.ErrCurrencyMismatch),
		errors.Is(err, entity.ErrNegativeAmount),
		errors.Is(err, entity.ErrInvalidCurrency):
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
	case errors.Is(err, entity.ErrDebtAlreadyPaid),
		errors.Is(err, entity.ErrDebtAlreadyCancelled),
		errors.Is(err, entity.ErrDebtOverpayment),
		errors.Is(err, entity.ErrDebtNotPayable):
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
	default:
		slog.Error("unhandled error in debt handler", "error", err)
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}
```

- [ ] **Step 3: Write Debt handler tests**

Create `backend/tests/unit/debt_handler_test.go`:

```go
package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newDebtHandler(debtRepo *mocks.DebtRepositoryMock, userRepo *mocks.UserRepositoryMock, tenantRepo *mocks.TenantRepositoryMock, propRepo *mocks.PropertyRepositoryMock) *handler.DebtHandler {
	createUC := debtuc.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propRepo)
	getUC := debtuc.NewGetDebtUseCase(debtRepo)
	listUC := debtuc.NewListDebtsUseCase(debtRepo)
	updateUC := debtuc.NewUpdateDebtUseCase(debtRepo)
	cancelUC := debtuc.NewCancelDebtUseCase(debtRepo)
	markPaidUC := debtuc.NewMarkDebtPaidUseCase(debtRepo)
	deleteUC := debtuc.NewDeleteDebtUseCase(debtRepo)
	return handler.NewDebtHandler(createUC, getUC, listUC, updateUC, cancelUC, markPaidUC, deleteUC)
}

func newTestDebtEntity() *entity.Debt {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	d, _ := entity.NewDebt(uuid.New(), uuid.New(), nil, entity.DebtTypeRent, "Monthly rent", amount, time.Now().Add(30*24*time.Hour), nil)
	return d
}

func TestDebtHandler_Create_Success(t *testing.T) {
	landlordID := uuid.New()
	tenantID := uuid.New()

	debtRepo := &mocks.DebtRepositoryMock{
		CreateFn: func(ctx context.Context, d *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: landlordID, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}
	propRepo := &mocks.PropertyRepositoryMock{}

	h := newDebtHandler(debtRepo, userRepo, tenantRepo, propRepo)
	body := map[string]any{
		"tenant_id":       tenantID.String(),
		"debt_type":       "RENT",
		"description":     "Monthly rent",
		"original_amount": map[string]any{"amount": "1000.00", "currency": "PHP"},
		"due_date":        time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02"),
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader(b))
	req = withAuthContext(req, landlordID, []string{"landlord"})
	w := httptest.NewRecorder()

	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Create_InvalidJSON(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader([]byte("bad")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDebtHandler_Create_ValidationFailed(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	body := map[string]any{"description": ""} // missing required fields
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/debts", bytes.NewReader(b))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Get_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
			return d, nil
		},
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/"+d.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Get_InvalidID(t *testing.T) {
	h := newDebtHandler(&mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/not-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDebtHandler_Get_NotFound(t *testing.T) {
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return nil, nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	r := chi.NewRouter()
	r.Get("/api/v1/debts/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDebtHandler_List_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			return []*entity.Debt{d}, 1, nil
		},
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/debts?limit=10", nil)
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Cancel_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		UpdateFn:  func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	body := map[string]any{"reason": "No longer valid"}
	b, _ := json.Marshal(body)

	r := chi.NewRouter()
	r.Put("/api/v1/debts/{id}/cancel", h.Cancel)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/debts/"+d.ID.String()+"/cancel", bytes.NewReader(b))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDebtHandler_Delete_Success(t *testing.T) {
	d := newTestDebtEntity()
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		DeleteFn:  func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	h := newDebtHandler(debtRepo, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{}, &mocks.PropertyRepositoryMock{})

	r := chi.NewRouter()
	r.Delete("/api/v1/debts/{id}", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/debts/"+d.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./tests/unit/ -run "TestDebtHandler_" -v
```

Expected: All ~10 debt handler tests PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/delivery/http/dto/debt_dto.go internal/delivery/http/dto/transaction_dto.go
cd backend && git add internal/delivery/http/handler/debt_handler.go
cd backend && git add ../backend/tests/unit/debt_handler_test.go
git commit -m "feat: add debt DTOs, handler with 7 endpoints, and handler tests"
```

---

## Task 8: Transaction Handler + Tests

**Files:**
- Create: `backend/internal/delivery/http/handler/transaction_handler.go`
- Test: `backend/tests/unit/transaction_handler_test.go`

- [ ] **Step 1: Implement TransactionHandler**

Create `backend/internal/delivery/http/handler/transaction_handler.go`:

```go
package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	txuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/validator"
)

type TransactionHandler struct {
	recordPaymentUC *txuc.RecordPaymentUseCase
	recordRefundUC  *txuc.RecordRefundUseCase
	getUC           *txuc.GetTransactionUseCase
	listUC          *txuc.ListTransactionsUseCase
	verifyUC        *txuc.VerifyTransactionUseCase
}

func NewTransactionHandler(
	recordPaymentUC *txuc.RecordPaymentUseCase,
	recordRefundUC *txuc.RecordRefundUseCase,
	getUC *txuc.GetTransactionUseCase,
	listUC *txuc.ListTransactionsUseCase,
	verifyUC *txuc.VerifyTransactionUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		recordPaymentUC: recordPaymentUC,
		recordRefundUC:  recordRefundUC,
		getUC:           getUC,
		listUC:          listUC,
		verifyUC:        verifyUC,
	}
}

// RecordPayment godoc
// @Summary      Record a payment
// @Description  Records a payment transaction against a debt
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RecordPaymentRequest  true  "Payment data"
// @Success      201   {object}  dto.TransactionResponse
// @Failure      400,404,409,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/payment [post]
func (h *TransactionHandler) RecordPayment(w http.ResponseWriter, r *http.Request) {
	var req dto.RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	recorderID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.Amount)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	txDate, err := time.Parse(time.RFC3339, req.TransactionDate)
	if err != nil {
		txDate, err = time.Parse("2006-01-02", req.TransactionDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction_date format"))
			return
		}
	}

	tx, err := h.recordPaymentUC.Execute(r.Context(), req.DebtID, req.TenantID, &recorderID, amount, entity.PaymentMethod(req.PaymentMethod), txDate, req.Description, req.ReceiptNumber, req.ReferenceNumber)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// RecordRefund godoc
// @Summary      Record a refund
// @Description  Records a refund transaction against a debt
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RecordRefundRequest  true  "Refund data"
// @Success      201   {object}  dto.TransactionResponse
// @Failure      400,404,409,422  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/refund [post]
func (h *TransactionHandler) RecordRefund(w http.ResponseWriter, r *http.Request) {
	var req dto.RecordRefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}
	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	recorderID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	amount, err := dto.ToMoney(req.Amount)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	refundDate, err := time.Parse(time.RFC3339, req.RefundDate)
	if err != nil {
		refundDate, err = time.Parse("2006-01-02", req.RefundDate)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid refund_date format"))
			return
		}
	}

	tx, err := h.recordRefundUC.Execute(r.Context(), req.DebtID, req.TenantID, &recorderID, amount, entity.PaymentMethod(req.PaymentMethod), refundDate, req.Description, req.ReferenceNumber)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// Get godoc
// @Summary      Get a transaction
// @Description  Retrieves a transaction by ID
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID (UUID)"
// @Success      200  {object}  dto.TransactionResponse
// @Failure      400,404  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/{id} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction ID format"))
		return
	}

	tx, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

// List godoc
// @Summary      List transactions
// @Description  Returns a paginated list of transactions with optional filtering
// @Tags         transactions
// @Produce      json
// @Param        debt_id     query  string  false  "Filter by debt ID"
// @Param        tenant_id   query  string  false  "Filter by tenant ID"
// @Param        type        query  string  false  "Filter by transaction type"
// @Param        is_verified query  bool    false  "Filter by verification status"
// @Param        limit       query  int     false  "Page size (default 20, max 100)"
// @Param        offset      query  int     false  "Offset"
// @Success      200         {object}  dto.TransactionListResponse
// @Router       /api/v1/transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	landlordID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))
	filter := repository.TransactionFilter{
		LandlordID: &landlordID,
		Limit:      20,
		Offset:     0,
	}

	if s := r.URL.Query().Get("debt_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid debt_id filter"))
			return
		}
		filter.DebtID = &id
	}
	if s := r.URL.Query().Get("tenant_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			apperror.WriteError(w, apperror.NewBadRequest("Invalid tenant_id filter"))
			return
		}
		filter.TenantID = &id
	}
	if s := r.URL.Query().Get("type"); s != "" {
		txType := entity.TransactionType(s)
		filter.Type = &txType
	}
	if s := r.URL.Query().Get("is_verified"); s != "" {
		v := s == "true"
		filter.IsVerified = &v
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Limit = v
		}
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			filter.Offset = v
		}
	}

	txs, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionListResponse(txs, total, filter.Limit, filter.Offset))
}

// Verify godoc
// @Summary      Verify a transaction
// @Description  Marks a transaction as verified by the authenticated user
// @Tags         transactions
// @Produce      json
// @Param        id   path      string  true  "Transaction ID (UUID)"
// @Success      200  {object}  dto.TransactionResponse
// @Failure      400,404,409  {object}  apperror.ErrorResponse
// @Router       /api/v1/transactions/{id}/verify [put]
func (h *TransactionHandler) Verify(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid transaction ID format"))
		return
	}

	verifierID, _ := uuid.Parse(r.Context().Value(middleware.AuthUserIDKey).(string))

	if err := h.verifyUC.Execute(r.Context(), id, verifierID); err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	tx, err := h.getUC.Execute(r.Context(), id)
	if err != nil {
		handleTransactionDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewTransactionResponse(tx))
}

func handleTransactionDomainError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		apperror.WriteError(w, appErr)
		return
	}

	switch {
	case errors.Is(err, entity.ErrTransactionAmountRequired),
		errors.Is(err, entity.ErrTransactionInvalidType),
		errors.Is(err, entity.ErrTransactionInvalidPaymentMethod),
		errors.Is(err, entity.ErrTransactionDateRequired),
		errors.Is(err, entity.ErrCurrencyMismatch),
		errors.Is(err, entity.ErrNegativeAmount),
		errors.Is(err, entity.ErrInvalidCurrency),
		errors.Is(err, entity.ErrDebtDescriptionRequired):
		apperror.WriteError(w, apperror.NewValidation(err.Error()))
	case errors.Is(err, entity.ErrTransactionAlreadyVerified),
		errors.Is(err, entity.ErrTransactionDuplicateReference),
		errors.Is(err, entity.ErrDebtAlreadyPaid),
		errors.Is(err, entity.ErrDebtAlreadyCancelled),
		errors.Is(err, entity.ErrDebtOverpayment),
		errors.Is(err, entity.ErrInsufficientAmount):
		apperror.WriteError(w, apperror.NewConflict(err.Error()))
	default:
		slog.Error("unhandled error in transaction handler", "error", err)
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}
```

- [ ] **Step 2: Write Transaction handler tests**

Create `backend/tests/unit/transaction_handler_test.go`:

```go
package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	txuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTransactionHandler(txRepo *mocks.TransactionRepositoryMock, debtRepo *mocks.DebtRepositoryMock, userRepo *mocks.UserRepositoryMock, tenantRepo *mocks.TenantRepositoryMock) *handler.TransactionHandler {
	recordPaymentUC := txuc.NewRecordPaymentUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	recordRefundUC := txuc.NewRecordRefundUseCase(txRepo, debtRepo, userRepo, tenantRepo)
	getUC := txuc.NewGetTransactionUseCase(txRepo)
	listUC := txuc.NewListTransactionsUseCase(txRepo)
	verifyUC := txuc.NewVerifyTransactionUseCase(txRepo)
	return handler.NewTransactionHandler(recordPaymentUC, recordRefundUC, getUC, listUC, verifyUC)
}

func TestTransactionHandler_RecordPayment_Success(t *testing.T) {
	tenantID := uuid.New()
	landlordID := uuid.New()
	debtID := uuid.New()

	amount, _ := entity.NewMoney(decimal.NewFromFloat(1000), entity.CurrencyPHP)
	d, _ := entity.NewDebt(tenantID, landlordID, nil, entity.DebtTypeRent, "Rent", amount, time.Now().Add(30*24*time.Hour), nil)
	d.ID = debtID

	txRepo := &mocks.TransactionRepositoryMock{
		CreateFn: func(ctx context.Context, tx *entity.Transaction) error { return nil },
		ExistsByReferenceNumberFn: func(ctx context.Context, dID uuid.UUID, ref string) (bool, error) {
			return false, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Debt, error) { return d, nil },
		UpdateFn:  func(ctx context.Context, debt *entity.Debt) error { return nil },
	}
	userRepo := &mocks.UserRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
			return &entity.User{ID: id, IsActive: true}, nil
		},
	}
	tenantRepo := &mocks.TenantRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
			return &entity.Tenant{ID: tenantID, LandlordID: landlordID, IsActive: true}, nil
		},
	}

	h := newTransactionHandler(txRepo, debtRepo, userRepo, tenantRepo)
	body := map[string]any{
		"debt_id":          debtID.String(),
		"tenant_id":        tenantID.String(),
		"amount":           map[string]any{"amount": "300.00", "currency": "PHP"},
		"payment_method":   "CASH",
		"transaction_date": time.Now().Format("2006-01-02"),
		"description":      "Partial payment",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/payment", bytes.NewReader(b))
	req = withAuthContext(req, landlordID, []string{"landlord"})
	w := httptest.NewRecorder()
	h.RecordPayment(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_RecordPayment_InvalidJSON(t *testing.T) {
	h := newTransactionHandler(&mocks.TransactionRepositoryMock{}, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/payment", bytes.NewReader([]byte("bad")))
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.RecordPayment(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTransactionHandler_Get_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/transactions/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/"+tx.ID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Get_NotFound(t *testing.T) {
	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return nil, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Get("/api/v1/transactions/{id}", h.Get)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestTransactionHandler_List_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
			return []*entity.Transaction{tx}, 1, nil
		},
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?limit=10", nil)
	req = withAuthContext(req, uuid.New(), []string{"landlord"})
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Verify_Success(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
		UpdateFn:  func(ctx context.Context, t *entity.Transaction) error { return nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Put("/api/v1/transactions/{id}/verify", h.Verify)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+tx.ID.String()+"/verify", nil)
	req = withAuthContext(req, uuid.New(), []string{"admin"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTransactionHandler_Verify_AlreadyVerified(t *testing.T) {
	amount, _ := entity.NewMoney(decimal.NewFromFloat(100), entity.CurrencyPHP)
	tx, _ := entity.NewTransaction(uuid.New(), uuid.New(), uuid.New(), nil, entity.TransactionTypePayment, amount, entity.PaymentMethodCash, time.Now(), "P", nil, nil)
	_ = tx.Verify(uuid.New())

	txRepo := &mocks.TransactionRepositoryMock{
		GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) { return tx, nil },
	}
	h := newTransactionHandler(txRepo, &mocks.DebtRepositoryMock{}, &mocks.UserRepositoryMock{}, &mocks.TenantRepositoryMock{})

	r := chi.NewRouter()
	r.Put("/api/v1/transactions/{id}/verify", h.Verify)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/transactions/"+tx.ID.String()+"/verify", nil)
	req = withAuthContext(req, uuid.New(), []string{"admin"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd backend && go test ./tests/unit/ -run "TestTransactionHandler_" -v
```

Expected: All ~8 transaction handler tests PASS.

- [ ] **Step 4: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: All ~230 tests PASS.

- [ ] **Step 5: Commit**

```bash
cd backend && git add internal/delivery/http/handler/transaction_handler.go
cd backend && git add ../backend/tests/unit/transaction_handler_test.go
git commit -m "feat: add transaction handler with 5 endpoints and handler tests"
```

---

## Task 9: PG Repos + Router + Main.go Wiring

**Files:**
- Create: `backend/internal/infrastructure/persistence/pg/debt_repo_pg.go`
- Create: `backend/internal/infrastructure/persistence/pg/transaction_repo_pg.go`
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Implement DebtRepoPG**

Create `backend/internal/infrastructure/persistence/pg/debt_repo_pg.go`:

```go
package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type DebtRepoPG struct {
	pool *pgxpool.Pool
}

func NewDebtRepoPG(pool *pgxpool.Pool) *DebtRepoPG {
	return &DebtRepoPG{pool: pool}
}

const debtColumns = `id, tenant_id, landlord_id, property_id, debt_type, description,
	original_amount, original_currency, amount_paid, amount_paid_currency,
	due_date, status, notes, created_at, updated_at, deleted_at`

func (r *DebtRepoPG) Create(ctx context.Context, d *entity.Debt) error {
	query := `
		INSERT INTO debts (id, tenant_id, landlord_id, property_id, debt_type, description,
			original_amount, original_currency, amount_paid, amount_paid_currency,
			due_date, status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.pool.Exec(ctx, query,
		d.ID, d.TenantID, d.LandlordID, d.PropertyID, string(d.DebtType), d.Description,
		d.OriginalAmount.Amount, string(d.OriginalAmount.Currency),
		d.AmountPaid.Amount, string(d.AmountPaid.Currency),
		d.DueDate, string(d.Status), d.Notes, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *DebtRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error) {
	query := `SELECT ` + debtColumns + ` FROM debts WHERE id = $1 AND deleted_at IS NULL`
	return r.scanDebt(r.pool.QueryRow(ctx, query, id))
}

func (r *DebtRepoPG) List(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, "d.deleted_at IS NULL")

	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("d.tenant_id = $%d", argIdx))
		args = append(args, *filter.TenantID)
		argIdx++
	}
	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("d.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}
	if filter.PropertyID != nil {
		conditions = append(conditions, fmt.Sprintf("d.property_id = $%d", argIdx))
		args = append(args, *filter.PropertyID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.DebtType != nil {
		conditions = append(conditions, fmt.Sprintf("d.debt_type = $%d", argIdx))
		args = append(args, string(*filter.DebtType))
		argIdx++
	}
	if filter.IsOverdue != nil && *filter.IsOverdue {
		conditions = append(conditions, "d.due_date < NOW()")
		conditions = append(conditions, "d.status NOT IN ('PAID', 'CANCELLED')")
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("d.description ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM debts d " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT d.id, d.tenant_id, d.landlord_id, d.property_id, d.debt_type, d.description,
			d.original_amount, d.original_currency, d.amount_paid, d.amount_paid_currency,
			d.due_date, d.status, d.notes, d.created_at, d.updated_at, d.deleted_at
		FROM debts d %s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var debts []*entity.Debt
	for rows.Next() {
		d, err := r.scanDebtRow(rows)
		if err != nil {
			return nil, 0, err
		}
		debts = append(debts, d)
	}

	return debts, total, rows.Err()
}

func (r *DebtRepoPG) Update(ctx context.Context, d *entity.Debt) error {
	query := `
		UPDATE debts
		SET debt_type = $1, description = $2,
			original_amount = $3, original_currency = $4,
			amount_paid = $5, amount_paid_currency = $6,
			due_date = $7, status = $8, notes = $9,
			property_id = $10
		WHERE id = $11 AND deleted_at IS NULL`

	_, err := r.pool.Exec(ctx, query,
		string(d.DebtType), d.Description,
		d.OriginalAmount.Amount, string(d.OriginalAmount.Currency),
		d.AmountPaid.Amount, string(d.AmountPaid.Currency),
		d.DueDate, string(d.Status), d.Notes,
		d.PropertyID, d.ID,
	)
	return err
}

func (r *DebtRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE debts SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	return err
}

func (r *DebtRepoPG) HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM debts WHERE property_id = $1 AND status NOT IN ('PAID', 'CANCELLED') AND deleted_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, propertyID).Scan(&exists)
	return exists, err
}

func (r *DebtRepoPG) scanDebt(row pgx.Row) (*entity.Debt, error) {
	d := &entity.Debt{}
	var origAmount decimal.Decimal
	var origCurrency string
	var paidAmount decimal.Decimal
	var paidCurrency string
	var debtType, status string

	err := row.Scan(
		&d.ID, &d.TenantID, &d.LandlordID, &d.PropertyID, &debtType, &d.Description,
		&origAmount, &origCurrency, &paidAmount, &paidCurrency,
		&d.DueDate, &status, &d.Notes, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	d.DebtType = entity.DebtType(debtType)
	d.Status = entity.DebtStatus(status)
	d.OriginalAmount = entity.Money{Amount: origAmount, Currency: entity.Currency(origCurrency)}
	d.AmountPaid = entity.Money{Amount: paidAmount, Currency: entity.Currency(paidCurrency)}

	return d, nil
}

func (r *DebtRepoPG) scanDebtRow(rows pgx.Rows) (*entity.Debt, error) {
	d := &entity.Debt{}
	var origAmount decimal.Decimal
	var origCurrency string
	var paidAmount decimal.Decimal
	var paidCurrency string
	var debtType, status string

	err := rows.Scan(
		&d.ID, &d.TenantID, &d.LandlordID, &d.PropertyID, &debtType, &d.Description,
		&origAmount, &origCurrency, &paidAmount, &paidCurrency,
		&d.DueDate, &status, &d.Notes, &d.CreatedAt, &d.UpdatedAt, &d.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	d.DebtType = entity.DebtType(debtType)
	d.Status = entity.DebtStatus(status)
	d.OriginalAmount = entity.Money{Amount: origAmount, Currency: entity.Currency(origCurrency)}
	d.AmountPaid = entity.Money{Amount: paidAmount, Currency: entity.Currency(paidCurrency)}

	return d, nil
}
```

- [ ] **Step 2: Implement TransactionRepoPG**

Create `backend/internal/infrastructure/persistence/pg/transaction_repo_pg.go`:

```go
package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type TransactionRepoPG struct {
	pool *pgxpool.Pool
}

func NewTransactionRepoPG(pool *pgxpool.Pool) *TransactionRepoPG {
	return &TransactionRepoPG{pool: pool}
}

const transactionColumns = `id, debt_id, tenant_id, landlord_id, recorded_by_user_id,
	transaction_type, amount, currency, payment_method, transaction_date,
	description, receipt_number, reference_number,
	is_verified, verified_by_user_id, verified_at, created_at, updated_at`

func (r *TransactionRepoPG) Create(ctx context.Context, tx *entity.Transaction) error {
	query := `
		INSERT INTO transactions (id, debt_id, tenant_id, landlord_id, recorded_by_user_id,
			transaction_type, amount, currency, payment_method, transaction_date,
			description, receipt_number, reference_number,
			is_verified, verified_by_user_id, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.DebtID, tx.TenantID, tx.LandlordID, tx.RecordedByUserID,
		string(tx.TransactionType), tx.Amount.Amount, string(tx.Amount.Currency),
		string(tx.PaymentMethod), tx.TransactionDate,
		tx.Description, tx.ReceiptNumber, tx.ReferenceNumber,
		tx.IsVerified, tx.VerifiedByUserID, tx.VerifiedAt,
		tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

func (r *TransactionRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	query := `SELECT ` + transactionColumns + ` FROM transactions WHERE id = $1`
	return r.scanTransaction(r.pool.QueryRow(ctx, query, id))
}

func (r *TransactionRepoPG) List(ctx context.Context, filter repository.TransactionFilter) ([]*entity.Transaction, int, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.DebtID != nil {
		conditions = append(conditions, fmt.Sprintf("t.debt_id = $%d", argIdx))
		args = append(args, *filter.DebtID)
		argIdx++
	}
	if filter.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("t.tenant_id = $%d", argIdx))
		args = append(args, *filter.TenantID)
		argIdx++
	}
	if filter.LandlordID != nil {
		conditions = append(conditions, fmt.Sprintf("t.landlord_id = $%d", argIdx))
		args = append(args, *filter.LandlordID)
		argIdx++
	}
	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("t.transaction_type = $%d", argIdx))
		args = append(args, string(*filter.Type))
		argIdx++
	}
	if filter.IsVerified != nil {
		conditions = append(conditions, fmt.Sprintf("t.is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM transactions t " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT t.id, t.debt_id, t.tenant_id, t.landlord_id, t.recorded_by_user_id,
			t.transaction_type, t.amount, t.currency, t.payment_method, t.transaction_date,
			t.description, t.receipt_number, t.reference_number,
			t.is_verified, t.verified_by_user_id, t.verified_at, t.created_at, t.updated_at
		FROM transactions t %s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []*entity.Transaction
	for rows.Next() {
		tx, err := r.scanTransactionRow(rows)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}

	return txs, total, rows.Err()
}

func (r *TransactionRepoPG) Update(ctx context.Context, tx *entity.Transaction) error {
	query := `
		UPDATE transactions
		SET is_verified = $1, verified_by_user_id = $2, verified_at = $3
		WHERE id = $4`

	_, err := r.pool.Exec(ctx, query,
		tx.IsVerified, tx.VerifiedByUserID, tx.VerifiedAt, tx.ID,
	)
	return err
}

func (r *TransactionRepoPG) ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE debt_id = $1 AND reference_number = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, debtID, refNum).Scan(&exists)
	return exists, err
}

func (r *TransactionRepoPG) scanTransaction(row pgx.Row) (*entity.Transaction, error) {
	tx := &entity.Transaction{}
	var txType, currency, method string
	var amount decimal.Decimal

	err := row.Scan(
		&tx.ID, &tx.DebtID, &tx.TenantID, &tx.LandlordID, &tx.RecordedByUserID,
		&txType, &amount, &currency, &method, &tx.TransactionDate,
		&tx.Description, &tx.ReceiptNumber, &tx.ReferenceNumber,
		&tx.IsVerified, &tx.VerifiedByUserID, &tx.VerifiedAt,
		&tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	tx.TransactionType = entity.TransactionType(txType)
	tx.PaymentMethod = entity.PaymentMethod(method)
	tx.Amount = entity.Money{Amount: amount, Currency: entity.Currency(currency)}

	return tx, nil
}

func (r *TransactionRepoPG) scanTransactionRow(rows pgx.Rows) (*entity.Transaction, error) {
	tx := &entity.Transaction{}
	var txType, currency, method string
	var amount decimal.Decimal

	err := rows.Scan(
		&tx.ID, &tx.DebtID, &tx.TenantID, &tx.LandlordID, &tx.RecordedByUserID,
		&txType, &amount, &currency, &method, &tx.TransactionDate,
		&tx.Description, &tx.ReceiptNumber, &tx.ReferenceNumber,
		&tx.IsVerified, &tx.VerifiedByUserID, &tx.VerifiedAt,
		&tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	tx.TransactionType = entity.TransactionType(txType)
	tx.PaymentMethod = entity.PaymentMethod(method)
	tx.Amount = entity.Money{Amount: amount, Currency: entity.Currency(currency)}

	return tx, nil
}
```

- [ ] **Step 3: Update router.go**

Add `Debt` and `Transaction` fields to `Handlers` struct, add route groups:

```go
// In the Handlers struct, add:
Debt        *handler.DebtHandler
Transaction *handler.TransactionHandler

// In the route registration, after the Properties block, add:

// Debts (authenticated, landlord/admin)
r.Route("/debts", func(r chi.Router) {
    r.Use(requireAuth)
    r.Use(middleware.RequireRole("admin", "landlord"))
    r.Get("/", h.Debt.List)
    r.Get("/{id}", h.Debt.Get)
    r.Post("/", h.Debt.Create)
    r.Put("/{id}", h.Debt.Update)
    r.Put("/{id}/pay", h.Debt.MarkPaid)
    r.Put("/{id}/cancel", h.Debt.Cancel)
    r.Delete("/{id}", h.Debt.Delete)
})

// Transactions (authenticated, landlord/admin)
r.Route("/transactions", func(r chi.Router) {
    r.Use(requireAuth)
    r.Use(middleware.RequireRole("admin", "landlord"))
    r.Post("/payment", h.Transaction.RecordPayment)
    r.Post("/refund", h.Transaction.RecordRefund)
    r.Get("/{id}", h.Transaction.Get)
    r.Get("/", h.Transaction.List)
    r.Put("/{id}/verify", h.Transaction.Verify)
})
```

- [ ] **Step 4: Update main.go**

After the Property module wiring block, add debt + transaction wiring:

```go
// Wire Debt module
debtRepo := pg.NewDebtRepoPG(pgPool)
createDebtUC := debtuc.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propertyRepo)
getDebtUC := debtuc.NewGetDebtUseCase(debtRepo)
listDebtsUC := debtuc.NewListDebtsUseCase(debtRepo)
updateDebtUC := debtuc.NewUpdateDebtUseCase(debtRepo)
cancelDebtUC := debtuc.NewCancelDebtUseCase(debtRepo)
markDebtPaidUC := debtuc.NewMarkDebtPaidUseCase(debtRepo)
deleteDebtUC := debtuc.NewDeleteDebtUseCase(debtRepo)
debtHandler := handler.NewDebtHandler(createDebtUC, getDebtUC, listDebtsUC, updateDebtUC, cancelDebtUC, markDebtPaidUC, deleteDebtUC)

// Wire Transaction module
transactionRepo := pg.NewTransactionRepoPG(pgPool)
recordPaymentUC := transactionuc.NewRecordPaymentUseCase(transactionRepo, debtRepo, userRepo, tenantRepo)
recordRefundUC := transactionuc.NewRecordRefundUseCase(transactionRepo, debtRepo, userRepo, tenantRepo)
getTransactionUC := transactionuc.NewGetTransactionUseCase(transactionRepo)
listTransactionsUC := transactionuc.NewListTransactionsUseCase(transactionRepo)
verifyTransactionUC := transactionuc.NewVerifyTransactionUseCase(transactionRepo)
transactionHandler := handler.NewTransactionHandler(recordPaymentUC, recordRefundUC, getTransactionUC, listTransactionsUC, verifyTransactionUC)
```

**IMPORTANT — wiring order in main.go:** The `debtRepo` must be created BEFORE the property module block because `DeactivatePropertyUseCase` now depends on it. The final wiring order should be:

```go
// Wire Debt repo (needed by Property deactivation)
debtRepo := pg.NewDebtRepoPG(pgPool)

// Wire Property module (uses debtRepo for deactivation constraint)
propertyRepo := pg.NewPropertyRepoPG(pgPool)
createPropertyUC := propertyuc.NewCreatePropertyUseCase(propertyRepo, userRepo)
getPropertyUC := propertyuc.NewGetPropertyUseCase(propertyRepo)
listPropertiesUC := propertyuc.NewListPropertiesUseCase(propertyRepo)
updatePropertyUC := propertyuc.NewUpdatePropertyUseCase(propertyRepo)
deactivatePropertyUC := propertyuc.NewDeactivatePropertyUseCase(propertyRepo, debtRepo) // <-- now takes debtRepo
deletePropertyUC := propertyuc.NewDeletePropertyUseCase(propertyRepo)
propertyHandler := handler.NewPropertyHandler(createPropertyUC, getPropertyUC, listPropertiesUC, updatePropertyUC, deactivatePropertyUC, deletePropertyUC)

// Wire Debt use cases (debtRepo already declared above)
createDebtUC := debtuc.NewCreateDebtUseCase(debtRepo, userRepo, tenantRepo, propertyRepo)
// ... rest of debt use cases ...
```

Add the new handler fields to the `router.Handlers` struct initialization:

```go
r := router.NewRouter(router.Handlers{
    Health:      healthHandler,
    Program:     programHandler,
    Auth:        authHandler,
    User:        userHandler,
    Beneficiary: beneficiaryHandler,
    Tenant:      tenantHandler,
    Property:    propertyHandler,
    Debt:        debtHandler,
    Transaction: transactionHandler,
}, cfg.App.FrontendURL, jwtManager, tokenBlocklist)
```

Add imports:

```go
debtuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/debt"
transactionuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/transaction"
```

- [ ] **Step 5: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: Clean compilation.

- [ ] **Step 6: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: All ~230 tests PASS.

- [ ] **Step 7: Commit**

```bash
cd backend && git add internal/infrastructure/persistence/pg/debt_repo_pg.go internal/infrastructure/persistence/pg/transaction_repo_pg.go
cd backend && git add internal/delivery/http/router/router.go cmd/server/main.go
git commit -m "feat: add PG repos, wire debt/transaction modules in router and main"
```

---

## Task 10: CLAUDE.md Update + Final Verification

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md routes table**

Add Debts and Transactions routes to the API Routes table:

```markdown
| `/debts` | GET, GET/:id, POST, PUT/:id, DELETE/:id | Required | admin, landlord |
| `/debts/{id}/pay` | PUT | Required | admin, landlord |
| `/debts/{id}/cancel` | PUT | Required | admin, landlord |
| `/transactions/payment` | POST | Required | admin, landlord |
| `/transactions/refund` | POST | Required | admin, landlord |
| `/transactions` | GET, GET/:id | Required | admin, landlord |
| `/transactions/{id}/verify` | PUT | Required | admin, landlord |
```

Update the "Current Modules" section to include Debts & Transactions:

```markdown
Programs, Users & Auth (JWT + RBAC), Beneficiaries (with program enrollment), Tenants (landlord-scoped CRUD with Address value object, deactivation), Properties (landlord-scoped with deactivation), Debts & Transactions (Money value object, debt state machine, payment/refund orchestration, lazy overdue detection, transaction verification).
```

- [ ] **Step 2: Run full test suite one final time**

```bash
cd backend && go test ./tests/... -v -count=1 2>&1 | tail -20
```

Expected: All ~230 tests PASS (10 money + 12 debt entity + 6 transaction entity + 18 debt UC + 14 transaction UC + 10 debt handler + 8 transaction handler + 152 existing = ~230).

- [ ] **Step 3: Verify build**

```bash
cd backend && go build ./cmd/server/main.go
```

Expected: Clean build.

- [ ] **Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with debts and transactions module"
```
