package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/pkg/audit"
)

func TestAuditContext_SetsValues(t *testing.T) {
	var capturedCtx context.Context
	handler := middleware.AuditContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	info := audit.FromContext(capturedCtx)
	if info.IPAddress == "" {
		t.Error("expected IPAddress to be set")
	}
	if info.UserAgent != "TestAgent/1.0" {
		t.Errorf("expected UserAgent 'TestAgent/1.0', got '%s'", info.UserAgent)
	}
	if info.Endpoint != "/api/v1/tenants" {
		t.Errorf("expected Endpoint '/api/v1/tenants', got '%s'", info.Endpoint)
	}
	if info.Method != "POST" {
		t.Errorf("expected Method 'POST', got '%s'", info.Method)
	}
}

func TestFromContext_EmptyContext(t *testing.T) {
	info := audit.FromContext(context.Background())
	if info.IPAddress != "" {
		t.Errorf("expected empty IPAddress, got '%s'", info.IPAddress)
	}
	if info.UserEmail != "" {
		t.Errorf("expected empty UserEmail, got '%s'", info.UserEmail)
	}
	if info.Method != "" {
		t.Errorf("expected empty Method, got '%s'", info.Method)
	}
}
