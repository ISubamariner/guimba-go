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
