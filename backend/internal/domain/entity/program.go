package entity

import (
	"time"

	"github.com/google/uuid"
)

// ProgramStatus represents the possible statuses of a program.
type ProgramStatus string

const (
	ProgramStatusActive   ProgramStatus = "active"
	ProgramStatusInactive ProgramStatus = "inactive"
	ProgramStatusClosed   ProgramStatus = "closed"
)

// Program represents a social protection program.
type Program struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      ProgramStatus `json:"status"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DeletedAt   *time.Time    `json:"deleted_at,omitempty"`
}

// NewProgram creates a new Program with generated ID and timestamps.
func NewProgram(name, description string, status ProgramStatus, startDate, endDate *time.Time) (*Program, error) {
	p := &Program{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Status:      status,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// Validate checks business rules for a Program.
func (p *Program) Validate() error {
	if p.Name == "" {
		return ErrProgramNameRequired
	}
	if len(p.Name) > 255 {
		return ErrProgramNameTooLong
	}
	if !p.Status.IsValid() {
		return ErrProgramInvalidStatus
	}
	if p.StartDate != nil && p.EndDate != nil && p.EndDate.Before(*p.StartDate) {
		return ErrProgramEndBeforeStart
	}
	return nil
}

// IsValid checks if a ProgramStatus value is valid.
func (s ProgramStatus) IsValid() bool {
	switch s {
	case ProgramStatusActive, ProgramStatusInactive, ProgramStatusClosed:
		return true
	}
	return false
}

// String returns the string representation of ProgramStatus.
func (s ProgramStatus) String() string {
	return string(s)
}
