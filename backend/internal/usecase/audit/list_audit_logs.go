package audit

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListAuditLogsUseCase struct {
	repo repository.AuditRepository
}

func NewListAuditLogsUseCase(repo repository.AuditRepository) *ListAuditLogsUseCase {
	return &ListAuditLogsUseCase{repo: repo}
}

func (uc *ListAuditLogsUseCase) Execute(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
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
