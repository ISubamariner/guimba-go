# Debts & Transactions Module Design Spec

**Date:** 2026-03-27
**Status:** Approved
**Approach:** Faithful Port — port Debt, Transaction, and Money value object from business logic reference.

## Context

The Debts & Transactions module is the core financial system. Debts represent money owed by tenants to landlords, optionally tied to a property. Transactions are immutable records of payments and refunds that mutate debt state. A Money value object provides currency-safe decimal arithmetic.

**Completed modules:** Programs, Users & Auth, Beneficiaries, Tenants, Properties
**Build order:** ~~Tenants~~ → ~~Properties~~ → **Debts & Transactions** → Audit → Dashboard

**Key decisions:**
- Full Money value object with multi-currency support (PHP base, 7 currencies)
- Single spec covering Debt + Transaction (tightly coupled)
- Service-layer orchestration for cross-entity flows (payment recording)
- Audit logging deferred to Audit module
- Property deactivation constraint added (cannot deactivate with outstanding debts)

## 1. Money Value Object (`domain/entity/money.go`)

```go
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

type Money struct {
    Amount   decimal.Decimal // shopspring/decimal, rounded to 2dp
    Currency Currency
}
```

**Constructor:** `NewMoney(amount decimal.Decimal, currency Currency) (Money, error)` — validates amount >= 0 (zero allowed), rounds to 2dp, validates currency.

**Methods:**
- `Add(other Money) (Money, error)` — same currency required
- `Subtract(other Money) (Money, error)` — same currency, result >= 0
- `Multiply(factor decimal.Decimal) Money`
- `IsZero() bool`
- `IsGreaterThan(other Money) (bool, error)` — same currency required

**Package-level helper:** `ZeroMoney(currency Currency) Money` — creates Money(0, currency). Used for initializing AmountPaid on new debts.

**MoneyDTO helpers** (in `delivery/http/dto/`):
- `func ToMoney(dto MoneyDTO) (entity.Money, error)` — parses amount string to decimal, creates Money
- `func NewMoneyDTO(m entity.Money) MoneyDTO` — converts Money to DTO

**Validation rules:**
- All arithmetic validates same currency → `ErrCurrencyMismatch`
- Negative amounts → `ErrNegativeAmount`
- Invalid currency → `ErrInvalidCurrency`
- Subtraction resulting in negative → `ErrInsufficientAmount`

**Domain errors** (added to `backend/internal/domain/entity/errors.go`):
- `ErrCurrencyMismatch`, `ErrNegativeAmount`, `ErrInvalidCurrency`, `ErrInsufficientAmount`

**Dependency:** Add `github.com/shopspring/decimal` via `go get github.com/shopspring/decimal`.

**`ValidCurrencies`** for validation: `map[Currency]bool` with all 7 currencies.

## 2. Enums (`domain/entity/debt_enums.go`)

```go
type DebtStatus string
const (
    DebtStatusPending   DebtStatus = "PENDING"
    DebtStatusPartial   DebtStatus = "PARTIAL"
    DebtStatusPaid      DebtStatus = "PAID"
    DebtStatusOverdue   DebtStatus = "OVERDUE"
    DebtStatusCancelled DebtStatus = "CANCELLED"
)

type DebtType string
const (
    DebtTypeRent        DebtType = "RENT"
    DebtTypeUtilities   DebtType = "UTILITIES"
    DebtTypeMaintenance DebtType = "MAINTENANCE"
    DebtTypePenalty     DebtType = "PENALTY"
    DebtTypeOther       DebtType = "OTHER"
)

type TransactionType string
const (
    TransactionTypePayment    TransactionType = "PAYMENT"
    TransactionTypeRefund     TransactionType = "REFUND"
    TransactionTypePenalty    TransactionType = "PENALTY"
    TransactionTypeAdjustment TransactionType = "ADJUSTMENT"
)

type PaymentMethod string
const (
    PaymentMethodCash         PaymentMethod = "CASH"
    PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
    PaymentMethodMobileMoney  PaymentMethod = "MOBILE_MONEY"
    PaymentMethodCheck        PaymentMethod = "CHECK"
    PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
    PaymentMethodOther        PaymentMethod = "OTHER"
)
```

