package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// ExportPropertiesUseCase handles CSV export of properties.
type ExportPropertiesUseCase struct {
	propertyRepo repository.PropertyRepository
}

// NewExportPropertiesUseCase creates a new ExportPropertiesUseCase.
func NewExportPropertiesUseCase(propertyRepo repository.PropertyRepository) *ExportPropertiesUseCase {
	return &ExportPropertiesUseCase{propertyRepo: propertyRepo}
}

// Execute exports properties to CSV format for a given owner.
func (uc *ExportPropertiesUseCase) Execute(ctx context.Context, ownerID uuid.UUID, w io.Writer) error {
	properties, _, err := uc.propertyRepo.List(ctx, repository.PropertyFilter{
		OwnerID: &ownerID,
		Limit:   10000,
		Offset:  0,
	})
	if err != nil {
		return err
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	if err := cw.Write([]string{"ID", "Name", "Property Code", "Type", "Size (sqm)", "Active", "Available for Rent", "Created At"}); err != nil {
		return err
	}

	// Write data rows
	for _, p := range properties {
		active := "Yes"
		if !p.IsActive {
			active = "No"
		}

		availableForRent := "Yes"
		if !p.IsAvailableForRent {
			availableForRent = "No"
		}

		if err := cw.Write([]string{
			p.ID.String(),
			p.Name,
			p.PropertyCode,
			p.PropertyType,
			fmt.Sprintf("%.2f", p.SizeInSqm),
			active,
			availableForRent,
			p.CreatedAt.Format("2006-01-02"),
		}); err != nil {
			return err
		}
	}

	return cw.Error()
}
