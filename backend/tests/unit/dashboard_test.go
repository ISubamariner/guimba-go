package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	dashuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/dashboard"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestGetDashboardStats(t *testing.T) {
	landlordID := uuid.New()
	ctx := context.Background()

	// Mock tenant repository: 5 active tenants
	tenantRepo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.TenantFilter) ([]*entity.Tenant, int, error) {
			if filter.LandlordID != nil && *filter.LandlordID == landlordID &&
				filter.IsActive != nil && *filter.IsActive == true {
				return nil, 5, nil
			}
			return nil, 0, nil
		},
	}

	// Mock property repository: 3 active properties
	propertyRepo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.PropertyFilter) ([]*entity.Property, int, error) {
			if filter.OwnerID != nil && *filter.OwnerID == landlordID &&
				filter.IsActive != nil && *filter.IsActive == true {
				return nil, 3, nil
			}
			return nil, 0, nil
		},
	}

	// Mock debt repository: 4 PENDING + 1 PARTIAL = 5 active, 2 overdue
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.DebtFilter) ([]*entity.Debt, int, error) {
			if filter.LandlordID != nil && *filter.LandlordID == landlordID {
				if filter.Status != nil {
					if *filter.Status == entity.DebtStatusPending {
						return nil, 4, nil
					}
					if *filter.Status == entity.DebtStatusPartial {
						return nil, 1, nil
					}
				}
				if filter.IsOverdue != nil && *filter.IsOverdue == true {
					return nil, 2, nil
				}
			}
			return nil, 0, nil
		},
	}

	uc := dashuc.NewGetStatsUseCase(tenantRepo, propertyRepo, debtRepo)
	stats, err := uc.Execute(ctx, landlordID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stats.TotalTenants != 5 {
		t.Errorf("expected 5 tenants, got %d", stats.TotalTenants)
	}
	if stats.TotalProperties != 3 {
		t.Errorf("expected 3 properties, got %d", stats.TotalProperties)
	}
	if stats.ActiveDebts != 5 {
		t.Errorf("expected 5 active debts (4 pending + 1 partial), got %d", stats.ActiveDebts)
	}
	if stats.OverdueDebts != 2 {
		t.Errorf("expected 2 overdue debts, got %d", stats.OverdueDebts)
	}
}

func TestGetRecentActivities(t *testing.T) {
	landlordID := uuid.New()
	ctx := context.Background()

	now := time.Now().UTC()

	// Mock audit repository: 2 successful entries
	auditRepo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			if filter.LandlordID != nil && *filter.LandlordID == landlordID &&
				filter.Success != nil && *filter.Success == true {
				entries := []*repository.AuditEntry{
					{
						ID:           uuid.New(),
						UserID:       landlordID,
						UserEmail:    "landlord@test.com",
						UserRole:     "landlord",
						Action:       "CREATE_TENANT",
						ResourceType: "tenant",
						ResourceID:   uuid.New(),
						IPAddress:    "127.0.0.1",
						UserAgent:    "test",
						Endpoint:     "/api/v1/tenants",
						Method:       "POST",
						StatusCode:   201,
						Success:      true,
						ErrorMessage: nil,
						Metadata: map[string]any{
							"tenant_name": "John Doe",
						},
						Timestamp: now.Add(-5 * time.Minute),
					},
					{
						ID:           uuid.New(),
						UserID:       landlordID,
						UserEmail:    "landlord@test.com",
						UserRole:     "landlord",
						Action:       "APPLY_PAYMENT",
						ResourceType: "debt",
						ResourceID:   uuid.New(),
						IPAddress:    "127.0.0.1",
						UserAgent:    "test",
						Endpoint:     "/api/v1/transactions/payment",
						Method:       "POST",
						StatusCode:   201,
						Success:      true,
						ErrorMessage: nil,
						Metadata: map[string]any{
							"tenant_name":    "Jane Smith",
							"payment_amount": "1000.00 USD",
						},
						Timestamp: now.Add(-10 * time.Minute),
					},
				}
				return entries, len(entries), nil
			}
			return nil, 0, nil
		},
	}

	uc := dashuc.NewGetRecentActivitiesUseCase(auditRepo)
	activities, err := uc.Execute(ctx, landlordID, 10)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(activities))
	}

	// Check first activity
	if activities[0].Action != "CREATE_TENANT" {
		t.Errorf("expected action CREATE_TENANT, got %s", activities[0].Action)
	}
	expectedDesc := "Added new tenant: John Doe"
	if activities[0].Description != expectedDesc {
		t.Errorf("expected description %q, got %q", expectedDesc, activities[0].Description)
	}

	// Check second activity
	if activities[1].Action != "APPLY_PAYMENT" {
		t.Errorf("expected action APPLY_PAYMENT, got %s", activities[1].Action)
	}
	expectedDesc2 := "Recorded payment: 1000.00 USD from Jane Smith"
	if activities[1].Description != expectedDesc2 {
		t.Errorf("expected description %q, got %q", expectedDesc2, activities[1].Description)
	}
}

func TestGetRecentActivities_LimitBounds(t *testing.T) {
	landlordID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name          string
		inputLimit    int
		expectedLimit int
	}{
		{"zero limit defaults to 10", 0, 10},
		{"negative limit defaults to 10", -5, 10},
		{"valid limit respected", 20, 20},
		{"exceeds max capped at 50", 100, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedLimit := 0

			auditRepo := &mocks.AuditRepositoryMock{
				ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
					capturedLimit = filter.Limit
					return nil, 0, nil
				},
			}

			uc := dashuc.NewGetRecentActivitiesUseCase(auditRepo)
			_, err := uc.Execute(ctx, landlordID, tt.inputLimit)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if capturedLimit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, capturedLimit)
			}
		})
	}
}
