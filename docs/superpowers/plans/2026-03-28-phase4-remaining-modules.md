# Phase 4 Remaining Modules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete Phase 4 of the Guimba-GO masterplan by wiring the audit module, adding change-password auth, building the dashboard module, implementing CSV data export, and adding an overdue debt scheduler.

**Architecture:** Each module follows Clean Architecture (domain -> usecase -> infrastructure -> delivery). Dashboard and Export are read-only modules that reuse existing repository interfaces. The overdue scheduler runs as a background goroutine. All new endpoints follow existing handler/DTO/usecase patterns.

**Tech Stack:** Go 1.26+, Chi v5, pgx v5, mongo-go-driver v2, go-redis v9, golang-jwt/jwt/v5, encoding/csv (stdlib)

---

## File Structure

### Task 1: Wire Audit Routes
- Modify: `backend/internal/delivery/http/router/router.go` (add Audit to Handlers struct + routes)
- Modify: `backend/cmd/server/main.go` (wire audit handler)

### Task 2: Change Password
- Modify: `backend/internal/delivery/http/dto/user_dto.go` (add ChangePasswordRequest)
- Create: `backend/internal/usecase/auth/change_password.go`
- Modify: `backend/internal/domain/repository/user_repository.go` (add UpdatePassword)
- Modify: `backend/internal/infrastructure/persistence/pg/user_repo_pg.go` (implement UpdatePassword)
- Modify: `backend/tests/mocks/user_repository_mock.go` (add UpdatePasswordFn)
- Modify: `backend/internal/delivery/http/handler/auth_handler.go` (add ChangePassword + wire UC)
- Modify: `backend/internal/delivery/http/router/router.go` (add route)
- Modify: `backend/cmd/server/main.go` (wire UC)
- Create: `backend/tests/unit/change_password_test.go`

### Task 3: Dashboard Module
- Create: `backend/internal/delivery/http/dto/dashboard_dto.go`
- Create: `backend/internal/usecase/dashboard/get_stats.go`
- Create: `backend/internal/usecase/dashboard/get_recent_activities.go`
- Create: `backend/internal/delivery/http/handler/dashboard_handler.go`
- Modify: `backend/internal/delivery/http/router/router.go` (add Dashboard to Handlers + routes)
- Modify: `backend/cmd/server/main.go` (wire dashboard)
- Create: `backend/tests/unit/dashboard_test.go`

### Task 4: Data Export
- Create: `backend/internal/usecase/export/export_tenants.go`
- Create: `backend/internal/usecase/export/export_properties.go`
- Create: `backend/internal/usecase/export/export_debts.go`
- Create: `backend/internal/delivery/http/handler/export_handler.go`
- Modify: `backend/internal/delivery/http/router/router.go` (add Export to Handlers + routes)
- Modify: `backend/cmd/server/main.go` (wire export)
- Create: `backend/tests/unit/export_test.go`

### Task 5: Overdue Debt Scheduler
- Create: `backend/internal/infrastructure/scheduler/overdue_scheduler.go`
- Modify: `backend/cmd/server/main.go` (start scheduler)
- Create: `backend/tests/unit/overdue_scheduler_test.go`

---

### Task 1: Wire Audit Routes into Router and Main

**Files:**
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Add Audit field to Handlers struct and audit routes in router.go**

In `backend/internal/delivery/http/router/router.go`, add `Audit` to the `Handlers` struct and add audit route group:

```go
// In Handlers struct, add after Transaction:
Audit       *handler.AuditHandler
```

In the `r.Route("/api/v1", ...)` block, add after the transactions route group:

```go
// Audit logs (admin/auditor for all, landlord for own)
r.Route("/audit", func(r chi.Router) {
    r.Use(requireAuth)
    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireRole("admin", "auditor"))
        r.Get("/", h.Audit.List)
    })
    r.Get("/landlord", h.Audit.LandlordList)
})
```

- [ ] **Step 2: Wire audit handler in main.go**

In `backend/cmd/server/main.go`, add audit use case imports and wiring. After the `auditRepo` setup (line ~87) and before `// Set up router`, add:

```go
// Wire Audit query module
audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
```

Add to the import block above. Then add wiring before the router:

```go
// Wire Audit query handler
listAuditLogsUC := audituc.NewListAuditLogsUseCase(auditRepo)
listLandlordAuditLogsUC := audituc.NewListLandlordAuditLogsUseCase(auditRepo)
auditHandler := handler.NewAuditHandler(listAuditLogsUC, listLandlordAuditLogsUC)
```

Add `Audit: auditHandler,` to the `router.Handlers{}` struct literal.

- [ ] **Step 3: Run tests to verify nothing is broken**

Run: `cd backend && go test ./tests/unit/... -count=1`
Expected: All tests pass (265 PASS)

- [ ] **Step 4: Build to verify compilation**

