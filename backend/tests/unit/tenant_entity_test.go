package unit

import (
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/google/uuid"
)

func TestNewTenant_ValidWithEmail(t *testing.T) {
	email := "tenant@example.com"
	landlordID := uuid.New()
	tenant, err := entity.NewTenant("John Doe", &email, nil, nil, nil, landlordID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.FullName != "John Doe" {
		t.Errorf("expected FullName 'John Doe', got %q", tenant.FullName)
	}
	if tenant.ID == (uuid.UUID{}) {
		t.Error("expected non-zero UUID")
	}
	if tenant.LandlordID != landlordID {
		t.Error("expected landlord ID to match")
	}
	if !tenant.IsActive {
		t.Error("expected IsActive to default to true")
	}
}

func TestNewTenant_ValidWithPhone(t *testing.T) {
	phone := "+639171234567"
	tenant, err := entity.NewTenant("Jane Doe", nil, &phone, nil, nil, uuid.New(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *tenant.PhoneNumber != phone {
		t.Errorf("expected phone %q, got %q", phone, *tenant.PhoneNumber)
	}
}

func TestNewTenant_NameRequired(t *testing.T) {
	email := "test@example.com"
	_, err := entity.NewTenant("", &email, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantFullNameRequired {
		t.Errorf("expected ErrTenantFullNameRequired, got %v", err)
	}
}

func TestNewTenant_NameTooLong(t *testing.T) {
	email := "test@example.com"
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewTenant(string(longName), &email, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantFullNameTooLong {
		t.Errorf("expected ErrTenantFullNameTooLong, got %v", err)
	}
}

func TestNewTenant_ContactRequired(t *testing.T) {
	_, err := entity.NewTenant("John", nil, nil, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantContactRequired {
		t.Errorf("expected ErrTenantContactRequired, got %v", err)
	}
}

func TestNewTenant_EmptyContactStrings(t *testing.T) {
	empty := ""
	_, err := entity.NewTenant("John", &empty, &empty, nil, nil, uuid.New(), nil)
	if err != entity.ErrTenantContactRequired {
		t.Errorf("expected ErrTenantContactRequired with empty strings, got %v", err)
	}
}

func TestNewTenant_WithAddress(t *testing.T) {
	email := "test@example.com"
	addr := entity.NewAddress("123 Main St", "Guimba", "Nueva Ecija", "3115", "")
	tenant, err := entity.NewTenant("John", &email, nil, nil, addr, uuid.New(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tenant.Address == nil {
		t.Fatal("expected address to be set")
	}
	if tenant.Address.Country != "Philippines" {
		t.Errorf("expected default country 'Philippines', got %q", tenant.Address.Country)
	}
}

func TestNewTenant_WithAllFields(t *testing.T) {
	email := "john@example.com"
	phone := "+639171234567"
	nid := "NID-123"
	notes := "Good tenant"
	addr := entity.NewAddress("123 Main St", "Guimba", "Nueva Ecija", "3115", "Philippines")

	tenant, err := entity.NewTenant("John Doe", &email, &phone, &nid, addr, uuid.New(), &notes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if *tenant.Email != email {
		t.Error("email mismatch")
	}
	if *tenant.PhoneNumber != phone {
		t.Error("phone mismatch")
	}
	if *tenant.NationalID != nid {
		t.Error("national_id mismatch")
	}
	if *tenant.Notes != notes {
		t.Error("notes mismatch")
	}
	if tenant.Address.Street != "123 Main St" {
		t.Error("address street mismatch")
	}
}
