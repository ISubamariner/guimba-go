package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeactivateTenantUseCase handles deactivating a tenant.
type DeactivateTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

// NewDeactivateTenantUseCase creates a new DeactivateTenantUseCase.
func NewDeactivateTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *DeactivateTenantUseCase {
	return &DeactivateTenantUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute deactivates a tenant by setting IsActive to false.
func (uc *DeactivateTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	existing.IsActive = false
	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DEACTIVATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": existing.FullName, "landlord_id": existing.LandlordID.String()},
	})

	return nil
}
