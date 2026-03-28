package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// RecentActivity represents a human-readable activity log entry.
type RecentActivity struct {
	Action      string
	Description string
	Timestamp   time.Time
}

// GetRecentActivitiesUseCase retrieves recent audit activities for a landlord.
type GetRecentActivitiesUseCase struct {
	auditRepo repository.AuditRepository
}

// NewGetRecentActivitiesUseCase creates a new GetRecentActivitiesUseCase.
func NewGetRecentActivitiesUseCase(auditRepo repository.AuditRepository) *GetRecentActivitiesUseCase {
	return &GetRecentActivitiesUseCase{auditRepo: auditRepo}
}

// Execute retrieves recent activities for the given landlord.
func (uc *GetRecentActivitiesUseCase) Execute(ctx context.Context, landlordID uuid.UUID, limit int) ([]RecentActivity, error) {
	// Apply default and max limits
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	success := true
	entries, _, err := uc.auditRepo.List(ctx, repository.AuditFilter{
		LandlordID: &landlordID,
		Success:    &success,
		Limit:      limit,
	})
	if err != nil {
		return nil, err
	}

	activities := make([]RecentActivity, 0, len(entries))
	for _, entry := range entries {
		activities = append(activities, RecentActivity{
			Action:      entry.Action,
			Description: describeAction(entry.Action, entry.ResourceType, entry.Metadata),
			Timestamp:   entry.Timestamp,
		})
	}

	return activities, nil
}

// describeAction converts an audit action into a human-readable description.
func describeAction(action, resourceType string, meta map[string]any) string {
	switch action {
	case "CREATE_TENANT":
		return fmt.Sprintf("Added new tenant: %s", metaStr(meta, "tenant_name"))
	case "UPDATE_TENANT":
		return fmt.Sprintf("Updated tenant details: %s", metaStr(meta, "tenant_name"))
	case "CREATE_DEBT":
		return fmt.Sprintf("Created debt record: %s for %s", metaStr(meta, "amount"), metaStr(meta, "tenant_name"))
	case "APPLY_PAYMENT":
		return fmt.Sprintf("Recorded payment: %s from %s", metaStr(meta, "payment_amount"), metaStr(meta, "tenant_name"))
	case "MARK_DEBT_PAID":
		return fmt.Sprintf("Marked debt as paid for %s", metaStr(meta, "tenant_name"))
	case "CANCEL_DEBT":
		return fmt.Sprintf("Cancelled debt for %s", metaStr(meta, "tenant_name"))
	case "CREATE_PROPERTY":
		return fmt.Sprintf("Added new property: %s", metaStr(meta, "property_name"))
	case "UPDATE_PROPERTY":
		return fmt.Sprintf("Updated property: %s", metaStr(meta, "property_name"))
	case "DEACTIVATE_PROPERTY":
		return fmt.Sprintf("Deactivated property: %s", metaStr(meta, "property_name"))
	case "DEACTIVATE_TENANT":
		return fmt.Sprintf("Deactivated tenant: %s", metaStr(meta, "tenant_name"))
	case "RECORD_REFUND":
		return fmt.Sprintf("Recorded refund: %s for %s", metaStr(meta, "refund_amount"), metaStr(meta, "tenant_name"))
	case "VERIFY_TRANSACTION":
		return "Verified transaction"
	default:
		return fmt.Sprintf("%s on %s", action, resourceType)
	}
}

// metaStr safely extracts a string from metadata.
func metaStr(meta map[string]any, key string) string {
	if val, ok := meta[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
