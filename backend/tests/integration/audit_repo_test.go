//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/mongo"
)

func newTestAuditEntry(userID uuid.UUID, action string) *repository.AuditEntry {
	resourceID := uuid.New()
	errorMsg := "test error"
	return &repository.AuditEntry{
		ID:           uuid.New(),
		UserID:       userID,
		UserEmail:    "test@example.com",
		UserRole:     "admin",
		Action:       action,
		ResourceType: "test_resource",
		ResourceID:   resourceID,
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
		Endpoint:     "/api/v1/test",
		Method:       "POST",
		StatusCode:   200,
		Success:      true,
		ErrorMessage: &errorMsg,
		Metadata: map[string]any{
			"test_key": "test_value",
		},
		Timestamp: time.Now().UTC(),
	}
}

func TestAuditRepo_Log_And_List(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs collection
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()
	entry := newTestAuditEntry(userID, "create")

	// Log the entry
	err := repo.Log(ctx, entry)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// List all entries
	entries, total, err := repo.List(ctx, repository.AuditFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// Verify entry fields
	found := entries[0]
	if found.UserID != userID {
		t.Errorf("expected user_id %v, got %v", userID, found.UserID)
	}
	if found.Action != "create" {
		t.Errorf("expected action 'create', got %q", found.Action)
	}
	if found.ResourceType != "test_resource" {
		t.Errorf("expected resource_type 'test_resource', got %q", found.ResourceType)
	}
	if !found.Success {
		t.Error("expected success true, got false")
	}
}

func TestAuditRepo_List_FilterByUserID(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	user1 := uuid.New()
	user2 := uuid.New()

	// Create entries for user1
	for i := 0; i < 3; i++ {
		if err := repo.Log(ctx, newTestAuditEntry(user1, "create")); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// Create entries for user2
	for i := 0; i < 2; i++ {
		if err := repo.Log(ctx, newTestAuditEntry(user2, "update")); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// Filter by user1
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		UserID: &user1,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3 for user1, got %d", total)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries for user1, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.UserID != user1 {
			t.Errorf("expected user_id %v, got %v", user1, entry.UserID)
		}
	}
}

func TestAuditRepo_List_FilterByAction(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()

	// Create entries with different actions
	actions := []string{"create", "update", "delete", "create"}
	for _, action := range actions {
		if err := repo.Log(ctx, newTestAuditEntry(userID, action)); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// Filter by "create" action
	createAction := "create"
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		Action: &createAction,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2 for 'create' action, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for 'create' action, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.Action != "create" {
			t.Errorf("expected action 'create', got %q", entry.Action)
		}
	}
}

func TestAuditRepo_List_FilterByResourceType(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()

	// Create entries with different resource types
	entry1 := newTestAuditEntry(userID, "create")
	entry1.ResourceType = "program"
	if err := repo.Log(ctx, entry1); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entry2 := newTestAuditEntry(userID, "update")
	entry2.ResourceType = "beneficiary"
	if err := repo.Log(ctx, entry2); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entry3 := newTestAuditEntry(userID, "delete")
	entry3.ResourceType = "program"
	if err := repo.Log(ctx, entry3); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Filter by "program" resource type
	resourceType := "program"
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		ResourceType: &resourceType,
		Limit:        10,
		Offset:       0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2 for 'program' resource type, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for 'program' resource type, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.ResourceType != "program" {
			t.Errorf("expected resource_type 'program', got %q", entry.ResourceType)
		}
	}
}

func TestAuditRepo_List_FilterBySuccess(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()

	// Create successful entry
	successEntry := newTestAuditEntry(userID, "create")
	successEntry.Success = true
	if err := repo.Log(ctx, successEntry); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Create failed entries
	for i := 0; i < 2; i++ {
		failedEntry := newTestAuditEntry(userID, "update")
		failedEntry.Success = false
		failedEntry.StatusCode = 500
		if err := repo.Log(ctx, failedEntry); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// Filter by success = false
	successFalse := false
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		Success: &successFalse,
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2 for failed entries, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 failed entries, got %d", len(entries))
	}
	for _, entry := range entries {
		if entry.Success {
			t.Error("expected success false, got true")
		}
		if entry.StatusCode != 500 {
			t.Errorf("expected status code 500, got %d", entry.StatusCode)
		}
	}
}

