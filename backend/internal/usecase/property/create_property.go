package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreatePropertyUseCase struct {
	repo      repository.PropertyRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
}

func NewCreatePropertyUseCase(repo repository.PropertyRepository, userRepo repository.UserRepository, auditRepo repository.AuditRepository) *CreatePropertyUseCase {
	return &CreatePropertyUseCase{repo: repo, userRepo: userRepo, auditRepo: auditRepo}
}

func (uc *CreatePropertyUseCase) Execute(ctx context.Context, property *entity.Property) error {
	if err := property.Validate(); err != nil {
		return err
	}

	owner, err := uc.userRepo.GetByID(ctx, property.OwnerID)
	if err != nil {
		return err
	}
	if owner == nil {
		return apperror.NewNotFound("User", property.OwnerID)
	}

	existing, err := uc.repo.GetByPropertyCode(ctx, property.PropertyCode)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperror.NewConflict(entity.ErrPropertyCodeExists.Error())
	}

	if err := uc.repo.Create(ctx, property); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   property.ID,
		Success:      true,
		Metadata:     map[string]any{"property_name": property.Name, "property_code": property.PropertyCode, "owner_id": property.OwnerID.String()},
	})

	return nil
}