Run: `cd backend && go build ./cmd/server/`
Expected: Clean build, no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/delivery/http/router/router.go backend/cmd/server/main.go
git commit -m "feat: wire audit handler routes into router and main"
```

---

### Task 2: Change Password Endpoint

**Files:**
- Modify: `backend/internal/delivery/http/dto/user_dto.go`
- Create: `backend/internal/usecase/auth/change_password.go`
- Modify: `backend/internal/domain/repository/user_repository.go`
- Modify: `backend/internal/infrastructure/persistence/pg/user_repo_pg.go`
- Modify: `backend/tests/mocks/user_repository_mock.go`
- Modify: `backend/internal/delivery/http/handler/auth_handler.go`
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Create: `backend/tests/unit/change_password_test.go`

- [ ] **Step 1: Add ChangePasswordRequest DTO**

In `backend/internal/delivery/http/dto/user_dto.go`, add after `AssignRoleRequest`:

```go
// ChangePasswordRequest is the request body for changing a user's password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=128"`
}
```

- [ ] **Step 2: Add UpdatePassword to UserRepository interface**

In `backend/internal/domain/repository/user_repository.go`, add to the `UserRepository` interface:

```go
UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
```

- [ ] **Step 3: Implement UpdatePassword in PG repo**

In `backend/internal/infrastructure/persistence/pg/user_repo_pg.go`, add after `UpdateLastLogin`:

```go
func (r *UserRepoPG) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET hashed_password = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, hashedPassword, time.Now().UTC(), userID)
	return err
}
```

- [ ] **Step 4: Add UpdatePasswordFn to mock**

In `backend/tests/mocks/user_repository_mock.go`, add to the struct:

```go
UpdatePasswordFn func(ctx context.Context, userID uuid.UUID, hashedPassword string) error
```

And add the method:

```go
func (m *UserRepositoryMock) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	if m.UpdatePasswordFn != nil {
		return m.UpdatePasswordFn(ctx, userID, hashedPassword)
	}
	return nil
}
```

- [ ] **Step 5: Create ChangePasswordUseCase**

Create `backend/internal/usecase/auth/change_password.go`:

```go
package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
)

type ChangePasswordUseCase struct {
	userRepo repository.UserRepository
}