Each enum type has an `IsValid() bool` method.

Domain errors: `ErrInvalidDebtStatus`, `ErrInvalidDebtType`, `ErrInvalidTransactionType`, `ErrInvalidPaymentMethod`.

## 3. Debt Entity (`domain/entity/debt.go`)

```go
type Debt struct {
    ID             uuid.UUID   `json:"id"`
    TenantID       uuid.UUID   `json:"tenant_id"`
    LandlordID     uuid.UUID   `json:"landlord_id"`
    PropertyID     *uuid.UUID  `json:"property_id,omitempty"`
    DebtType       DebtType    `json:"debt_type"`
    Description    string      `json:"description"`
    OriginalAmount Money       `json:"original_amount"`
    AmountPaid     Money       `json:"amount_paid"`
    DueDate        time.Time   `json:"due_date"`
    Status         DebtStatus  `json:"status"`
    Notes          *string     `json:"notes,omitempty"`
    CreatedAt      time.Time   `json:"created_at"`
    UpdatedAt      time.Time   `json:"updated_at"`
    DeletedAt      *time.Time  `json:"deleted_at,omitempty"`
}
```

**Constructor:** `NewDebt(tenantID, landlordID uuid.UUID, propertyID *uuid.UUID, debtType DebtType, description string, originalAmount Money, dueDate time.Time, notes *string) (*Debt, error)` — sets ID, Status=PENDING, AmountPaid=Zero(same currency), validates.

**Domain methods:**

`RecordPayment(amount Money) error`:
- Validates not PAID/CANCELLED → `ErrDebtAlreadyPaid` / `ErrDebtAlreadyCancelled`
- Validates same currency → `ErrCurrencyMismatch`
- Validates amount <= GetBalance() → `ErrDebtOverpayment`
- Increments AmountPaid
- Auto-transitions: if fully paid → PAID, else → PARTIAL

`ReversePayment(amount Money) error`:
- Validates same currency
- Validates amount <= AmountPaid → `ErrInsufficientAmount`
- Decrements AmountPaid
- Recalculates: if AmountPaid == 0 → PENDING, else → PARTIAL

`MarkAsOverdue()`:
- Only transitions if status is PENDING or PARTIAL

`Cancel(reason *string) error`:
- Validates not PAID → `ErrDebtAlreadyPaid`
- Sets CANCELLED, appends reason to notes

`GetBalance() Money` — `OriginalAmount.Subtract(AmountPaid)`

`IsFullyPaid() bool` — `AmountPaid >= OriginalAmount`

`IsOverdue() bool` — `DueDate < now AND status not in [PAID, CANCELLED]`

**Validation (`Validate()`):**
- Description required, non-empty, max 500 chars
- OriginalAmount must be > 0
- DebtType must be valid
- DueDate must be non-zero

**Validation (`Validate()`) note:** DueDate must be non-zero; past dates are allowed (for recording historical debts).

**Domain errors** (added to `backend/internal/domain/entity/errors.go`):
- `ErrDebtDescriptionRequired`
- `ErrDebtDescriptionTooLong`
- `ErrDebtAmountRequired` (amount <= 0)
- `ErrDebtInvalidType`
- `ErrDebtDueDateRequired`
- `ErrDebtAlreadyPaid`
- `ErrDebtAlreadyCancelled`
- `ErrDebtOverpayment`
- `ErrDebtInvalidStateTransition`
- `ErrDebtNotPayable` (status is PAID or CANCELLED)

## 4. Transaction Entity (`domain/entity/transaction.go`)

```go
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
```

