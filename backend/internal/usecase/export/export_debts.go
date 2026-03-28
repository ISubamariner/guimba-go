package export

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ExportDebtsUseCase handles CSV export of debts.
type ExportDebtsUseCase struct {
	debtRepo repository.DebtRepository
}

// NewExportDebtsUseCase creates a new ExportDebtsUseCase.
func NewExportDebtsUseCase(debtRepo repository.DebtRepository) *ExportDebtsUseCase {
	return &ExportDebtsUseCase{debtRepo: debtRepo}
}

// Execute exports debts to CSV format for a given landlord.
func (uc *ExportDebtsUseCase) Execute(ctx context.Context, landlordID uuid.UUID, w io.Writer) error {
	debts, _, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		Limit:      10000,
		Offset:     0,
	})
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	if err := cw.Write([]string{"ID", "Tenant ID", "Type", "Description", "Original Amount", "Amount Paid", "Balance", "Status", "Due Date", "Created At"}); err != nil {
		return err
	}

	// Write data rows
	for _, d := range debts {
		if err := cw.Write([]string{
			d.ID.String(),
			d.TenantID.String(),
			string(d.DebtType),
			d.Description,
			d.OriginalAmount.Amount.StringFixed(2),
			d.AmountPaid.Amount.StringFixed(2),
			d.GetBalance().Amount.StringFixed(2),
			string(d.Status),
			d.DueDate.Format("2006-01-02"),
			d.CreatedAt.Format("2006-01-02"),
		}); err != nil {
			return err
		}
	}

	return cw.Error()
}