func NewChangePasswordUseCase(userRepo repository.UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{userRepo: userRepo}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.NewInternal(err)
	}
	if user == nil {
		return apperror.NewNotFound("User not found")
	}

	if !auth.CheckPassword(currentPassword, user.HashedPassword) {
		return apperror.NewValidation("Current password is incorrect")
	}

	if auth.CheckPassword(newPassword, user.HashedPassword) {
		return apperror.NewValidation("New password must be different from current password")
	}

	hashed, err := auth.HashPassword(newPassword)
	if err != nil {
		return apperror.NewInternal(err)
	}

	return uc.userRepo.UpdatePassword(ctx, userID, hashed)
}
```

- [ ] **Step 6: Write failing tests for ChangePasswordUseCase**

Create `backend/tests/unit/change_password_test.go`:

```go
package unit

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	authuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/auth"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
	"github.com/ISubamariner/guimba-go/backend/pkg/auth"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestChangePassword(t *testing.T) {
	ctx := context.Background()

	oldHash, _ := auth.HashPassword("OldPass123!")
	userID := uuid.New()
	testUser := &entity.User{
		ID:             userID,
		Email:          "test@example.com",
		FullName:       "Test User",
		HashedPassword: oldHash,
		IsActive:       true,
	}

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		getUserReturn   *entity.User
		getUserErr      error
		updateCalled    bool
		wantErrCode     string
	}{
		{
			name:            "success",
			currentPassword: "OldPass123!",
			newPassword:     "NewPass456!",
			getUserReturn:   testUser,
			updateCalled:    true,
		},
		{
			name:            "wrong current password",
			currentPassword: "WrongPass!",
			newPassword:     "NewPass456!",
			getUserReturn:   testUser,
			wantErrCode:     "VALIDATION_ERROR",
		},
		{
			name:            "same password",
			currentPassword: "OldPass123!",
			newPassword:     "OldPass123!",
			getUserReturn:   testUser,
			wantErrCode:     "VALIDATION_ERROR",
		},
		{
			name:          "user not found",
			currentPassword: "OldPass123!",
			newPassword:     "NewPass456!",
			getUserReturn: nil,
			wantErrCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateCalled := false
			userRepo := &mocks.UserRepositoryMock{
				GetByIDFn: func(ctx context.Context, id uuid.UUID) (*entity.User, error) {
					return tt.getUserReturn, tt.getUserErr
				},
				UpdatePasswordFn: func(ctx context.Context, uid uuid.UUID, hash string) error {
					updateCalled = true
					return nil
				},
			}

			uc := authuc.NewChangePasswordUseCase(userRepo)
			err := uc.Execute(ctx, userID, tt.currentPassword, tt.newPassword)

			if tt.wantErrCode != "" {
				if err == nil {
					t.Fatalf("expected error with code %s, got nil", tt.wantErrCode)
				}
				appErr, ok := err.(*apperror.AppError)
				if !ok {
					t.Fatalf("expected AppError, got %T", err)
				}
				if appErr.Code != tt.wantErrCode {
					t.Errorf("expected error code %s, got %s", tt.wantErrCode, appErr.Code)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.updateCalled != updateCalled {
				t.Errorf("expected UpdatePassword called=%v, got=%v", tt.updateCalled, updateCalled)
			}
		})
	}
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `cd backend && go test ./tests/unit/ -run TestChangePassword -v`
Expected: All 4 test cases PASS

- [ ] **Step 8: Add ChangePassword handler method to AuthHandler**

In `backend/internal/delivery/http/handler/auth_handler.go`:

Add `changePasswordUC` field to the struct:
```go
type AuthHandler struct {
	registerUC       *authuc.RegisterUseCase
	loginUC          *authuc.LoginUseCase
	refreshUC        *authuc.RefreshTokenUseCase
	profileUC        *authuc.GetProfileUseCase
	changePasswordUC *authuc.ChangePasswordUseCase
	jwt              *auth.JWTManager
	blocklist        *cache.TokenBlocklist
}
```

Update constructor to accept it:
```go
func NewAuthHandler(
	registerUC *authuc.RegisterUseCase,
	loginUC *authuc.LoginUseCase,
	refreshUC *authuc.RefreshTokenUseCase,
	profileUC *authuc.GetProfileUseCase,
	changePasswordUC *authuc.ChangePasswordUseCase,
	jwt *auth.JWTManager,
	blocklist *cache.TokenBlocklist,
) *AuthHandler {
	return &AuthHandler{
		registerUC:       registerUC,
		loginUC:          loginUC,
		refreshUC:        refreshUC,
		profileUC:        profileUC,
		changePasswordUC: changePasswordUC,
		jwt:              jwt,
		blocklist:        blocklist,
	}
}
```

Add the handler method after `Logout`:
```go
// ChangePassword godoc
// @Summary      Change password
// @Description  Changes the authenticated user's password
// @Tags         auth
// @Accept       json
// @Security     BearerAuth
// @Param        body  body  dto.ChangePasswordRequest  true  "Password change data"
// @Success      204  "No Content"
// @Failure      400  {object}  apperror.ErrorResponse
// @Failure      401  {object}  apperror.ErrorResponse
// @Failure      422  {object}  apperror.ErrorResponse
// @Router       /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid JSON request body"))
		return
	}

	if errs := validator.ValidateStruct(req); errs != nil {
		apperror.WriteError(w, apperror.NewValidation("Validation failed", errs...))
		return
	}

	if err := h.changePasswordUC.Execute(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		handleDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 9: Add route in router.go**

In `backend/internal/delivery/http/router/router.go`, inside the authenticated auth routes group (after `r.Post("/logout", h.Auth.Logout)`):

```go
r.Post("/change-password", h.Auth.ChangePassword)
```

- [ ] **Step 10: Wire in main.go**

In `backend/cmd/server/main.go`, add after `profileUC` creation:

```go
changePasswordUC := authuc.NewChangePasswordUseCase(userRepo)
```

Update the `NewAuthHandler` call to include it:
```go
authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshUC, profileUC, changePasswordUC, jwtManager, tokenBlocklist)
```

- [ ] **Step 11: Fix any tests broken by AuthHandler constructor change**

The existing auth handler tests likely call `NewAuthHandler` with the old signature. Search for `NewAuthHandler(` in test files and add the `changePasswordUC` parameter (can pass `nil` in tests that don't test change-password).

Run: `cd backend && go test ./tests/unit/ -count=1`
Expected: All tests pass

- [ ] **Step 12: Build to verify compilation**

Run: `cd backend && go build ./cmd/server/`
Expected: Clean build

- [ ] **Step 13: Commit**

```bash
git add backend/internal/delivery/http/dto/user_dto.go \
  backend/internal/domain/repository/user_repository.go \
  backend/internal/infrastructure/persistence/pg/user_repo_pg.go \
  backend/tests/mocks/user_repository_mock.go \
  backend/internal/usecase/auth/change_password.go \
  backend/internal/delivery/http/handler/auth_handler.go \
  backend/internal/delivery/http/router/router.go \
  backend/cmd/server/main.go \
  backend/tests/unit/change_password_test.go
git commit -m "feat: add change-password auth endpoint"
```

---

### Task 3: Dashboard Module (Stats + Recent Activities)

**Files:**
- Create: `backend/internal/delivery/http/dto/dashboard_dto.go`
- Create: `backend/internal/usecase/dashboard/get_stats.go`
- Create: `backend/internal/usecase/dashboard/get_recent_activities.go`
- Create: `backend/internal/delivery/http/handler/dashboard_handler.go`
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Create: `backend/tests/unit/dashboard_test.go`

- [ ] **Step 1: Create Dashboard DTOs**

Create `backend/internal/delivery/http/dto/dashboard_dto.go`:

```go
package dto

// DashboardStatsResponse is the response body for dashboard statistics.
type DashboardStatsResponse struct {
	TotalTenants    int `json:"total_tenants"`
	TotalProperties int `json:"total_properties"`
	ActiveDebts     int `json:"active_debts"`
	OverdueDebts    int `json:"overdue_debts"`
}

// RecentActivityResponse represents a single recent activity entry.
type RecentActivityResponse struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

// RecentActivitiesResponse is the response body for recent activities.
type RecentActivitiesResponse struct {
	Data []RecentActivityResponse `json:"data"`
}
```

- [ ] **Step 2: Create GetDashboardStatsUseCase**

Create `backend/internal/usecase/dashboard/get_stats.go`:

```go
package dashboard

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type DashboardStats struct {
	TotalTenants    int
	TotalProperties int
	ActiveDebts     int
	OverdueDebts    int
}

type GetStatsUseCase struct {
	tenantRepo   repository.TenantRepository
	propertyRepo repository.PropertyRepository
	debtRepo     repository.DebtRepository
}

func NewGetStatsUseCase(
	tenantRepo repository.TenantRepository,
	propertyRepo repository.PropertyRepository,
	debtRepo repository.DebtRepository,
) *GetStatsUseCase {
	return &GetStatsUseCase{
		tenantRepo:   tenantRepo,
		propertyRepo: propertyRepo,
		debtRepo:     debtRepo,
	}
}

func (uc *GetStatsUseCase) Execute(ctx context.Context, landlordID uuid.UUID) (*DashboardStats, error) {
	active := true

	_, totalTenants, err := uc.tenantRepo.List(ctx, repository.TenantFilter{
		LandlordID: &landlordID,
		IsActive:   &active,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	_, totalProperties, err := uc.propertyRepo.List(ctx, repository.PropertyFilter{
		OwnerID:  &landlordID,
		IsActive: &active,
		Limit:    1,
		Offset:   0,
	})
	if err != nil {
		return nil, err
	}

	pendingStatus := entity.DebtStatusPending
	_, activePending, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		Status:     &pendingStatus,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	partialStatus := entity.DebtStatusPartial
	_, activePartial, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		Status:     &partialStatus,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	overdueFlag := true
	_, overdueCount, err := uc.debtRepo.List(ctx, repository.DebtFilter{
		LandlordID: &landlordID,
		IsOverdue:  &overdueFlag,
		Limit:      1,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalTenants:    totalTenants,
		TotalProperties: totalProperties,
		ActiveDebts:     activePending + activePartial,
		OverdueDebts:    overdueCount,
	}, nil
}
```

- [ ] **Step 3: Create GetRecentActivitiesUseCase**

Create `backend/internal/usecase/dashboard/get_recent_activities.go`:

```go
package dashboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type RecentActivity struct {
	Action      string
	Description string
	Timestamp   string
}

type GetRecentActivitiesUseCase struct {
	auditRepo repository.AuditRepository
}

func NewGetRecentActivitiesUseCase(auditRepo repository.AuditRepository) *GetRecentActivitiesUseCase {
	return &GetRecentActivitiesUseCase{auditRepo: auditRepo}
}

func (uc *GetRecentActivitiesUseCase) Execute(ctx context.Context, landlordID uuid.UUID, limit int) ([]RecentActivity, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	successOnly := true
	entries, _, err := uc.auditRepo.List(ctx, repository.AuditFilter{
		LandlordID: &landlordID,
		Success:    &successOnly,
		Limit:      limit,
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}

	activities := make([]RecentActivity, 0, len(entries))
	for _, e := range entries {
		activities = append(activities, RecentActivity{
			Action:      e.Action,
			Description: describeAction(e),
			Timestamp:   e.Timestamp.Format("2006-01-02T15:04:05Z"),
		})
	}
	return activities, nil
}

func describeAction(e *repository.AuditEntry) string {
	meta := e.Metadata
	switch e.Action {
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
		return fmt.Sprintf("%s on %s", e.Action, e.ResourceType)
	}
}

func metaStr(meta map[string]any, key string) string {
	if v, ok := meta[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return "unknown"
}
```

- [ ] **Step 4: Create DashboardHandler**

Create `backend/internal/delivery/http/handler/dashboard_handler.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	dashuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/dashboard"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DashboardHandler struct {
	statsUC      *dashuc.GetStatsUseCase
	activitiesUC *dashuc.GetRecentActivitiesUseCase
}

func NewDashboardHandler(statsUC *dashuc.GetStatsUseCase, activitiesUC *dashuc.GetRecentActivitiesUseCase) *DashboardHandler {
	return &DashboardHandler{statsUC: statsUC, activitiesUC: activitiesUC}
}

// Stats godoc
// @Summary      Get dashboard statistics
// @Description  Returns portfolio statistics for the authenticated landlord
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.DashboardStatsResponse
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/dashboard/stats [get]
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	stats, err := h.statsUC.Execute(r.Context(), userID)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.DashboardStatsResponse{
		TotalTenants:    stats.TotalTenants,
		TotalProperties: stats.TotalProperties,
		ActiveDebts:     stats.ActiveDebts,
		OverdueDebts:    stats.OverdueDebts,
	})
}

// RecentActivities godoc
// @Summary      Get recent activities
// @Description  Returns recent audit log entries as human-readable activities
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "Limit (default 10, max 50)"
// @Success      200    {object}  dto.RecentActivitiesResponse
// @Failure      401    {object}  apperror.ErrorResponse
// @Router       /api/v1/dashboard/recent-activities [get]
func (h *DashboardHandler) RecentActivities(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			limit = v
		}
	}

	activities, err := h.activitiesUC.Execute(r.Context(), userID, limit)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	data := make([]dto.RecentActivityResponse, 0, len(activities))
	for _, a := range activities {
		data = append(data, dto.RecentActivityResponse{
			Action:      a.Action,
			Description: a.Description,
			Timestamp:   a.Timestamp,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.RecentActivitiesResponse{Data: data})
}
```

- [ ] **Step 5: Add Dashboard to router.go**

In `backend/internal/delivery/http/router/router.go`, add to Handlers struct:
```go
Dashboard   *handler.DashboardHandler
```

Add route group inside `/api/v1`:
```go
// Dashboard (authenticated)
r.Route("/dashboard", func(r chi.Router) {
    r.Use(requireAuth)
    r.Get("/stats", h.Dashboard.Stats)
    r.Get("/recent-activities", h.Dashboard.RecentActivities)
})
```

- [ ] **Step 6: Wire Dashboard in main.go**

Add import:
```go
dashuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/dashboard"
```

Add wiring before the router:
```go
// Wire Dashboard module
getStatsUC := dashuc.NewGetStatsUseCase(tenantRepo, propertyRepo, debtRepo)
getRecentActivitiesUC := dashuc.NewGetRecentActivitiesUseCase(auditRepo)
dashboardHandler := handler.NewDashboardHandler(getStatsUC, getRecentActivitiesUC)
```

Add `Dashboard: dashboardHandler,` to the `router.Handlers{}` struct literal.

- [ ] **Step 7: Write dashboard tests**

Create `backend/tests/unit/dashboard_test.go`:

```go
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	dashuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/dashboard"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestGetDashboardStats(t *testing.T) {
	ctx := context.Background()
	landlordID := uuid.New()

	tenantRepo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, f repository.TenantFilter) ([]*entity.Tenant, int, error) {
			return nil, 5, nil
		},
	}
	propertyRepo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, f repository.PropertyFilter) ([]*entity.Property, int, error) {
			return nil, 3, nil
		},
	}
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, f repository.DebtFilter) ([]*entity.Debt, int, error) {
			if f.IsOverdue != nil && *f.IsOverdue {
				return nil, 2, nil
			}
			if f.Status != nil && *f.Status == entity.DebtStatusPending {
				return nil, 4, nil
			}
			if f.Status != nil && *f.Status == entity.DebtStatusPartial {
				return nil, 1, nil
			}
			return nil, 0, nil
		},
	}

	uc := dashuc.NewGetStatsUseCase(tenantRepo, propertyRepo, debtRepo)
	stats, err := uc.Execute(ctx, landlordID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalTenants != 5 {
		t.Errorf("TotalTenants = %d, want 5", stats.TotalTenants)
	}
	if stats.TotalProperties != 3 {
		t.Errorf("TotalProperties = %d, want 3", stats.TotalProperties)
	}
	if stats.ActiveDebts != 5 {
		t.Errorf("ActiveDebts = %d, want 5 (4 pending + 1 partial)", stats.ActiveDebts)
	}
	if stats.OverdueDebts != 2 {
		t.Errorf("OverdueDebts = %d, want 2", stats.OverdueDebts)
	}
}

