package dashboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// DashboardStats holds aggregated stats for a landlord's portfolio.
type DashboardStats struct {
	TotalTenants    int
	TotalProperties int
	ActiveDebts     int
	OverdueDebts    int
}

// GetStatsUseCase retrieves dashboard statistics for a landlord.
type GetStatsUseCase struct {
	tenantRepo   repository.TenantRepository
	propertyRepo repository.PropertyRepository
	debtRepo     repository.DebtRepository
}

// NewGetStatsUseCase creates a new GetStatsUseCase.
func NewGetStatsUseCase(
	tenantRepo repository.TenantRepository,
	propertyRepo repository.PropertyRepository,
	debtRepo repository.DebtRepository,
) *GetStatsUseCase {
	return &GetStatsUseCase{
		tenantRepo:   tenantRepo,
		propertyRepo: propertyRepo,
		debtRepo:     debtRepo,
	}
}

// Execute computes stats for the given landlord.
func (uc *GetStatsUseCase) Execute(ctx context.Context, landlordID uuid.UUID) (*DashboardStats, error) {
	isActive := true

	// Count active tenants
	_, totalTenants, err := uc.tenantRepo.List(ctx, repository.TenantFilter{
		LandlordID: &landlordID,
		IsActive:   &isActive,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	// Count active properties
	_, totalProperties, err := uc.propertyRepo.List(ctx, repository.PropertyFilter{
		OwnerID:  &landlordID,
		IsActive: &isActive,
		Limit:    1,
	})
	if err != nil {
		return nil, err
	}

	// Count active debts (PENDING + PARTIAL)
	pendingStatus := entity.DebtStatusPending
	_, pendingCount, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		Status:     &pendingStatus,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	partialStatus := entity.DebtStatusPartial
	_, partialCount, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		Status:     &partialStatus,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	activeDebts := pendingCount + partialCount

	// Count overdue debts
	isOverdue := true
	_, overdueCount, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		IsOverdue:  &isOverdue,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalTenants:    totalTenants,
		TotalProperties: totalProperties,
		ActiveDebts:     activeDebts,
		OverdueDebts:    overdueCount,
	}, nil
}
