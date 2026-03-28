package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// DeleteTenantUseCase handles soft-deleting a tenant.
type DeleteTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

// NewDeleteTenantUseCase creates a new DeleteTenantUseCase.
func NewDeleteTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute soft-deletes a tenant by ID.
func (uc *DeleteTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": existing.FullName, "landlord_id": existing.LandlordID.String()},
	})

	return nil
}
