package mocks

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// AuditRepositoryMock is a test mock for repository.AuditRepository.
type AuditRepositoryMock struct {
	LogFn  func(ctx context.Context, entry *repository.AuditEntry) error
	ListFn func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error)
}

func (m *AuditRepositoryMock) Log(ctx context.Context, entry *repository.AuditEntry) error {
	if m.LogFn != nil {
		return m.LogFn(ctx, entry)
	}
	return nil
}

func (m *AuditRepositoryMock) List(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}
