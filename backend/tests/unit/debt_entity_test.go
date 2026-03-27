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