func TestGetRecentActivities(t *testing.T) {
	ctx := context.Background()
	landlordID := uuid.New()

	auditRepo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, f repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			return []*repository.AuditEntry{
				{
					ID:        uuid.New(),
					Action:    "CREATE_TENANT",
					Metadata:  map[string]any{"tenant_name": "Juan dela Cruz"},
					Timestamp: time.Now().UTC(),
				},
				{
					ID:        uuid.New(),
					Action:    "APPLY_PAYMENT",
					Metadata:  map[string]any{"payment_amount": "5000", "tenant_name": "Maria Santos"},
					Timestamp: time.Now().UTC(),
				},
			}, 2, nil
		},
	}

	uc := dashuc.NewGetRecentActivitiesUseCase(auditRepo)
	activities, err := uc.Execute(ctx, landlordID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(activities) != 2 {
		t.Fatalf("expected 2 activities, got %d", len(activities))
	}

	if activities[0].Description != "Added new tenant: Juan dela Cruz" {
		t.Errorf("unexpected description: %s", activities[0].Description)
	}
	if activities[1].Description != "Recorded payment: 5000 from Maria Santos" {
		t.Errorf("unexpected description: %s", activities[1].Description)
	}
}

func TestGetRecentActivities_LimitBounds(t *testing.T) {
	ctx := context.Background()
	landlordID := uuid.New()

	var capturedLimit int
	auditRepo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, f repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedLimit = f.Limit
			return nil, 0, nil
		},
	}

	uc := dashuc.NewGetRecentActivitiesUseCase(auditRepo)

	// Test default limit
	_, _ = uc.Execute(ctx, landlordID, 0)
	if capturedLimit != 10 {
		t.Errorf("default limit = %d, want 10", capturedLimit)
	}

	// Test max limit
	_, _ = uc.Execute(ctx, landlordID, 100)
	if capturedLimit != 50 {
		t.Errorf("max limit = %d, want 50", capturedLimit)
	}
}
```

- [ ] **Step 8: Run tests**

Run: `cd backend && go test ./tests/unit/ -run TestGetDashboard -v && go test ./tests/unit/ -run TestGetRecentActivities -v`
Expected: All tests PASS

- [ ] **Step 9: Run full test suite**

Run: `cd backend && go test ./tests/unit/... -count=1`
Expected: All tests pass

- [ ] **Step 10: Build to verify compilation**

Run: `cd backend && go build ./cmd/server/`
Expected: Clean build

- [ ] **Step 11: Commit**

```bash
git add backend/internal/delivery/http/dto/dashboard_dto.go \
  backend/internal/usecase/dashboard/ \
  backend/internal/delivery/http/handler/dashboard_handler.go \
  backend/internal/delivery/http/router/router.go \
  backend/cmd/server/main.go \
  backend/tests/unit/dashboard_test.go
