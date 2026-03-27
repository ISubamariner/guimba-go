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
