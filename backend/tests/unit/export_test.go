package unit

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	exportuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/export"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestExportTenants(t *testing.T) {
	landlordID := uuid.New()
	email := "john@example.com"
	phone := "+1234567890"

	mockTenant := &entity.Tenant{
		ID:          uuid.New(),
		FullName:    "John Doe",
		Email:       &email,
		PhoneNumber: &phone,
		LandlordID:  landlordID,
		IsActive:    true,
		CreatedAt:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	tenantRepo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			if *filter.LandlordID == landlordID {
				return []*entity.Tenant{mockTenant}, 1, nil
			}
			return nil, 0, nil
		},
	}

	uc := exportuc.NewExportTenantsUseCase(tenantRepo)

	var buf bytes.Buffer
	err := uc.Execute(context.Background(), landlordID, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) != 2 { // header + 1 data row
		t.Errorf("Expected 2 rows, got %d", len(records))
	}

	// Check header
	if records[0][0] != "ID" || records[0][1] != "Full Name" {
		t.Errorf("Unexpected header: %v", records[0])
	}

	// Check data
	if records[1][1] != "John Doe" {
		t.Errorf("Expected Full Name 'John Doe', got %s", records[1][1])
	}
	if records[1][2] != "john@example.com" {
		t.Errorf("Expected Email 'john@example.com', got %s", records[1][2])
	}
	if records[1][5] != "Yes" {
		t.Errorf("Expected Active 'Yes', got %s", records[1][5])
	}
}

func TestExportProperties(t *testing.T) {
	ownerID := uuid.New()

	mockProperty := &entity.Property{
		ID:                 uuid.New(),
		Name:               "Beach House",
		PropertyCode:       "BH-001",
		PropertyType:       "RESIDENTIAL",
		SizeInSqm:          150.75,
		OwnerID:            ownerID,
		IsActive:           true,
		IsAvailableForRent: false,
		CreatedAt:          time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC),
	}

	propertyRepo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			if *filter.OwnerID == ownerID {
				return []*entity.Property{mockProperty}, 1, nil
			}
			return nil, 0, nil
		},
	}

	uc := exportuc.NewExportPropertiesUseCase(propertyRepo)

	var buf bytes.Buffer
	err := uc.Execute(context.Background(), ownerID, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) != 2 { // header + 1 data row
		t.Errorf("Expected 2 rows, got %d", len(records))
	}

	// Check header
	if records[0][0] != "ID" || records[0][1] != "Name" {
		t.Errorf("Unexpected header: %v", records[0])
	}

	// Check data
	if records[1][1] != "Beach House" {
		t.Errorf("Expected Name 'Beach House', got %s", records[1][1])
	}
	if records[1][4] != "150.75" {
		t.Errorf("Expected Size (sqm) '150.75', got %s", records[1][4])
	}
	if records[1][6] != "No" {
		t.Errorf("Expected Available for Rent 'No', got %s", records[1][6])
	}
}

func TestExportDebts(t *testing.T) {
	landlordID := uuid.New()
	tenantID := uuid.New()

	origAmt, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)
	paidAmt, _ := entity.NewMoney(decimal.NewFromInt(2000), entity.CurrencyPHP)

	mockDebt := &entity.Debt{
		ID:             uuid.New(),
		TenantID:       tenantID,
		LandlordID:     landlordID,
		DebtType:       entity.DebtTypeRent,
		Description:    "Monthly rent for March",
		OriginalAmount: origAmt,
		AmountPaid:     paidAmt,
		Status:         entity.DebtStatusPartial,
		DueDate:        time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			if *filter.LandlordID == landlordID {
				return []*entity.Debt{mockDebt}, 1, nil
			}
			return nil, 0, nil
		},
	}

	uc := exportuc.NewExportDebtsUseCase(debtRepo)

	var buf bytes.Buffer
	err := uc.Execute(context.Background(), landlordID, &buf)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Parse CSV
	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	if len(records) != 2 { // header + 1 data row
		t.Errorf("Expected 2 rows, got %d", len(records))
	}

	// Check header
	if records[0][0] != "ID" || records[0][4] != "Original Amount" {
		t.Errorf("Unexpected header: %v", records[0])
	}

	// Check data
	if records[1][4] != "5000.00" {
		t.Errorf("Expected Original Amount '5000.00', got %s", records[1][4])
	}
	if records[1][5] != "2000.00" {
		t.Errorf("Expected Amount Paid '2000.00', got %s", records[1][5])
	}
	if records[1][6] != "3000.00" {
		t.Errorf("Expected Balance '3000.00', got %s", records[1][6])
	}
	if records[1][7] != "PARTIAL" {
		t.Errorf("Expected Status 'PARTIAL', got %s", records[1][7])
	}
}