**Key rules:**
- **Immutable** — once created, only verification fields can be updated
- No `DeletedAt` — transactions are permanent records

**Constructor:** `NewTransaction(debtID, tenantID, landlordID uuid.UUID, recordedBy *uuid.UUID, txType TransactionType, amount Money, method PaymentMethod, txDate time.Time, description string, receipt, reference *string) (*Transaction, error)`

`RecordedByUserID` is optional (null for system-generated transactions).

**Methods:**
- `Verify(userID uuid.UUID) error` — validates not already verified → `ErrTransactionAlreadyVerified`
- `Validate()` — Amount > 0, valid type, valid method, date non-zero

**Domain errors** (added to `backend/internal/domain/entity/errors.go`):
- `ErrTransactionAmountRequired`
- `ErrTransactionInvalidType`
- `ErrTransactionInvalidPaymentMethod`
- `ErrTransactionDateRequired`
- `ErrTransactionAlreadyVerified`
- `ErrTransactionDuplicateReference`

## 5. Repository Interfaces

### DebtRepository (`domain/repository/debt_repository.go`)

```go
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

type DebtRepository interface {
    Create(ctx context.Context, debt *entity.Debt) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error)
    List(ctx context.Context, filter DebtFilter) ([]*entity.Debt, int, error)
    Update(ctx context.Context, debt *entity.Debt) error
    Delete(ctx context.Context, id uuid.UUID) error
    HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error)
}
```

`HasActiveDebtsForProperty`: queries `SELECT EXISTS(... WHERE property_id = $1 AND status NOT IN ('PAID', 'CANCELLED') AND deleted_at IS NULL)`.

### TransactionRepository (`domain/repository/transaction_repository.go`)

```go
type TransactionFilter struct {
    DebtID     *uuid.UUID
    TenantID   *uuid.UUID
    LandlordID *uuid.UUID
    Type       *entity.TransactionType
    IsVerified *bool
    Limit      int
    Offset     int
}

type TransactionRepository interface {
    Create(ctx context.Context, tx *entity.Transaction) error
    GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
    List(ctx context.Context, filter TransactionFilter) ([]*entity.Transaction, int, error)
    Update(ctx context.Context, tx *entity.Transaction) error
    ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error)
}
```

No `Delete` — transactions are immutable and permanent.

## 6. Use Cases

### Debt Use Cases (`usecase/debt/`)

| Use Case | Dependencies | Key Logic |
|:---|:---|:---|
| `CreateDebt` | DebtRepo, UserRepo, TenantRepo, PropertyRepo | Validates tenant exists + belongs to landlord, property belongs to landlord (if provided), amount > 0, creates with PENDING |
| `GetDebt` | DebtRepo | Gets by ID, **lazy overdue**: if `IsOverdue()` and status != OVERDUE → mark + persist. Use `SELECT ... FOR UPDATE` to prevent race conditions |
| `ListDebts` | DebtRepo | Filters + pagination, lazy overdue on all results (mark + persist individually) |
| `UpdateDebt` | DebtRepo | Existence check, preserves immutable fields (TenantID, LandlordID, AmountPaid, Status), re-validates |
| `CancelDebt` | DebtRepo | Calls `debt.Cancel(reason)`, persists |
| `MarkDebtPaid` | DebtRepo | Calculates remaining balance, calls `RecordPayment(balance)` which auto-transitions to PAID |
| `DeleteDebt` | DebtRepo | Existence check, soft delete via DeletedAt (consistent with other entities) |

### Transaction Use Cases (`usecase/transaction/`)

| Use Case | Dependencies | Key Logic |
|:---|:---|:---|
| `RecordPayment` | TransactionRepo, DebtRepo, UserRepo, TenantRepo | Full validation chain, creates Transaction(PAYMENT), calls `debt.RecordPayment()`, persists both |
| `RecordRefund` | TransactionRepo, DebtRepo, UserRepo, TenantRepo | Validates refund <= amount_paid, creates Transaction(REFUND), calls `debt.ReversePayment()`, persists both |
| `GetTransaction` | TransactionRepo | Get by ID |
| `ListTransactions` | TransactionRepo | Filters + pagination |
| `VerifyTransaction` | TransactionRepo | Calls `transaction.Verify(userID)`, persists |

