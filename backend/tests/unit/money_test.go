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