git commit -m "feat: add dashboard module with stats and recent activities endpoints"
```

---

### Task 4: CSV Data Export

**Files:**
- Create: `backend/internal/usecase/export/export_tenants.go`
- Create: `backend/internal/usecase/export/export_properties.go`
- Create: `backend/internal/usecase/export/export_debts.go`
- Create: `backend/internal/delivery/http/handler/export_handler.go`
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`
- Create: `backend/tests/unit/export_test.go`

- [ ] **Step 1: Create ExportTenantsUseCase**

Create `backend/internal/usecase/export/export_tenants.go`:

```go
package export

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ExportTenantsUseCase struct {
	tenantRepo repository.TenantRepository
}

func NewExportTenantsUseCase(tenantRepo repository.TenantRepository) *ExportTenantsUseCase {
	return &ExportTenantsUseCase{tenantRepo: tenantRepo}
}

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

	cw.Write([]string{"ID", "Full Name", "Email", "Phone Number", "National ID", "Active", "Created At"})

	for _, t := range tenants {
		email := ""
		if t.Email != nil {
			email = *t.Email
		}
		phone := ""
		if t.PhoneNumber != nil {
			phone = *t.PhoneNumber
		}
		nid := ""
		if t.NationalID != nil {
			nid = *t.NationalID
		}
		active := "Yes"
		if !t.IsActive {
			active = "No"
		}
		cw.Write([]string{
			t.ID.String(),
			t.FullName,
			email,
			phone,
			nid,
			active,
			t.CreatedAt.Format("2006-01-02"),
		})
	}
	return cw.Error()
}
```

