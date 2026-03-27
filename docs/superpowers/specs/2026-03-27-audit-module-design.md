# Audit System Module — Design Spec

**Goal:** Implement a cross-cutting audit logging system that records all mutations across every domain module, stores entries in MongoDB, buffers through Redis for durability, and exposes query endpoints for admins and landlords.

**Source of truth:** `documentation/prompts/business-logic-reference.md` Sections 10 (Audit System) and 8.2 (Recent Activities).

---

## 1. Architecture Overview

The Audit System uses **direct injection** — each mutating use case receives an `AuditRepository` dependency and explicitly calls `Log()` after successful mutations. A thin HTTP middleware extracts request metadata (IP, user agent, endpoint, method) into context so use cases can include it without coupling to HTTP.

**Data flow:**

```
Use Case → AuditRepository.Log(entry)
               ↓
         BufferedAuditLogger (serializes to JSON, pushes to Redis list)
               ↓
         Background goroutine (pops from Redis, writes to MongoDB)
               ↓
         MongoDB "audit_logs" collection (append-only, immutable)
```

**Failure handling:** If MongoDB is unavailable, entries remain in the Redis queue and are retried with exponential backoff (1s, 2s, 4s, max 30s). Mutations are never blocked by audit failures. On graceful shutdown, the background worker drains the queue before exiting.

---

## 2. Domain Layer

### 2.1 AuditEntry Struct

```go
// domain/repository/audit_repository.go

type AuditEntry struct {
    ID           uuid.UUID
    UserID       uuid.UUID
    UserEmail    string
    UserRole     string
    Action       string           // e.g. "CREATE_TENANT", "APPLY_PAYMENT"
    ResourceType string           // e.g. "Tenant", "Debt", "Transaction"
    ResourceID   uuid.UUID
    IPAddress    string
    UserAgent    string
    Endpoint     string
    Method       string
    StatusCode   int
    Success      bool
    ErrorMessage *string
    Metadata     map[string]any   // rich context per action type
    Timestamp    time.Time
}
```

### 2.2 AuditRepository Interface

```go
type AuditRepository interface {
    Log(ctx context.Context, entry *AuditEntry) error
    List(ctx context.Context, filter AuditFilter) ([]*AuditEntry, int, error)
}
```

### 2.3 AuditFilter

```go
type AuditFilter struct {
    UserID       *uuid.UUID
    LandlordID   *uuid.UUID       // for landlord-scoped queries (user_id OR metadata.landlord_id)
    Action       *string
    ResourceType *string
    Success      *bool
    FromDate     *time.Time
    ToDate       *time.Time
    Limit        int
    Offset       int
}
```

When `LandlordID` is set, the query returns entries where `user_id = landlordID` OR `metadata.landlord_id = landlordID`. This supports the landlord-scoped endpoint.

---

## 3. Request Metadata Middleware

```go
// delivery/http/middleware/audit_context.go

type auditContextKey string

const (
    AuditIPKey        auditContextKey = "audit_ip"
    AuditUserAgentKey auditContextKey = "audit_user_agent"
    AuditEndpointKey  auditContextKey = "audit_endpoint"
    AuditMethodKey    auditContextKey = "audit_method"
)
```

A `AuditContext` middleware added to the global middleware stack extracts `RemoteAddr`, `User-Agent`, `URL.Path`, and `Method` into context values.

A helper function `audit.FromContext(ctx)` returns a struct with all four fields, using safe empty-string defaults when keys are missing (so unit tests work without HTTP context). This helper also pulls `UserID`, `UserEmail`, and `UserRole` from the existing auth middleware context keys (`middleware.AuthUserIDKey`, `middleware.AuthEmailKey`, `middleware.AuthRolesKey`).

---

## 4. Infrastructure: BufferedAuditLogger

### 4.1 MongoAuditRepo

Implements the `Log` and `List` methods directly against MongoDB.

- **Collection:** `audit_logs` in the configured MongoDB database
- **Log:** Inserts a single BSON document
- **List:** Queries with filters, pagination, sorted by `timestamp DESC`
- **Indexes:**
  - `user_id` (for admin queries by user)
  - `resource_type` (for filtering by entity type)
  - `action` (for filtering by action)
  - `timestamp` (for date range queries and sorting)
  - Compound `(user_id, timestamp)` (for landlord-scoped queries)
  - `metadata.landlord_id` (for landlord portfolio queries)

### 4.2 BufferedAuditLogger

Wraps `MongoAuditRepo` with Redis-backed durability.

