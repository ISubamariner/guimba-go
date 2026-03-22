package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// BeneficiaryFilter holds optional filters for listing beneficiaries.
type BeneficiaryFilter struct {
	Status    *entity.BeneficiaryStatus
	ProgramID *uuid.UUID
	Search    *string
	Limit     int
	Offset    int
}

// BeneficiaryRepository defines the interface for beneficiary persistence operations.
type BeneficiaryRepository interface {
	Create(ctx context.Context, beneficiary *entity.Beneficiary) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Beneficiary, error)
	List(ctx context.Context, filter BeneficiaryFilter) ([]*entity.Beneficiary, int, error)
	Update(ctx context.Context, beneficiary *entity.Beneficiary) error
	Delete(ctx context.Context, id uuid.UUID) error
	EnrollInProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error
	RemoveFromProgram(ctx context.Context, beneficiaryID, programID uuid.UUID) error
}
