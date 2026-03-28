package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdatePropertyUseCase struct {
	repo      repository.PropertyRepository
	auditRepo repository.AuditRepository
}

func NewUpdatePropertyUseCase(repo repository.PropertyRepository, auditRepo repository.AuditRepository) *UpdatePropertyUseCase {
	return &UpdatePropertyUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *UpdatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID, property *entity.Property) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	property.ID = id
	property.CreatedAt = existing.CreatedAt
	property.OwnerID = existing.OwnerID

	if err := property.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, property); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": property.Name, "owner_id": property.OwnerID.String()},
	})

	return nil
}
