package debt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdateDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewUpdateDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *UpdateDebtUseCase {
	return &UpdateDebtUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *UpdateDebtUseCase) Execute(ctx context.Context, id uuid.UUID, updates *entity.Debt) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	// Apply mutable fields only
	existing.Description = updates.Description
	existing.DebtType = updates.DebtType
	existing.DueDate = updates.DueDate
	existing.PropertyID = updates.PropertyID
	existing.Notes = updates.Notes
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": existing.LandlordID.String(), "description": existing.Description},
	})

	return nil
}