### Property Deactivation Modification

Modify `DeactivatePropertyUseCase`: inject `DebtRepository`, call `HasActiveDebtsForProperty()` before deactivating. New error (add to `errors.go`): `ErrPropertyHasActiveDebts = errors.New("cannot deactivate property with active debts")`. Update constructor call in `main.go` to pass `debtRepo`.

## 7. Database Migrations

### `000009_create_debts.up.sql`

- debts table with UUID PK
- `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT`
- `landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
- `property_id UUID REFERENCES properties(id) ON DELETE SET NULL`
- `original_amount DECIMAL(14,2) NOT NULL`, `original_currency VARCHAR(3) NOT NULL DEFAULT 'PHP'`
- `amount_paid DECIMAL(14,2) NOT NULL DEFAULT 0`, `amount_paid_currency VARCHAR(3) NOT NULL DEFAULT 'PHP'`
- `debt_type VARCHAR(20) NOT NULL`, `status VARCHAR(20) NOT NULL DEFAULT 'PENDING'`
- `due_date DATE NOT NULL`, `description TEXT NOT NULL`, `notes TEXT`
- Soft delete via `deleted_at TIMESTAMPTZ`
- Auto-updated_at trigger
- Indexes: tenant_id, landlord_id, property_id, status, due_date (WHERE status NOT IN ('PAID','CANCELLED')), deleted_at

### `000010_create_transactions.up.sql`

- transactions table with UUID PK
- `debt_id UUID NOT NULL REFERENCES debts(id) ON DELETE RESTRICT`
- `tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT`
- `landlord_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT`
- `recorded_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL`
- `amount DECIMAL(14,2) NOT NULL`, `currency VARCHAR(3) NOT NULL DEFAULT 'PHP'`
- `transaction_type VARCHAR(20) NOT NULL`, `payment_method VARCHAR(20) NOT NULL`
- `transaction_date DATE NOT NULL`
- Verification fields: `is_verified BOOLEAN NOT NULL DEFAULT false`, `verified_by_user_id`, `verified_at`
- **No deleted_at** — permanent records
- Unique partial index: `(debt_id, reference_number) WHERE reference_number IS NOT NULL`
- Auto-updated_at trigger
- Indexes: debt_id, tenant_id, landlord_id, transaction_type

## 8. DTOs

- `MoneyDTO` — `amount string` (decimal as string), `currency string`
- `CreateDebtRequest` — tenant_id, property_id?, debt_type, description, original_amount (MoneyDTO), due_date, notes
- `UpdateDebtRequest` — description, debt_type, due_date, property_id?, notes (cannot change amounts/status)
- `CancelDebtRequest` — reason (optional)
- `DebtResponse` — all fields with MoneyDTO for amounts, timestamps as RFC3339
- `DebtListResponse` — array + total/limit/offset
- `RecordPaymentRequest` — debt_id, tenant_id, amount (MoneyDTO), payment_method, transaction_date, description, receipt_number?, reference_number?
- `RecordRefundRequest` — debt_id, tenant_id, amount (MoneyDTO), payment_method, refund_date, description, reference_number?
- `TransactionResponse` — all fields with MoneyDTO
- `TransactionListResponse` — array + total/limit/offset

## 9. HTTP Handlers & Routes

### DebtHandler — `/api/v1/debts`, auth + landlord/admin

| Method | Path | Handler |
|:---|:---|:---|
| POST | `/` | Create |
| GET | `/` | List |
| GET | `/{id}` | Get |
| PUT | `/{id}` | Update |
| PUT | `/{id}/pay` | MarkPaid |
| PUT | `/{id}/cancel` | Cancel |
| DELETE | `/{id}` | Delete |

### TransactionHandler — `/api/v1/transactions`, auth + landlord/admin

| Method | Path | Handler |
|:---|:---|:---|
| POST | `/payment` | RecordPayment |
| POST | `/refund` | RecordRefund |
| GET | `/{id}` | Get |
| GET | `/` | List |
| PUT | `/{id}/verify` | Verify |

## 10. Tests

| File | Estimated Count |
|:---|:---|
| `money_test.go` | ~10 tests |
| `debt_entity_test.go` | ~12 tests |
| `transaction_entity_test.go` | ~6 tests |
| `debt_usecase_test.go` | ~18 tests |
| `transaction_usecase_test.go` | ~14 tests |
| `debt_handler_test.go` | ~10 tests |
| `transaction_handler_test.go` | ~8 tests |

~78 new tests, bringing total to ~230.

## Files to Create/Modify

### New Files (37)
1. `backend/internal/domain/entity/money.go`
2. `backend/internal/domain/entity/debt_enums.go`
3. `backend/internal/domain/entity/debt.go`
4. `backend/internal/domain/entity/transaction.go`
5. `backend/internal/domain/repository/debt_repository.go`
6. `backend/internal/domain/repository/transaction_repository.go`
7. `backend/internal/usecase/debt/create_debt.go`
8. `backend/internal/usecase/debt/get_debt.go`
9. `backend/internal/usecase/debt/list_debts.go`
10. `backend/internal/usecase/debt/update_debt.go`
11. `backend/internal/usecase/debt/cancel_debt.go`
12. `backend/internal/usecase/debt/mark_debt_paid.go`
13. `backend/internal/usecase/debt/delete_debt.go`
14. `backend/internal/usecase/transaction/record_payment.go`
15. `backend/internal/usecase/transaction/record_refund.go`
16. `backend/internal/usecase/transaction/get_transaction.go`
17. `backend/internal/usecase/transaction/list_transactions.go`
18. `backend/internal/usecase/transaction/verify_transaction.go`
19. `backend/internal/delivery/http/dto/debt_dto.go`
20. `backend/internal/delivery/http/dto/transaction_dto.go`
21. `backend/internal/delivery/http/handler/debt_handler.go`
22. `backend/internal/delivery/http/handler/transaction_handler.go`
23. `backend/internal/infrastructure/persistence/pg/debt_repo_pg.go`
24. `backend/internal/infrastructure/persistence/pg/transaction_repo_pg.go`
25. `backend/migrations/000009_create_debts.up.sql`
26. `backend/migrations/000009_create_debts.down.sql`
27. `backend/migrations/000010_create_transactions.up.sql`
28. `backend/migrations/000010_create_transactions.down.sql`
29. `backend/tests/mocks/debt_repository_mock.go`
30. `backend/tests/mocks/transaction_repository_mock.go`
31. `backend/tests/unit/money_test.go`
32. `backend/tests/unit/debt_entity_test.go`
33. `backend/tests/unit/transaction_entity_test.go`
34. `backend/tests/unit/debt_usecase_test.go`
35. `backend/tests/unit/transaction_usecase_test.go`
36. `backend/tests/unit/debt_handler_test.go`
37. `backend/tests/unit/transaction_handler_test.go`

### Modified Files (4)
1. `backend/internal/domain/entity/errors.go` — add debt + transaction domain errors
2. `backend/internal/usecase/property/deactivate_property.go` — add debt check
3. `backend/internal/delivery/http/router/router.go` — add Debt + Transaction handlers + routes
4. `backend/cmd/server/main.go` — wire debt + transaction modules, update `DeactivatePropertyUseCase` to inject `DebtRepository`

## Deferred
- Audit logging (will be added in Audit module)
- OCR extracted data handling (field stored but not processed)
- Background overdue scanning (lazy detection only for now)
