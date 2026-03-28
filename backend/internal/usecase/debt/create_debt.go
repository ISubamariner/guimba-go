package debt

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreateDebtUseCase struct {
	repo       repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	propRepo   repository.PropertyRepository
	auditRepo  repository.AuditRepository
}

func NewCreateDebtUseCase(repo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, propRepo repository.PropertyRepository, auditRepo repository.AuditRepository) *CreateDebtUseCase {
	return &CreateDebtUseCase{repo: repo, userRepo: userRepo, tenantRepo: tenantRepo, propRepo: propRepo, auditRepo: auditRepo}
}

func (uc *CreateDebtUseCase) Execute(ctx context.Context, d *entity.Debt) error {
	if err := d.Validate(); err != nil {
		return err
	}

	// Validate landlord exists
	landlord, err := uc.userRepo.GetByID(ctx, d.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", d.LandlordID)
	}

	// Validate tenant exists and belongs to landlord
	tenant, err := uc.tenantRepo.GetByID(ctx, d.TenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return apperror.NewNotFound("Tenant", d.TenantID)
	}
	if tenant.LandlordID != d.LandlordID {
		return apperror.NewForbidden("tenant does not belong to this landlord")
	}

	// Validate property if provided
	if d.PropertyID != nil {
		prop, err := uc.propRepo.GetByID(ctx, *d.PropertyID)
		if err != nil {
			return err
		}
		if prop == nil {
			return apperror.NewNotFound("Property", *d.PropertyID)
		}
		if prop.OwnerID != d.LandlordID {
			return apperror.NewForbidden("property does not belong to this landlord")
		}
	}

	if err := uc.repo.Create(ctx, d); err != nil {
		return err
	}

	metadata := map[string]any{
		"landlord_id": d.LandlordID.String(),
		"tenant_id":   d.TenantID.String(),
		"amount":      d.OriginalAmount.Amount.String(),
		"currency":    string(d.OriginalAmount.Currency),
		"debt_type":   string(d.DebtType),
		"description": d.Description,
	}
	if d.PropertyID != nil {
		metadata["property_id"] = d.PropertyID.String()
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_DEBT",
		ResourceType: "Debt",
		ResourceID:   d.ID,
		Success:      true,
		Metadata:     metadata,
	})

	return nil
}
