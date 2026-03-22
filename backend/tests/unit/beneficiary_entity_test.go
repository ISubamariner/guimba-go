package unit

import (
	"testing"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

func strPtr(s string) *string { return &s }

func TestNewBeneficiary_Valid(t *testing.T) {
	email := "john@example.com"
	b, err := entity.NewBeneficiary("John Doe", &email, nil, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if b.FullName != "John Doe" {
		t.Errorf("expected FullName 'John Doe', got %q", b.FullName)
	}
	if b.ID == [16]byte{} {
		t.Error("expected non-zero UUID")
	}
}

func TestNewBeneficiary_WithPhone(t *testing.T) {
	phone := "+639171234567"
	b, err := entity.NewBeneficiary("Jane Doe", nil, &phone, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *b.PhoneNumber != phone {
		t.Errorf("expected phone %q, got %q", phone, *b.PhoneNumber)
	}
}

func TestNewBeneficiary_NameRequired(t *testing.T) {
	email := "test@example.com"
	_, err := entity.NewBeneficiary("", &email, nil, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != entity.ErrBeneficiaryFullNameRequired {
		t.Errorf("expected ErrBeneficiaryFullNameRequired, got %v", err)
	}
}

func TestNewBeneficiary_NameTooLong(t *testing.T) {
	email := "test@example.com"
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewBeneficiary(string(longName), &email, nil, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != entity.ErrBeneficiaryFullNameTooLong {
		t.Errorf("expected ErrBeneficiaryFullNameTooLong, got %v", err)
	}
}

func TestNewBeneficiary_InvalidStatus(t *testing.T) {
	email := "test@example.com"
	_, err := entity.NewBeneficiary("John", &email, nil, nil, nil, nil, "bogus", nil)
	if err != entity.ErrBeneficiaryInvalidStatus {
		t.Errorf("expected ErrBeneficiaryInvalidStatus, got %v", err)
	}
}

func TestNewBeneficiary_ContactRequired(t *testing.T) {
	_, err := entity.NewBeneficiary("John", nil, nil, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != entity.ErrBeneficiaryContactRequired {
		t.Errorf("expected ErrBeneficiaryContactRequired, got %v", err)
	}
}

func TestNewBeneficiary_EmptyContactStrings(t *testing.T) {
	empty := ""
	_, err := entity.NewBeneficiary("John", &empty, &empty, nil, nil, nil, entity.BeneficiaryStatusActive, nil)
	if err != entity.ErrBeneficiaryContactRequired {
		t.Errorf("expected ErrBeneficiaryContactRequired with empty strings, got %v", err)
	}
}

func TestBeneficiaryStatus_IsValid(t *testing.T) {
	tests := []struct {
		status entity.BeneficiaryStatus
		valid  bool
	}{
		{entity.BeneficiaryStatusActive, true},
		{entity.BeneficiaryStatusInactive, true},
		{entity.BeneficiaryStatusSuspended, true},
		{"bogus", false},
		{"", false},
	}

	for _, tc := range tests {
		if tc.status.IsValid() != tc.valid {
			t.Errorf("status %q: expected IsValid()=%v, got %v", tc.status, tc.valid, !tc.valid)
		}
	}
}

func TestNewBeneficiary_WithAllFields(t *testing.T) {
	email := "john@example.com"
	phone := "+639171234567"
	nid := "NID-123"
	addr := "123 Main St"
	dob := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)
	notes := "VIP beneficiary"

	b, err := entity.NewBeneficiary("John Doe", &email, &phone, &nid, &addr, &dob, entity.BeneficiaryStatusActive, &notes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *b.Email != email {
		t.Errorf("email mismatch")
	}
	if *b.NationalID != nid {
		t.Errorf("national_id mismatch")
	}
	if *b.Address != addr {
		t.Errorf("address mismatch")
	}
	if !b.DateOfBirth.Equal(dob) {
		t.Errorf("dob mismatch")
	}
	if *b.Notes != notes {
		t.Errorf("notes mismatch")
	}
}
