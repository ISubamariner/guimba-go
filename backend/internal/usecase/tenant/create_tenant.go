package tenant

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

// CreateTenantUseCase handles creating a new tenant.
type CreateTenantUseCase struct {
	repo      repository.TenantRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
}

// NewCreateTenantUseCase creates a new CreateTenantUseCase.
func NewCreateTenantUseCase(repo repository.TenantRepository, userRepo repository.UserRepository, auditRepo repository.AuditRepository) *CreateTenantUseCase {
	return &CreateTenantUseCase{repo: repo, userRepo: userRepo, auditRepo: auditRepo}
}

// Execute creates a new tenant after validating the landlord exists and email is unique.
func (uc *CreateTenantUseCase) Execute(ctx context.Context, tenant *entity.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}

	// Validate landlord exists
	landlord, err := uc.userRepo.GetByID(ctx, tenant.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", tenant.LandlordID)
	}

	// Check email uniqueness among tenants
	if tenant.Email != nil && *tenant.Email != "" {
		existing, err := uc.repo.GetByEmail(ctx, *tenant.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return apperror.NewConflict(entity.ErrTenantEmailExists.Error())
		}
	}

	if err := uc.repo.Create(ctx, tenant); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   tenant.ID,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": tenant.FullName, "landlord_id": tenant.LandlordID.String()},
	})

	return nil
}
