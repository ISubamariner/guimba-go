package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// DebtFilter holds optional filters for listing debts.
type DebtFilter struct {
	TenantID   *uuid.UUID
	LandlordID *uuid.UUID
	PropertyID *uuid.UUID
	Status     *entity.DebtStatus
	DebtType   *entity.DebtType
	IsOverdue  *bool
	Search     *string
	Limit      int
	Offset     int
}

// DebtRepository defines the interface for debt persistence operations.
type DebtRepository interface {
	Create(ctx context.Context, debt *entity.Debt) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Debt, error)
	List(ctx context.Context, filter DebtFilter) ([]*entity.Debt, int, error)
	Update(ctx context.Context, debt *entity.Debt) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasActiveDebtsForProperty(ctx context.Context, propertyID uuid.UUID) (bool, error)
}
