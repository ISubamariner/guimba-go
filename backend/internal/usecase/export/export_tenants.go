package export

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ExportTenantsUseCase handles CSV export of tenants.
type ExportTenantsUseCase struct {
	tenantRepo repository.TenantRepository
}

// NewExportTenantsUseCase creates a new ExportTenantsUseCase.
func NewExportTenantsUseCase(tenantRepo repository.TenantRepository) *ExportTenantsUseCase {
	return &ExportTenantsUseCase{tenantRepo: tenantRepo}
}

// Execute exports tenants to CSV format for a given landlord.
func (uc *ExportTenantsUseCase) Execute(ctx context.Context, landlordID uuid.UUID, w io.Writer) error {
	tenants, _, err := uc.tenantRepo.List(ctx, repository.TenantFilter{
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
	if err := cw.Write([]string{"ID", "Full Name", "Email", "Phone Number", "National ID", "Active", "Created At"}); err != nil {
		return err
	}

	// Write data rows
	for _, t := range tenants {
		email, phone, nid := "", "", ""
		if t.Email != nil {
			email = *t.Email
		}
		if t.PhoneNumber != nil {
			phone = *t.PhoneNumber
		}
		if t.NationalID != nil {
			nid = *t.NationalID
		}

		active := "Yes"
		if !t.IsActive {
			active = "No"
		}

		if err := cw.Write([]string{
			t.ID.String(),
			t.FullName,
			email,
			phone,
			nid,
			active,
			t.CreatedAt.Format("2006-01-02"),
		}); err != nil {
			return err
		}
	}

	return cw.Error()
}
