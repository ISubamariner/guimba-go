package audit

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListLandlordAuditLogsUseCase struct {
	repo repository.AuditRepository
}

func NewListLandlordAuditLogsUseCase(repo repository.AuditRepository) *ListLandlordAuditLogsUseCase {
	return &ListLandlordAuditLogsUseCase{repo: repo}
}

func (uc *ListLandlordAuditLogsUseCase) Execute(ctx context.Context, landlordID uuid.UUID, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	filter.LandlordID = &landlordID

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return uc.repo.List(ctx, filter)
}
