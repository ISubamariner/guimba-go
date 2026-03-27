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
}

func NewCreateDebtUseCase(repo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, propRepo repository.PropertyRepository) *CreateDebtUseCase {
	return &CreateDebtUseCase{repo: repo, userRepo: userRepo, tenantRepo: tenantRepo, propRepo: propRepo}
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

	return uc.repo.Create(ctx, d)
}
