package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newAuditHandler() (*handler.AuditHandler, *mocks.AuditRepositoryMock) {
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			return []*repository.AuditEntry{newTestAuditEntry()}, 1, nil
		},
	}
	listUC := audituc.NewListAuditLogsUseCase(repo)
	listLandlordUC := audituc.NewListLandlordAuditLogsUseCase(repo)
	return handler.NewAuditHandler(listUC, listLandlordUC), repo
}

func TestAuditHandler_List_Success(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?limit=10", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp dto.AuditListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp.Data))
	}
}

func TestAuditHandler_List_InvalidFromDate(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?from_date=not-a-date", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAuditHandler_LandlordList_Success(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/landlord", nil)
	userID := uuid.New().String()
	ctx := context.WithValue(req.Context(), middleware.AuthUserIDKey, userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.LandlordList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuditHandler_LandlordList_MissingUserID(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/landlord", nil)
	rr := httptest.NewRecorder()

	h.LandlordList(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuditHandler_List_WithFilters(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	listUC := audituc.NewListAuditLogsUseCase(repo)
	listLandlordUC := audituc.NewListLandlordAuditLogsUseCase(repo)
	h := handler.NewAuditHandler(listUC, listLandlordUC)

	now := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?action=CREATE_TENANT&resource_type=Tenant&success=true&from_date="+now, nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedFilter.Action == nil || *capturedFilter.Action != "CREATE_TENANT" {
		t.Error("expected action filter to be set")
	}
	if capturedFilter.ResourceType == nil || *capturedFilter.ResourceType != "Tenant" {
		t.Error("expected resource_type filter to be set")
	}
	if capturedFilter.Success == nil || *capturedFilter.Success != true {
		t.Error("expected success filter to be set")
	}
	if capturedFilter.FromDate == nil {
		t.Error("expected from_date filter to be set")
	}
}