- **Log():** Serializes `AuditEntry` to JSON, pushes to Redis list key `audit:queue` via `LPUSH`
- **Background goroutine:** Pops entries via `BRPOP` (blocking pop, 1s timeout), writes to MongoDB
- **Batch size:** Up to 10 entries per flush cycle
- **On MongoDB failure:** Entry is re-pushed to Redis, retry with exponential backoff (1s → 2s → 4s → ... → max 30s)
- **Graceful shutdown:** Accepts a `context.Context`; on cancellation, switches to non-blocking `RPOP` loop to drain remaining entries before returning
- **List():** Delegates directly to `MongoAuditRepo.List()` (reads bypass the buffer)

### 4.3 Wiring in main.go

```go
mongoAuditRepo := mongo.NewAuditRepoMongo(mongoClient, cfg.Mongo.Database)
auditLogger := audit.NewBufferedAuditLogger(mongoAuditRepo, redisClient)
go auditLogger.Start(ctx) // background worker
defer auditLogger.Stop()  // drain on shutdown
```

The `auditLogger` (which implements `AuditRepository`) is injected into all mutating use cases.

---

## 5. Use Case Integration

### 5.1 Pattern

Each mutating use case gains an `auditRepo repository.AuditRepository` constructor parameter. After a successful mutation, the use case calls:

```go
uc.auditRepo.Log(ctx, &repository.AuditEntry{
    Action:       "CREATE_TENANT",
    ResourceType: "Tenant",
    ResourceID:   tenant.ID,
    Success:      true,
    Metadata:     map[string]any{
        "tenant_name": tenant.FullName,
        "landlord_id": tenant.LandlordID,
    },
})
```

The `BufferedAuditLogger` enriches the entry with `ID`, `Timestamp`, and context-derived fields (`UserID`, `UserEmail`, `UserRole`, `IPAddress`, `UserAgent`, `Endpoint`, `Method`) before pushing to Redis.

### 5.2 Affected Use Cases (~20)

| Module | Use Cases | Actions |
|:---|:---|:---|
| Tenant | Create, Update, Deactivate, Delete | `CREATE_TENANT`, `UPDATE_TENANT`, `DEACTIVATE_TENANT`, `DELETE_TENANT` |
| Property | Create, Update, Deactivate, Delete | `CREATE_PROPERTY`, `UPDATE_PROPERTY`, `DEACTIVATE_PROPERTY`, `DELETE_PROPERTY` |
| Debt | Create, Update, Cancel, MarkPaid, Delete | `CREATE_DEBT`, `UPDATE_DEBT`, `CANCEL_DEBT`, `MARK_DEBT_PAID`, `DELETE_DEBT` |
| Transaction | RecordPayment, RecordRefund, Verify | `APPLY_PAYMENT`, `APPLY_REFUND`, `VERIFY_TRANSACTION` |
| Beneficiary | Create, Update, Delete, Enroll, Remove | `CREATE_BENEFICIARY`, `UPDATE_BENEFICIARY`, `DELETE_BENEFICIARY`, `ENROLL_BENEFICIARY`, `REMOVE_BENEFICIARY` |
| User | Update, Delete, AssignRole, RemoveRole | `UPDATE_USER`, `DELETE_USER`, `ASSIGN_ROLE`, `REMOVE_ROLE` |
| Auth | Register | `REGISTER_USER` |

### 5.3 Metadata per Action

Each action includes relevant domain context in the `Metadata` map. Examples:

- **CREATE_DEBT:** `landlord_id`, `tenant_id`, `tenant_name`, `amount`, `currency`, `debt_type`, `description`, `property_id`, `property_name`
- **APPLY_PAYMENT:** `landlord_id`, `tenant_id`, `payment_amount`, `currency`, `balance_before`, `balance_after`, `debt_type`
- **UPDATE_TENANT:** `landlord_id`, `changes` (map of field → before/after)
- **ASSIGN_ROLE:** `role_name`, `target_user_email`

Simpler actions (Delete, Deactivate) include just the resource name/ID and landlord_id.

### 5.4 What Doesn't Change

Read-only use cases (Get, List) are not audited per the spec. Login could optionally be audited in a future iteration but is out of scope here.

---

## 6. Delivery Layer

### 6.1 Query Use Cases

**ListAuditLogsUseCase** — for admin/auditor access:
- Passes `AuditFilter` directly to `AuditRepository.List()`
- Default limit 20, max 100

**ListLandlordAuditLogsUseCase** — for landlord-scoped access:
- Sets `filter.LandlordID` to the authenticated user's ID
- Passes to `AuditRepository.List()`
- Default limit 20, max 100

### 6.2 DTOs

