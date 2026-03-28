package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo      repository.PropertyRepository
	debtRepo  repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository, debtRepo repository.DebtRepository, auditRepo repository.AuditRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo, debtRepo: debtRepo, auditRepo: auditRepo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	hasDebts, err := uc.debtRepo.HasActiveDebtsForProperty(ctx, id)
	if err != nil {
		return err
	}
	if hasDebts {
		return apperror.NewConflict(entity.ErrPropertyHasActiveDebts.Error())
	}

	existing.IsActive = false
	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DEACTIVATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": existing.Name, "owner_id": existing.OwnerID.String()},
	})

	return nil
}
