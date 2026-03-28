package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeletePropertyUseCase struct {
	repo      repository.PropertyRepository
	auditRepo repository.AuditRepository
}

func NewDeletePropertyUseCase(repo repository.PropertyRepository, auditRepo repository.AuditRepository) *DeletePropertyUseCase {
	return &DeletePropertyUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeletePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": existing.Name, "owner_id": existing.OwnerID.String()},
	})

	return nil
}