```go
// dto/audit_dto.go

type AuditEntryResponse struct {
    ID           string         `json:"id"`
    UserID       string         `json:"user_id"`
    UserEmail    string         `json:"user_email"`
    UserRole     string         `json:"user_role"`
    Action       string         `json:"action"`
    ResourceType string         `json:"resource_type"`
    ResourceID   string         `json:"resource_id"`
    IPAddress    string         `json:"ip_address"`
    Endpoint     string         `json:"endpoint"`
    Method       string         `json:"method"`
    StatusCode   int            `json:"status_code"`
    Success      bool           `json:"success"`
    ErrorMessage *string        `json:"error_message,omitempty"`
    Metadata     map[string]any `json:"metadata,omitempty"`
    Timestamp    string         `json:"timestamp"`
}

type AuditListResponse struct {
    Data   []AuditEntryResponse `json:"data"`
    Total  int                  `json:"total"`
    Limit  int                  `json:"limit"`
    Offset int                  `json:"offset"`
}
```

### 6.3 Handler

`AuditHandler` with two methods:
- `List(w, r)` — parses query filters, calls `ListAuditLogsUseCase`
- `LandlordList(w, r)` — parses query filters, injects landlord ID from auth context, calls `ListLandlordAuditLogsUseCase`

### 6.4 Routes

```go
r.Route("/audit", func(r chi.Router) {
    r.Use(requireAuth)
    r.With(middleware.RequireRole("admin", "auditor")).Get("/", h.Audit.List)
    r.With(middleware.RequireRole("admin", "landlord")).Get("/landlord", h.Audit.LandlordList)
})
```

---

## 7. Testing

### 7.1 New Tests (~25-30)

- **AuditEntry** validation tests (required fields)
- **BufferedAuditLogger** — mock Redis + mock MongoAuditRepo, verify entries flow
- **ListAuditLogsUseCase** — limit clamping, filter passthrough
- **ListLandlordAuditLogsUseCase** — landlord scoping injection
- **AuditHandler** — List success, LandlordList success, invalid filters
- **AuditContext middleware** — context values set correctly
- **FromContext helper** — safe defaults when keys missing

### 7.2 Modified Existing Tests (~20)

All existing mutating use case tests gain a mock `AuditRepository` in constructor calls. The mock's `LogFn` defaults to nil (no-op). A few key tests verify audit is called on success:
- `TestCreateTenant_AuditsOnSuccess`
- `TestRecordPayment_AuditsOnSuccess`
- `TestCancelDebt_AuditsOnSuccess`

### 7.3 Mock

```go
// tests/mocks/audit_repository_mock.go

type AuditRepositoryMock struct {
    LogFn  func(ctx context.Context, entry *repository.AuditEntry) error
    ListFn func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error)
}
```

### 7.4 Out of Scope

- MongoDB integration tests (deferred to integration test phase)
- Redis queue durability tests (integration-level concern)

---

## 8. File Map

### New Files (~12)

| # | Path | Responsibility |
|---|------|----------------|
| 1 | `backend/internal/domain/repository/audit_repository.go` | AuditEntry, AuditFilter, AuditRepository interface |
| 2 | `backend/internal/delivery/http/middleware/audit_context.go` | Request metadata extraction middleware |
| 3 | `backend/pkg/audit/context.go` | `FromContext` helper to pull audit fields from context |
| 4 | `backend/internal/infrastructure/persistence/mongo/audit_repo_mongo.go` | MongoDB AuditRepository implementation |
| 5 | `backend/internal/infrastructure/audit/buffered_logger.go` | Redis-buffered AuditLogger wrapping MongoAuditRepo |
| 6 | `backend/internal/usecase/audit/list_audit_logs.go` | ListAuditLogsUseCase (admin/auditor) |
| 7 | `backend/internal/usecase/audit/list_landlord_audit_logs.go` | ListLandlordAuditLogsUseCase (landlord-scoped) |
| 8 | `backend/internal/delivery/http/dto/audit_dto.go` | AuditEntryResponse, AuditListResponse, NewAuditEntryResponse |
| 9 | `backend/internal/delivery/http/handler/audit_handler.go` | AuditHandler with List and LandlordList |
| 10 | `backend/tests/mocks/audit_repository_mock.go` | Manual mock for AuditRepository |
| 11 | `backend/tests/unit/audit_usecase_test.go` | Use case tests |
| 12 | `backend/tests/unit/audit_handler_test.go` | Handler tests |

### Modified Files (~25)

| Path | Change |
|------|--------|
| `backend/internal/delivery/http/router/router.go` | Add `Audit` field to Handlers, add `/audit` routes |
| `backend/cmd/server/main.go` | Wire MongoAuditRepo, BufferedAuditLogger, audit use cases, handler; inject auditRepo into all mutating use cases |
| ~20 use case files | Add `auditRepo` constructor param + `Log()` call after mutations |
| ~10 existing test files | Add `AuditRepositoryMock{}` to constructor calls |
| `CLAUDE.md` | Update routes table and Current Modules |
