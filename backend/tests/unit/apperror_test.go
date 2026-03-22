package unit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *apperror.AppError
		contains string
	}{
		{
			name:     "not found error",
			err:      apperror.NewNotFound("Program", "123"),
			contains: "NOT_FOUND",
		},
		{
			name:     "validation error",
			err:      apperror.NewValidation("name is required", "name: required"),
			contains: "VALIDATION_ERROR",
		},
		{
			name:     "internal error with wrapped error",
			err:      apperror.NewInternal(errors.New("db connection failed")),
			contains: "db connection failed",
		},
		{
			name:     "unauthorized error",
			err:      apperror.NewUnauthorized("invalid token"),
			contains: "invalid token",
		},
		{
			name:     "bad request error",
			err:      apperror.NewBadRequest("missing field"),
			contains: "missing field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.Error()
			if msg == "" {
				t.Error("expected non-empty error message")
			}
			if !contains(msg, tt.contains) {
				t.Errorf("expected error to contain %q, got %q", tt.contains, msg)
			}
		})
	}
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		name   string
		err    *apperror.AppError
		status int
	}{
		{"not found", apperror.NewNotFound("User", 1), http.StatusNotFound},
		{"validation", apperror.NewValidation("bad input"), http.StatusUnprocessableEntity},
		{"unauthorized", apperror.NewUnauthorized("no token"), http.StatusUnauthorized},
		{"forbidden", apperror.NewForbidden("not allowed"), http.StatusForbidden},
		{"conflict", apperror.NewConflict("duplicate"), http.StatusConflict},
		{"internal", apperror.NewInternal(errors.New("boom")), http.StatusInternalServerError},
		{"bad request", apperror.NewBadRequest("invalid"), http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HTTPStatus != tt.status {
				t.Errorf("expected HTTP status %d, got %d", tt.status, tt.err.HTTPStatus)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	rr := httptest.NewRecorder()
	appErr := apperror.NewNotFound("Program", "abc-123")

	apperror.WriteError(rr, appErr)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var resp apperror.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Error.Code != apperror.CodeNotFound {
		t.Errorf("expected error code %q, got %q", apperror.CodeNotFound, resp.Error.Code)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	inner := errors.New("connection refused")
	appErr := apperror.NewInternal(inner)

	if !errors.Is(appErr, inner) {
		t.Error("expected Unwrap to return inner error")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
