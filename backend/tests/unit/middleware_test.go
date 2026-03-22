package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
)

func TestRequestID_GeneratesID(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(middleware.RequestIDKey)
		if id == nil || id == "" {
			t.Error("expected request ID in context")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get(middleware.RequestIDHeader) == "" {
		t.Error("expected X-Request-ID header in response")
	}
}

func TestRequestID_UsesExisting(t *testing.T) {
	existingID := "test-request-id-123"

	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(middleware.RequestIDKey)
		if id != existingID {
			t.Errorf("expected request ID %q, got %q", existingID, id)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(middleware.RequestIDHeader, existingID)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get(middleware.RequestIDHeader) != existingID {
		t.Errorf("expected X-Request-ID %q, got %q", existingID, rr.Header().Get(middleware.RequestIDHeader))
	}
}

func TestRecovery_HandlesPanic(t *testing.T) {
	handler := middleware.Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestHealthResponse_AllUp(t *testing.T) {
	services := map[string]string{
		"postgres": "up",
		"mongodb":  "up",
		"redis":    "up",
	}
	resp := dto.NewHealthResponse(services)

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestHealthResponse_Degraded(t *testing.T) {
	services := map[string]string{
		"postgres": "up",
		"mongodb":  "down",
		"redis":    "up",
	}
	resp := dto.NewHealthResponse(services)

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", resp.Status)
	}
}
