package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
)

// TransactionFilter holds optional filters for listing transactions.
type TransactionFilter struct {
	DebtID     *uuid.UUID
	TenantID   *uuid.UUID
	LandlordID *uuid.UUID
	Type       *entity.TransactionType
	IsVerified *bool
	Limit      int
	Offset     int
}

// TransactionRepository defines the interface for transaction persistence operations.
type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	List(ctx context.Context, filter TransactionFilter) ([]*entity.Transaction, int, error)
	Update(ctx context.Context, tx *entity.Transaction) error
	ExistsByReferenceNumber(ctx context.Context, debtID uuid.UUID, refNum string) (bool, error)
}
