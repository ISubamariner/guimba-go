package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// OverdueScheduler periodically checks for debts past their due date and marks them as overdue.
type OverdueScheduler struct {
	debtRepo repository.DebtRepository
	interval time.Duration
}

// NewOverdueScheduler creates a new OverdueScheduler with the given repository and check interval.
func NewOverdueScheduler(debtRepo repository.DebtRepository, interval time.Duration) *OverdueScheduler {
	return &OverdueScheduler{
		debtRepo: debtRepo,
		interval: interval,
	}
}

// Start runs the overdue check loop until the context is cancelled.
func (s *OverdueScheduler) Start(ctx context.Context) {
	slog.Info("overdue scheduler started", "interval", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.checkOverdue(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("overdue scheduler stopped")
			return
		case <-ticker.C:
			s.checkOverdue(ctx)
		}
	}
}

// checkOverdue queries all PENDING and PARTIAL debts and marks those past their due date as OVERDUE.
func (s *OverdueScheduler) checkOverdue(ctx context.Context) {
	now := time.Now().UTC()
	updated := 0

	// Check both PENDING and PARTIAL status debts
	for _, status := range []entity.DebtStatus{entity.DebtStatusPending, entity.DebtStatusPartial} {
		debts, _, err := s.debtRepo.List(ctx, repository.DebtFilter{
			Status: &status,
			Limit:  1000,
			Offset: 0,
		})
		if err != nil {
			slog.Error("overdue scheduler: failed to list debts", "status", status, "error", err)
			continue
		}

		for _, d := range debts {
			if d.DueDate.Before(now) && d.Status != entity.DebtStatusOverdue {
				d.MarkAsOverdue()
				if err := s.debtRepo.Update(ctx, d); err != nil {
					slog.Error("overdue scheduler: failed to update debt", "debt_id", d.ID, "error", err)
					continue
				}
				updated++
			}
		}
	}

	if updated > 0 {
		slog.Info("overdue scheduler: marked debts as overdue", "count", updated)
	}
}

// CheckOverdue exposes the check for testing.
func (s *OverdueScheduler) CheckOverdue(ctx context.Context) {
	s.checkOverdue(ctx)
}
