package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// CreateProgramRequest is the request body for creating a program.
type CreateProgramRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Status      string  `json:"status" validate:"required,oneof=active inactive closed"`
	StartDate   *string `json:"start_date" validate:"omitempty"`
	EndDate     *string `json:"end_date" validate:"omitempty"`
}

// UpdateProgramRequest is the request body for updating a program.
type UpdateProgramRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Status      string  `json:"status" validate:"required,oneof=active inactive closed"`
	StartDate   *string `json:"start_date" validate:"omitempty"`
	EndDate     *string `json:"end_date" validate:"omitempty"`
}

// ProgramResponse is the response body for a single program.
type ProgramResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	StartDate   *string   `json:"start_date,omitempty"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

// ProgramListResponse is the response body for a list of programs.
type ProgramListResponse struct {
	Data   []ProgramResponse `json:"data"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

const dateFormat = "2006-01-02"

// ToEntity converts a CreateProgramRequest to a domain entity.
func (r *CreateProgramRequest) ToEntity() (*entity.Program, error) {
	startDate, err := parseOptionalDate(r.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseOptionalDate(r.EndDate)
	if err != nil {
		return nil, err
	}

	return entity.NewProgram(
		r.Name,
		r.Description,
		entity.ProgramStatus(r.Status),
		startDate,
		endDate,
	)
}

// ToEntity converts an UpdateProgramRequest to a domain entity.
func (r *UpdateProgramRequest) ToEntity() (*entity.Program, error) {
	startDate, err := parseOptionalDate(r.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := parseOptionalDate(r.EndDate)
	if err != nil {
		return nil, err
	}

	p := &entity.Program{
		Name:        r.Name,
		Description: r.Description,
		Status:      entity.ProgramStatus(r.Status),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	return p, nil
}

// NewProgramResponse creates a ProgramResponse from a domain entity.
func NewProgramResponse(p *entity.Program) ProgramResponse {
	resp := ProgramResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      string(p.Status),
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
	}
	if p.StartDate != nil {
		s := p.StartDate.Format(dateFormat)
		resp.StartDate = &s
	}
	if p.EndDate != nil {
		s := p.EndDate.Format(dateFormat)
		resp.EndDate = &s
	}
	return resp
}

// NewProgramListResponse creates a ProgramListResponse from domain entities.
func NewProgramListResponse(programs []*entity.Program, total, limit, offset int) ProgramListResponse {
	data := make([]ProgramResponse, 0, len(programs))
	for _, p := range programs {
		data = append(data, NewProgramResponse(p))
	}
	return ProgramListResponse{
		Data:   data,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(dateFormat, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
