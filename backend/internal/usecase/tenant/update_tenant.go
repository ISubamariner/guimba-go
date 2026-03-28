package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// UpdateTenantUseCase handles updating an existing tenant.
type UpdateTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

// NewUpdateTenantUseCase creates a new UpdateTenantUseCase.
func NewUpdateTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{repo: repo, auditRepo: auditRepo}
}

// Execute updates a tenant after verifying it exists.
func (uc *UpdateTenantUseCase) Execute(ctx context.Context, id uuid.UUID, tenant *entity.Tenant) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	tenant.ID = id
	tenant.CreatedAt = existing.CreatedAt
	tenant.LandlordID = existing.LandlordID

	if err := tenant.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, tenant); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": tenant.FullName, "landlord_id": tenant.LandlordID.String()},
	})

	return nil
}