func TestAuditRepo_List_FilterByDateRange(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()
	now := time.Now().UTC()

	// Create entries with different timestamps
	entry1 := newTestAuditEntry(userID, "create")
	entry1.Timestamp = now.Add(-2 * time.Hour)
	if err := repo.Log(ctx, entry1); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entry2 := newTestAuditEntry(userID, "update")
	entry2.Timestamp = now.Add(-1 * time.Hour)
	if err := repo.Log(ctx, entry2); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	entry3 := newTestAuditEntry(userID, "delete")
	entry3.Timestamp = now
	if err := repo.Log(ctx, entry3); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Filter by date range (last 90 minutes)
	fromDate := now.Add(-90 * time.Minute)
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		FromDate: &fromDate,
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2 in date range, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries in date range, got %d", len(entries))
	}
}

func TestAuditRepo_List_WithPagination(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()

	// Create 7 entries
	for i := 0; i < 7; i++ {
		if err := repo.Log(ctx, newTestAuditEntry(userID, "create")); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// First page (limit 3)
	entries, total, err := repo.List(ctx, repository.AuditFilter{Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 7 {
		t.Errorf("expected total 7, got %d", total)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries on page 1, got %d", len(entries))
	}

	// Second page
	entries, total, err = repo.List(ctx, repository.AuditFilter{Limit: 3, Offset: 3})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 7 {
		t.Errorf("expected total 7, got %d", total)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries on page 2, got %d", len(entries))
	}

	// Third page (partial)
	entries, total, err = repo.List(ctx, repository.AuditFilter{Limit: 3, Offset: 6})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 7 {
		t.Errorf("expected total 7, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry on page 3, got %d", len(entries))
	}
}

func TestAuditRepo_List_FilterByLandlordID(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	landlordID := uuid.New()
	otherUserID := uuid.New()

	// Entry where user_id matches landlord_id
	entry1 := newTestAuditEntry(landlordID, "create")
	if err := repo.Log(ctx, entry1); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Entry where metadata.landlord_id matches
	entry2 := newTestAuditEntry(otherUserID, "update")
	entry2.Metadata = map[string]any{"landlord_id": landlordID.String()}
	if err := repo.Log(ctx, entry2); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Entry that should not match
	entry3 := newTestAuditEntry(uuid.New(), "delete")
	if err := repo.Log(ctx, entry3); err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// Filter by landlord_id (should match both user_id and metadata.landlord_id)
	entries, total, err := repo.List(ctx, repository.AuditFilter{
		LandlordID: &landlordID,
		Limit:      10,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2 for landlord_id filter, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for landlord_id filter, got %d", len(entries))
	}
}

func TestAuditRepo_List_SortedByTimestamp(t *testing.T) {
	ctx := context.Background()
	repo := mongo.NewAuditRepoMongo(testMongoDB.Client(), testMongoDB.Name())

	// Clear audit logs
	if err := testMongoDB.Collection("audit_logs").Drop(ctx); err != nil {
		t.Logf("warning: could not drop audit_logs: %v", err)
	}

	userID := uuid.New()
	baseTime := time.Now().UTC()

	// Create entries with different timestamps
	for i := 0; i < 5; i++ {
		entry := newTestAuditEntry(userID, "create")
		entry.Timestamp = baseTime.Add(time.Duration(i) * time.Second)
		if err := repo.Log(ctx, entry); err != nil {
			t.Fatalf("Log failed: %v", err)
		}
	}

	// List entries (should be sorted by timestamp DESC)
	entries, _, err := repo.List(ctx, repository.AuditFilter{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify descending order
	for i := 0; i < len(entries)-1; i++ {
		if entries[i].Timestamp.Before(entries[i+1].Timestamp) {
			t.Errorf("entries not sorted by timestamp DESC: entry %d (%v) before entry %d (%v)",
				i, entries[i].Timestamp, i+1, entries[i+1].Timestamp)
		}
	}
}
