package unit

import (
	"testing"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

func TestNewProgram_Valid(t *testing.T) {
	p, err := entity.NewProgram("Test Program", "A description", entity.ProgramStatusActive, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name != "Test Program" {
		t.Errorf("expected name 'Test Program', got %q", p.Name)
	}
	if p.Status != entity.ProgramStatusActive {
		t.Errorf("expected status 'active', got %q", p.Status)
	}
	if p.ID.String() == "" {
		t.Error("expected non-empty UUID")
	}
}

func TestNewProgram_EmptyName(t *testing.T) {
	_, err := entity.NewProgram("", "desc", entity.ProgramStatusActive, nil, nil)
	if err != entity.ErrProgramNameRequired {
		t.Errorf("expected ErrProgramNameRequired, got %v", err)
	}
}

func TestNewProgram_NameTooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := entity.NewProgram(string(longName), "desc", entity.ProgramStatusActive, nil, nil)
	if err != entity.ErrProgramNameTooLong {
		t.Errorf("expected ErrProgramNameTooLong, got %v", err)
	}
}

func TestNewProgram_InvalidStatus(t *testing.T) {
	_, err := entity.NewProgram("Test", "desc", entity.ProgramStatus("invalid"), nil, nil)
	if err != entity.ErrProgramInvalidStatus {
		t.Errorf("expected ErrProgramInvalidStatus, got %v", err)
	}
}

func TestNewProgram_EndBeforeStart(t *testing.T) {
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := entity.NewProgram("Test", "desc", entity.ProgramStatusActive, &start, &end)
	if err != entity.ErrProgramEndBeforeStart {
		t.Errorf("expected ErrProgramEndBeforeStart, got %v", err)
	}
}

func TestNewProgram_ValidDates(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	p, err := entity.NewProgram("Test", "desc", entity.ProgramStatusActive, &start, &end)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.StartDate == nil || p.EndDate == nil {
		t.Error("expected start and end dates to be set")
	}
}

func TestProgramStatus_IsValid(t *testing.T) {
	tests := []struct {
		status entity.ProgramStatus
		valid  bool
	}{
		{entity.ProgramStatusActive, true},
		{entity.ProgramStatusInactive, true},
		{entity.ProgramStatusClosed, true},
		{entity.ProgramStatus("pending"), false},
		{entity.ProgramStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("ProgramStatus(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}