- [ ] **Step 2: Create ExportPropertiesUseCase**

Create `backend/internal/usecase/export/export_properties.go`:

```go
package export

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ExportPropertiesUseCase struct {
	propertyRepo repository.PropertyRepository
}

func NewExportPropertiesUseCase(propertyRepo repository.PropertyRepository) *ExportPropertiesUseCase {
	return &ExportPropertiesUseCase{propertyRepo: propertyRepo}
}

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

	cw.Write([]string{"ID", "Name", "Property Code", "Type", "Size (sqm)", "Active", "Available for Rent", "Created At"})

	for _, p := range properties {
		active := "Yes"
		if !p.IsActive {
			active = "No"
		}
		available := "Yes"
		if !p.IsAvailableForRent {
			available = "No"
		}
		cw.Write([]string{
			p.ID.String(),
			p.Name,
			p.PropertyCode,
			p.PropertyType,
			fmt.Sprintf("%.2f", p.SizeInSqm),
			active,
			available,
			p.CreatedAt.Format("2006-01-02"),
		})
	}
	return cw.Error()
}
```

- [ ] **Step 3: Create ExportDebtsUseCase**

Create `backend/internal/usecase/export/export_debts.go`:

```go
package export

import (
	"context"
	"encoding/csv"
	"io"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ExportDebtsUseCase struct {
	debtRepo repository.DebtRepository
}

func NewExportDebtsUseCase(debtRepo repository.DebtRepository) *ExportDebtsUseCase {
	return &ExportDebtsUseCase{debtRepo: debtRepo}
}

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

	cw.Write([]string{"ID", "Tenant ID", "Type", "Description", "Original Amount", "Amount Paid", "Balance", "Status", "Due Date", "Created At"})

	for _, d := range debts {
		cw.Write([]string{
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
		})
	}
	return cw.Error()
}
```

- [ ] **Step 4: Create ExportHandler**

Create `backend/internal/delivery/http/handler/export_handler.go`:

```go
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	exportuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/export"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type ExportHandler struct {
	tenantsUC    *exportuc.ExportTenantsUseCase
	propertiesUC *exportuc.ExportPropertiesUseCase
	debtsUC      *exportuc.ExportDebtsUseCase
}

func NewExportHandler(
	tenantsUC *exportuc.ExportTenantsUseCase,
	propertiesUC *exportuc.ExportPropertiesUseCase,
	debtsUC *exportuc.ExportDebtsUseCase,
) *ExportHandler {
	return &ExportHandler{
		tenantsUC:    tenantsUC,
		propertiesUC: propertiesUC,
		debtsUC:      debtsUC,
	}
}

// ExportTenants godoc
// @Summary      Export tenants as CSV
// @Description  Downloads all tenants for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Security     BearerAuth
// @Success      200  {file}  file
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/tenants [get]
func (h *ExportHandler) ExportTenants(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "tenants")
	if err := h.tenantsUC.Execute(r.Context(), userID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}

// ExportProperties godoc
// @Summary      Export properties as CSV
// @Description  Downloads all properties for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Security     BearerAuth
// @Success      200  {file}  file
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/properties [get]
func (h *ExportHandler) ExportProperties(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "properties")
	if err := h.propertiesUC.Execute(r.Context(), userID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}

// ExportDebts godoc
// @Summary      Export debts as CSV
// @Description  Downloads all debts for the authenticated landlord as a CSV file
// @Tags         export
// @Produce      text/csv
// @Security     BearerAuth
// @Success      200  {file}  file
// @Failure      401  {object}  apperror.ErrorResponse
// @Router       /api/v1/export/debts [get]
func (h *ExportHandler) ExportDebts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.AuthUserIDKey).(uuid.UUID)
	if !ok {
		apperror.WriteError(w, apperror.NewUnauthorized("User not authenticated"))
		return
	}

	setCSVHeaders(w, "debts")
	if err := h.debtsUC.Execute(r.Context(), userID, w); err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
	}
}

func setCSVHeaders(w http.ResponseWriter, resource string) {
	filename := fmt.Sprintf("%s_%s.csv", resource, time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
}
```

- [ ] **Step 5: Add Export to router.go**

In `backend/internal/delivery/http/router/router.go`, add to Handlers struct:
```go
Export      *handler.ExportHandler
```

Add route group inside `/api/v1`:
```go
// Export (authenticated, landlord/admin)
r.Route("/export", func(r chi.Router) {
    r.Use(requireAuth)
    r.Use(middleware.RequireRole("admin", "landlord"))
    r.Get("/tenants", h.Export.ExportTenants)
    r.Get("/properties", h.Export.ExportProperties)
    r.Get("/debts", h.Export.ExportDebts)
})
```

- [ ] **Step 6: Wire Export in main.go**

Add import:
```go
exportuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/export"
```

Add wiring:
```go
// Wire Export module
exportTenantsUC := exportuc.NewExportTenantsUseCase(tenantRepo)
exportPropertiesUC := exportuc.NewExportPropertiesUseCase(propertyRepo)
exportDebtsUC := exportuc.NewExportDebtsUseCase(debtRepo)
exportHandler := handler.NewExportHandler(exportTenantsUC, exportPropertiesUC, exportDebtsUC)
```

Add `Export: exportHandler,` to the `router.Handlers{}` struct literal.

