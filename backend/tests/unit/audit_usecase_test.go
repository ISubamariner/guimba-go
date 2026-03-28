package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTestAuditEntry() *repository.AuditEntry {
	return &repository.AuditEntry{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		UserEmail:    "admin@test.com",
		UserRole:     "admin",
		Action:       "CREATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   uuid.New(),
		Success:      true,
		StatusCode:   200,
		Timestamp:    time.Now().UTC(),
	}
}

func TestListAuditLogs_Success(t *testing.T) {
	entries := []*repository.AuditEntry{newTestAuditEntry(), newTestAuditEntry()}
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			return entries, 2, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	result, total, err := uc.Execute(context.Background(), repository.AuditFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestListAuditLogs_DefaultLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.AuditFilter{Limit: 0})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}

func TestListAuditLogs_MaxLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.AuditFilter{Limit: 500})
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

func TestListLandlordAuditLogs_ScopesFilter(t *testing.T) {
	landlordID := uuid.New()
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListLandlordAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), landlordID, repository.AuditFilter{})
	if capturedFilter.LandlordID == nil {
		t.Fatal("expected LandlordID to be set")
	}
	if *capturedFilter.LandlordID != landlordID {
		t.Errorf("expected LandlordID %s, got %s", landlordID, *capturedFilter.LandlordID)
	}
}

func TestListLandlordAuditLogs_DefaultLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListLandlordAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), uuid.New(), repository.AuditFilter{Limit: 0})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}