- [ ] **Step 7: Write export tests**

Create `backend/tests/unit/export_test.go`:

```go
package unit

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	exportuc "github.com/ISubamariner/guimba-go/backend/internal/usecase/export"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestExportTenants(t *testing.T) {
	ctx := context.Background()
	landlordID := uuid.New()
	email := "juan@example.com"
	phone := "09171234567"

	tenantRepo := &mocks.TenantRepositoryMock{
		ListFn: func(ctx context.Context, f repository.TenantFilter) ([]*entity.Tenant, int, error) {
			return []*entity.Tenant{
				{
					ID:          uuid.New(),
					FullName:    "Juan dela Cruz",
					Email:       &email,
					PhoneNumber: &phone,
					LandlordID:  landlordID,
					IsActive:    true,
					CreatedAt:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				},
			}, 1, nil
		},
	}

	uc := exportuc.NewExportTenantsUseCase(tenantRepo)
	var buf bytes.Buffer
	err := uc.Execute(ctx, landlordID, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 rows (header + 1 data), got %d", len(records))
	}
	if records[0][0] != "ID" {
		t.Errorf("expected header 'ID', got %s", records[0][0])
	}
	if records[1][1] != "Juan dela Cruz" {
		t.Errorf("expected 'Juan dela Cruz', got %s", records[1][1])
	}
	if records[1][2] != "juan@example.com" {
		t.Errorf("expected email 'juan@example.com', got %s", records[1][2])
	}
}

func TestExportProperties(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()

	propertyRepo := &mocks.PropertyRepositoryMock{
		ListFn: func(ctx context.Context, f repository.PropertyFilter) ([]*entity.Property, int, error) {
			return []*entity.Property{
				{
					ID:                 uuid.New(),
					Name:               "Lot 1",
					PropertyCode:       "LOT-001",
					PropertyType:       "LAND",
					SizeInSqm:          500.00,
					OwnerID:            ownerID,
					IsActive:           true,
					IsAvailableForRent: true,
					CreatedAt:          time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			}, 1, nil
		},
	}

	uc := exportuc.NewExportPropertiesUseCase(propertyRepo)
	var buf bytes.Buffer
	err := uc.Execute(ctx, ownerID, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}
	if records[1][1] != "Lot 1" {
		t.Errorf("expected 'Lot 1', got %s", records[1][1])
	}
	if records[1][4] != "500.00" {
		t.Errorf("expected '500.00', got %s", records[1][4])
	}
}

func TestExportDebts(t *testing.T) {
	ctx := context.Background()
	landlordID := uuid.New()

	origAmt, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)
	paidAmt, _ := entity.NewMoney(decimal.NewFromInt(2000), entity.CurrencyPHP)

	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, f repository.DebtFilter) ([]*entity.Debt, int, error) {
			return []*entity.Debt{
				{
					ID:             uuid.New(),
					TenantID:       uuid.New(),
					LandlordID:     landlordID,
					DebtType:       entity.DebtTypeRent,
					Description:    "March 2026 rent",
					OriginalAmount: origAmt,
					AmountPaid:     paidAmt,
					Status:         entity.DebtStatusPartial,
					DueDate:        time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
					CreatedAt:      time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				},
			}, 1, nil
		},
	}

	uc := exportuc.NewExportDebtsUseCase(debtRepo)
	var buf bytes.Buffer
	err := uc.Execute(ctx, landlordID, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reader := csv.NewReader(&buf)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}
	if records[1][4] != "5000.00" {
		t.Errorf("expected original amount '5000.00', got %s", records[1][4])
	}
	if records[1][5] != "2000.00" {
		t.Errorf("expected amount paid '2000.00', got %s", records[1][5])
	}
	if records[1][6] != "3000.00" {
		t.Errorf("expected balance '3000.00', got %s", records[1][6])
	}
}
```

- [ ] **Step 8: Run tests**

Run: `cd backend && go test ./tests/unit/ -run TestExport -v`
Expected: All 3 export tests PASS

- [ ] **Step 9: Run full test suite and build**

Run: `cd backend && go test ./tests/unit/... -count=1 && go build ./cmd/server/`
Expected: All tests pass, clean build

- [ ] **Step 10: Commit**

```bash
git add backend/internal/usecase/export/ \
  backend/internal/delivery/http/handler/export_handler.go \
  backend/internal/delivery/http/router/router.go \
  backend/cmd/server/main.go \
  backend/tests/unit/export_test.go
git commit -m "feat: add CSV data export for tenants, properties, and debts"
```

---

### Task 5: Overdue Debt Scheduler

**Files:**
- Create: `backend/internal/infrastructure/scheduler/overdue_scheduler.go`
- Modify: `backend/cmd/server/main.go`
- Create: `backend/tests/unit/overdue_scheduler_test.go`

- [ ] **Step 1: Create OverdueScheduler**

Create `backend/internal/infrastructure/scheduler/overdue_scheduler.go`:

```go
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

// OverdueScheduler periodically checks for and marks overdue debts.
type OverdueScheduler struct {
	debtRepo repository.DebtRepository
	interval time.Duration
}

// NewOverdueScheduler creates a new scheduler that runs at the given interval.
func NewOverdueScheduler(debtRepo repository.DebtRepository, interval time.Duration) *OverdueScheduler {
	return &OverdueScheduler{debtRepo: debtRepo, interval: interval}
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

func (s *OverdueScheduler) checkOverdue(ctx context.Context) {
	now := time.Now().UTC()
	updated := 0

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
```

- [ ] **Step 2: Write overdue scheduler test**

Create `backend/tests/unit/overdue_scheduler_test.go`:

```go
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/shopspring/decimal"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/scheduler"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func TestOverdueScheduler_MarksOverdueDebts(t *testing.T) {
	ctx := context.Background()
	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	tomorrow := time.Now().UTC().Add(24 * time.Hour)

	amt5k, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)
	zeroAmt := entity.ZeroMoney(entity.CurrencyPHP)

	overdueDebt := &entity.Debt{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		LandlordID:     uuid.New(),
		DebtType:       entity.DebtTypeRent,
		Description:    "Past due rent",
		OriginalAmount: amt5k,
		AmountPaid:     zeroAmt,
		Status:         entity.DebtStatusPending,
		DueDate:        yesterday,
		CreatedAt:      yesterday,
		UpdatedAt:      yesterday,
	}

	notYetDue := &entity.Debt{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		LandlordID:     uuid.New(),
		DebtType:       entity.DebtTypeRent,
		Description:    "Future rent",
		OriginalAmount: amt5k,
		AmountPaid:     zeroAmt,
		Status:         entity.DebtStatusPending,
		DueDate:        tomorrow,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	var updatedIDs []uuid.UUID
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, f repository.DebtFilter) ([]*entity.Debt, int, error) {
			if f.Status != nil && *f.Status == entity.DebtStatusPending {
				return []*entity.Debt{overdueDebt, notYetDue}, 2, nil
			}
			return nil, 0, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			updatedIDs = append(updatedIDs, d.ID)
			return nil
		},
	}

	s := scheduler.NewOverdueScheduler(debtRepo, time.Hour)
	s.CheckOverdue(ctx)

	if len(updatedIDs) != 1 {
		t.Fatalf("expected 1 debt updated, got %d", len(updatedIDs))
	}
	if updatedIDs[0] != overdueDebt.ID {
		t.Errorf("expected debt %s to be updated, got %s", overdueDebt.ID, updatedIDs[0])
	}
	if overdueDebt.Status != entity.DebtStatusOverdue {
		t.Errorf("expected status OVERDUE, got %s", overdueDebt.Status)
	}
}

func TestOverdueScheduler_SkipsAlreadyOverdue(t *testing.T) {
	ctx := context.Background()
	yesterday := time.Now().UTC().Add(-24 * time.Hour)

	amt5k, _ := entity.NewMoney(decimal.NewFromInt(5000), entity.CurrencyPHP)
	zeroAmt := entity.ZeroMoney(entity.CurrencyPHP)

	alreadyOverdue := &entity.Debt{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		LandlordID:     uuid.New(),
		DebtType:       entity.DebtTypeRent,
		Description:    "Already overdue",
		OriginalAmount: amt5k,
		AmountPaid:     zeroAmt,
		Status:         entity.DebtStatusOverdue,
		DueDate:        yesterday,
		CreatedAt:      yesterday,
		UpdatedAt:      yesterday,
	}

	updateCalled := false
	debtRepo := &mocks.DebtRepositoryMock{
		ListFn: func(ctx context.Context, f repository.DebtFilter) ([]*entity.Debt, int, error) {
			// Overdue debts shouldn't be in PENDING/PARTIAL lists
			return []*entity.Debt{alreadyOverdue}, 1, nil
		},
		UpdateFn: func(ctx context.Context, d *entity.Debt) error {
			updateCalled = true
			return nil
		},
	}

	s := scheduler.NewOverdueScheduler(debtRepo, time.Hour)
	s.CheckOverdue(ctx)

	if updateCalled {
		t.Error("should not update already-overdue debt")
	}
}
```

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./tests/unit/ -run TestOverdueScheduler -v`
Expected: Both tests PASS

- [ ] **Step 4: Wire scheduler in main.go**

In `backend/cmd/server/main.go`, add import:
```go
"github.com/ISubamariner/guimba-go/backend/internal/infrastructure/scheduler"
```

After the export wiring and before `// Set up router`, add:
```go
// Start overdue debt scheduler (checks every 24 hours)
overdueScheduler := scheduler.NewOverdueScheduler(debtRepo, 24*time.Hour)
go overdueScheduler.Start(ctx)
```

- [ ] **Step 5: Run full test suite and build**

Run: `cd backend && go test ./tests/unit/... -count=1 && go build ./cmd/server/`
Expected: All tests pass, clean build

- [ ] **Step 6: Commit**

```bash
git add backend/internal/infrastructure/scheduler/overdue_scheduler.go \
  backend/cmd/server/main.go \
  backend/tests/unit/overdue_scheduler_test.go
git commit -m "feat: add overdue debt scheduler with 24-hour interval"
```

---

## Deferred Work

The following are **not included** in this plan and should be addressed in future phases:

- **Forgot Password / Reset Password** - Requires email service integration (SMTP/SendGrid) not yet configured
- **OCR Receipt Scanning** - Requires external AI API (Gemini) and file upload handling
- **Background Email Notifications** - Requires email service; the overdue scheduler above provides the detection mechanism
- **Audit Log Export** - Can be added to Task 4 later (reads from MongoDB via audit repo)
- **PDF Export** - Can be added alongside CSV export when PDF library is selected

## Verification Checklist

After completing all tasks:
- [ ] `cd backend && go test ./tests/unit/... -count=1` — all tests pass
- [ ] `cd backend && go build ./cmd/server/` — clean build
- [ ] `cd backend && go vet ./...` — no issues
- [ ] All new endpoints appear in Swagger after `swag init`
- [ ] CLAUDE.md routes table updated with new endpoints
