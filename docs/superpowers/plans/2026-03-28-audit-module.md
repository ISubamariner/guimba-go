# Audit System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a cross-cutting audit logging system — AuditEntry domain type, MongoDB persistence, Redis-buffered durability, audit query endpoints, and audit injection into all ~22 mutating use cases.

**Architecture:** Direct injection of `AuditRepository` into every mutating use case. A `BufferedAuditLogger` pushes entries to a Redis queue; a background goroutine drains the queue to MongoDB. A thin HTTP middleware captures request metadata (IP, user agent, endpoint, method) into context. Two query endpoints expose audit logs to admins and landlords.

**Tech Stack:** Go 1.26+, Chi v5, mongo-go-driver v2, go-redis v9, google/uuid

**Spec:** `docs/superpowers/specs/2026-03-27-audit-module-design.md`

---

## File Structure

### New Files (13)

| # | Path | Responsibility |
|---|------|----------------|
| 1 | `backend/internal/domain/repository/audit_repository.go` | AuditEntry struct, AuditFilter struct, AuditRepository interface |
| 2 | `backend/pkg/audit/context.go` | `FromContext(ctx)` helper — extracts auth + request metadata from context |
| 3 | `backend/internal/delivery/http/middleware/audit_context.go` | HTTP middleware that sets IP, user agent, endpoint, method in context |
| 4 | `backend/tests/mocks/audit_repository_mock.go` | Manual mock for AuditRepository (function-field pattern) |
| 5 | `backend/internal/infrastructure/persistence/mongo/audit_repo_mongo.go` | MongoDB AuditRepository implementation (Log + List) |
| 6 | `backend/internal/infrastructure/audit/buffered_logger.go` | Redis-buffered AuditLogger wrapping MongoAuditRepo |
| 7 | `backend/internal/usecase/audit/list_audit_logs.go` | ListAuditLogsUseCase (admin/auditor) |
| 8 | `backend/internal/usecase/audit/list_landlord_audit_logs.go` | ListLandlordAuditLogsUseCase (landlord-scoped) |
| 9 | `backend/internal/delivery/http/dto/audit_dto.go` | AuditEntryResponse, AuditListResponse DTOs |
| 10 | `backend/internal/delivery/http/handler/audit_handler.go` | AuditHandler with List and LandlordList endpoints |
| 11 | `backend/tests/unit/audit_context_test.go` | Tests for AuditContext middleware + FromContext helper |
| 12 | `backend/tests/unit/audit_usecase_test.go` | Tests for audit query use cases |
| 13 | `backend/tests/unit/audit_handler_test.go` | Tests for AuditHandler |

### Modified Files (~30)

| Path | Change |
|------|--------|
| `backend/internal/delivery/http/router/router.go` | Add `Audit` field to Handlers, add `/audit` route group |
| `backend/cmd/server/main.go` | Wire MongoAuditRepo, BufferedAuditLogger, audit use cases, handler; inject auditRepo into all mutating use cases; add AuditContext middleware |
| 22 use case files (see Tasks 7-9) | Add `auditRepo` constructor param + `Log()` call |
| ~8 test files | Add `&mocks.AuditRepositoryMock{}` to constructor calls |
| `CLAUDE.md` | Update routes table and Current Modules |

---

## Task 1: Domain Layer + Mock

**Files:**
- Create: `backend/internal/domain/repository/audit_repository.go`
- Create: `backend/tests/mocks/audit_repository_mock.go`

- [ ] **Step 1: Create AuditRepository interface**

Create `backend/internal/domain/repository/audit_repository.go`:

```go
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditEntry represents an immutable audit log record.
type AuditEntry struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	UserEmail    string
	UserRole     string
	Action       string
	ResourceType string
	ResourceID   uuid.UUID
	IPAddress    string
	UserAgent    string
	Endpoint     string
	Method       string
	StatusCode   int
	Success      bool
	ErrorMessage *string
	Metadata     map[string]any
	Timestamp    time.Time
}

// AuditFilter specifies criteria for querying audit logs.
type AuditFilter struct {
	UserID       *uuid.UUID
	LandlordID   *uuid.UUID // user_id OR metadata.landlord_id match
	Action       *string
	ResourceType *string
	Success      *bool
	FromDate     *time.Time
	ToDate       *time.Time
	Limit        int
	Offset       int
}

// AuditRepository defines the interface for audit log persistence.
type AuditRepository interface {
	Log(ctx context.Context, entry *AuditEntry) error
	List(ctx context.Context, filter AuditFilter) ([]*AuditEntry, int, error)
}
```

- [ ] **Step 2: Create AuditRepository mock**

Create `backend/tests/mocks/audit_repository_mock.go`:

```go
package mocks

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

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
```

- [ ] **Step 3: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: Clean compilation.

- [ ] **Step 4: Commit**

```bash
git add internal/domain/repository/audit_repository.go tests/mocks/audit_repository_mock.go
git commit -m "feat: add AuditEntry, AuditRepository interface, and mock"
```

---

## Task 2: Audit Context Middleware + Helper

**Files:**
- Create: `backend/internal/delivery/http/middleware/audit_context.go`
- Create: `backend/pkg/audit/context.go`
- Create: `backend/tests/unit/audit_context_test.go`

- [ ] **Step 1: Create AuditContext middleware**

Create `backend/internal/delivery/http/middleware/audit_context.go`:

```go
package middleware

import (
	"context"
	"net/http"
)

const (
	// AuditIPKey is the context key for the client's IP address.
	AuditIPKey contextKey = "audit_ip"
	// AuditUserAgentKey is the context key for the client's user agent.
	AuditUserAgentKey contextKey = "audit_user_agent"
	// AuditEndpointKey is the context key for the request endpoint path.
	AuditEndpointKey contextKey = "audit_endpoint"
	// AuditMethodKey is the context key for the HTTP method.
	AuditMethodKey contextKey = "audit_method"
)

// AuditContext extracts request metadata into context for audit logging.
func AuditContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, AuditIPKey, r.RemoteAddr)
		ctx = context.WithValue(ctx, AuditUserAgentKey, r.Header.Get("User-Agent"))
		ctx = context.WithValue(ctx, AuditEndpointKey, r.URL.Path)
		ctx = context.WithValue(ctx, AuditMethodKey, r.Method)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

- [ ] **Step 2: Create FromContext helper**

Create `backend/pkg/audit/context.go`:

```go
package audit

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// contextKey type matches middleware.contextKey for value extraction.
type contextKey string

const (
	auditIPKey        contextKey = "audit_ip"
	auditUserAgentKey contextKey = "audit_user_agent"
	auditEndpointKey  contextKey = "audit_endpoint"
	auditMethodKey    contextKey = "audit_method"
	authUserIDKey     contextKey = "auth_user_id"
	authEmailKey      contextKey = "auth_email"
	authRolesKey      contextKey = "auth_roles"
)

// RequestInfo holds audit-relevant metadata extracted from context.
type RequestInfo struct {
	UserID    uuid.UUID
	UserEmail string
	UserRole  string
	IPAddress string
	UserAgent string
	Endpoint  string
	Method    string
}

// FromContext extracts audit-relevant fields from context.
// Returns safe defaults when keys are missing (e.g., in unit tests).
func FromContext(ctx context.Context) RequestInfo {
	info := RequestInfo{}

	if v, ok := ctx.Value(authUserIDKey).(string); ok {
		if parsed, err := uuid.Parse(v); err == nil {
			info.UserID = parsed
		}
	}
	if v, ok := ctx.Value(authEmailKey).(string); ok {
		info.UserEmail = v
	}
	if v, ok := ctx.Value(authRolesKey).([]string); ok {
		info.UserRole = strings.Join(v, ",")
	}
	if v, ok := ctx.Value(auditIPKey).(string); ok {
		info.IPAddress = v
	}
	if v, ok := ctx.Value(auditUserAgentKey).(string); ok {
		info.UserAgent = v
	}
	if v, ok := ctx.Value(auditEndpointKey).(string); ok {
		info.Endpoint = v
	}
	if v, ok := ctx.Value(auditMethodKey).(string); ok {
		info.Method = v
	}

	return info
}
```

- [ ] **Step 3: Write tests**

Create `backend/tests/unit/audit_context_test.go`:

```go
package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/pkg/audit"
)

func TestAuditContext_SetsValues(t *testing.T) {
	var capturedCtx context.Context
	handler := middleware.AuditContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tenants", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	info := audit.FromContext(capturedCtx)
	if info.IPAddress == "" {
		t.Error("expected IPAddress to be set")
	}
	if info.UserAgent != "TestAgent/1.0" {
		t.Errorf("expected UserAgent 'TestAgent/1.0', got '%s'", info.UserAgent)
	}
	if info.Endpoint != "/api/v1/tenants" {
		t.Errorf("expected Endpoint '/api/v1/tenants', got '%s'", info.Endpoint)
	}
	if info.Method != "POST" {
		t.Errorf("expected Method 'POST', got '%s'", info.Method)
	}
}

func TestFromContext_EmptyContext(t *testing.T) {
	info := audit.FromContext(context.Background())
	if info.IPAddress != "" {
		t.Errorf("expected empty IPAddress, got '%s'", info.IPAddress)
	}
	if info.UserEmail != "" {
		t.Errorf("expected empty UserEmail, got '%s'", info.UserEmail)
	}
	if info.Method != "" {
		t.Errorf("expected empty Method, got '%s'", info.Method)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./tests/unit/ -run "TestAuditContext|TestFromContext" -v
```

Expected: 2 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/delivery/http/middleware/audit_context.go pkg/audit/context.go tests/unit/audit_context_test.go
git commit -m "feat: add audit context middleware and FromContext helper"
```

---

## Task 3: MongoDB Audit Repository

**Files:**
- Create: `backend/internal/infrastructure/persistence/mongo/audit_repo_mongo.go`

**Context needed:**
- MongoDB client is `go.mongodb.org/mongo-driver/v2/mongo`
- Database name comes from config: `cfg.Mongo.DB` (default `guimba_db`)
- AuditEntry and AuditFilter from `domain/repository/audit_repository.go`

- [ ] **Step 1: Implement AuditRepoMongo**

Create `backend/internal/infrastructure/persistence/mongo/audit_repo_mongo.go`:

```go
package mongo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type AuditRepoMongo struct {
	collection *mongo.Collection
}

func NewAuditRepoMongo(client *mongo.Client, dbName string) *AuditRepoMongo {
	return &AuditRepoMongo{
		collection: client.Database(dbName).Collection("audit_logs"),
	}
}

type auditDocument struct {
	ID           string         `bson:"_id"`
	UserID       string         `bson:"user_id"`
	UserEmail    string         `bson:"user_email"`
	UserRole     string         `bson:"user_role"`
	Action       string         `bson:"action"`
	ResourceType string         `bson:"resource_type"`
	ResourceID   string         `bson:"resource_id"`
	IPAddress    string         `bson:"ip_address"`
	UserAgent    string         `bson:"user_agent"`
	Endpoint     string         `bson:"endpoint"`
	Method       string         `bson:"method"`
	StatusCode   int            `bson:"status_code"`
	Success      bool           `bson:"success"`
	ErrorMessage *string        `bson:"error_message,omitempty"`
	Metadata     map[string]any `bson:"metadata,omitempty"`
	Timestamp    time.Time      `bson:"timestamp"`
}

func toDocument(e *repository.AuditEntry) auditDocument {
	return auditDocument{
		ID:           e.ID.String(),
		UserID:       e.UserID.String(),
		UserEmail:    e.UserEmail,
		UserRole:     e.UserRole,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID.String(),
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
		Endpoint:     e.Endpoint,
		Method:       e.Method,
		StatusCode:   e.StatusCode,
		Success:      e.Success,
		ErrorMessage: e.ErrorMessage,
		Metadata:     e.Metadata,
		Timestamp:    e.Timestamp,
	}
}

func fromDocument(d auditDocument) *repository.AuditEntry {
	id, _ := uuid.Parse(d.ID)
	userID, _ := uuid.Parse(d.UserID)
	resourceID, _ := uuid.Parse(d.ResourceID)
	return &repository.AuditEntry{
		ID:           id,
		UserID:       userID,
		UserEmail:    d.UserEmail,
		UserRole:     d.UserRole,
		Action:       d.Action,
		ResourceType: d.ResourceType,
		ResourceID:   resourceID,
		IPAddress:    d.IPAddress,
		UserAgent:    d.UserAgent,
		Endpoint:     d.Endpoint,
		Method:       d.Method,
		StatusCode:   d.StatusCode,
		Success:      d.Success,
		ErrorMessage: d.ErrorMessage,
		Metadata:     d.Metadata,
		Timestamp:    d.Timestamp,
	}
}

func (r *AuditRepoMongo) Log(ctx context.Context, entry *repository.AuditEntry) error {
	doc := toDocument(entry)
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *AuditRepoMongo) List(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	bsonFilter := bson.D{}

	if filter.UserID != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "user_id", Value: filter.UserID.String()})
	}
	if filter.LandlordID != nil {
		lid := filter.LandlordID.String()
		bsonFilter = append(bsonFilter, bson.E{
			Key: "$or",
			Value: bson.A{
				bson.D{{Key: "user_id", Value: lid}},
				bson.D{{Key: "metadata.landlord_id", Value: lid}},
			},
		})
	}
	if filter.Action != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "action", Value: *filter.Action})
	}
	if filter.ResourceType != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "resource_type", Value: *filter.ResourceType})
	}
	if filter.Success != nil {
		bsonFilter = append(bsonFilter, bson.E{Key: "success", Value: *filter.Success})
	}
	if filter.FromDate != nil || filter.ToDate != nil {
		dateFilter := bson.D{}
		if filter.FromDate != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$gte", Value: *filter.FromDate})
		}
		if filter.ToDate != nil {
			dateFilter = append(dateFilter, bson.E{Key: "$lte", Value: *filter.ToDate})
		}
		bsonFilter = append(bsonFilter, bson.E{Key: "timestamp", Value: dateFilter})
	}

	// Count total
	total, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination and sort
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(filter.Offset)).
		SetLimit(int64(filter.Limit))

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var entries []*repository.AuditEntry
	for cursor.Next(ctx) {
		var doc auditDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		entries = append(entries, fromDocument(doc))
	}

	return entries, int(total), cursor.Err()
}

// EnsureIndexes creates indexes for efficient audit log queries.
func (r *AuditRepoMongo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "resource_type", Value: 1}}},
		{Keys: bson.D{{Key: "action", Value: 1}}},
		{Keys: bson.D{{Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "metadata.landlord_id", Value: 1}}},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: Clean compilation.

- [ ] **Step 3: Commit**

```bash
git add internal/infrastructure/persistence/mongo/audit_repo_mongo.go
git commit -m "feat: add MongoDB AuditRepository implementation"
```

---

## Task 4: Buffered Audit Logger

**Files:**
- Create: `backend/internal/infrastructure/audit/buffered_logger.go`

**Context needed:**
- Redis client: `github.com/redis/go-redis/v9`
- AuditEntry from `domain/repository/audit_repository.go`
- `audit.FromContext(ctx)` from `pkg/audit/context.go`

- [ ] **Step 1: Implement BufferedAuditLogger**

Create `backend/internal/infrastructure/audit/buffered_logger.go`:

```go
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	pkgaudit "github.com/ISubamariner/guimba-go/backend/pkg/audit"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

const (
	auditQueueKey  = "audit:queue"
	maxBackoff     = 30 * time.Second
	batchSize      = 10
	popTimeout     = 1 * time.Second
)

// BufferedAuditLogger wraps a MongoAuditRepo with Redis-backed buffering.
// It implements repository.AuditRepository.
type BufferedAuditLogger struct {
	mongo  repository.AuditRepository
	rdb    *redis.Client
	cancel context.CancelFunc
	done   chan struct{}
}

// NewBufferedAuditLogger creates a new BufferedAuditLogger.
func NewBufferedAuditLogger(mongo repository.AuditRepository, rdb *redis.Client) *BufferedAuditLogger {
	return &BufferedAuditLogger{
		mongo: mongo,
		rdb:   rdb,
		done:  make(chan struct{}),
	}
}

// Log enriches the entry with context metadata and pushes to Redis queue.
// If Redis is unavailable, the error is logged and nil is returned (non-blocking).
func (b *BufferedAuditLogger) Log(ctx context.Context, entry *repository.AuditEntry) error {
	// Enrich with context metadata
	info := pkgaudit.FromContext(ctx)
	entry.ID = uuid.New()
	entry.UserID = info.UserID
	entry.UserEmail = info.UserEmail
	entry.UserRole = info.UserRole
	entry.IPAddress = info.IPAddress
	entry.UserAgent = info.UserAgent
	entry.Endpoint = info.Endpoint
	entry.Method = info.Method
	entry.StatusCode = 200 // audit only fires on success
	entry.Timestamp = time.Now().UTC()

	data, err := json.Marshal(entry)
	if err != nil {
		slog.Error("audit: failed to marshal entry", "error", err)
		return nil
	}

	if err := b.rdb.LPush(ctx, auditQueueKey, data).Err(); err != nil {
		slog.Error("audit: failed to push to Redis queue", "error", err, "action", entry.Action)
		return nil
	}

	return nil
}

// List delegates directly to the MongoDB implementation (reads bypass the buffer).
func (b *BufferedAuditLogger) List(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	return b.mongo.List(ctx, filter)
}

// Start begins the background worker that drains the Redis queue to MongoDB.
func (b *BufferedAuditLogger) Start(ctx context.Context) {
	workerCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	backoff := time.Second

	for {
		select {
		case <-workerCtx.Done():
			b.drain()
			close(b.done)
			return
		default:
		}

		processed := 0
		for i := 0; i < batchSize; i++ {
			result, err := b.rdb.BRPop(workerCtx, popTimeout, auditQueueKey).Result()
			if err != nil {
				if err == redis.Nil || err == context.Canceled || err == context.DeadlineExceeded {
					break
				}
				slog.Error("audit: failed to pop from Redis queue", "error", err)
				break
			}

			if len(result) < 2 {
				break
			}

			var entry repository.AuditEntry
			if err := json.Unmarshal([]byte(result[1]), &entry); err != nil {
				slog.Error("audit: failed to unmarshal entry", "error", err)
				continue
			}

			if err := b.mongo.Log(workerCtx, &entry); err != nil {
				slog.Error("audit: failed to write to MongoDB, re-queuing", "error", err, "action", entry.Action)
				data, _ := json.Marshal(entry)
				b.rdb.RPush(workerCtx, auditQueueKey, data)

				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				break
			}

			processed++
			backoff = time.Second // reset on success
		}

		if processed == 0 {
			// No items processed, brief sleep before next cycle
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Stop signals the background worker to stop and waits for it to drain.
func (b *BufferedAuditLogger) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
	<-b.done
}

// drain processes remaining entries in the Redis queue before shutdown.
func (b *BufferedAuditLogger) drain() {
	ctx := context.Background()
	for {
		result, err := b.rdb.RPop(ctx, auditQueueKey).Result()
		if err != nil {
			return // queue empty or Redis error
		}

		var entry repository.AuditEntry
		if err := json.Unmarshal([]byte(result), &entry); err != nil {
			slog.Error("audit: drain failed to unmarshal", "error", err)
			continue
		}

		if err := b.mongo.Log(ctx, &entry); err != nil {
			slog.Error("audit: drain failed to write to MongoDB", "error", err, "action", entry.Action)
			// On drain failure, re-push for next startup
			data, _ := json.Marshal(entry)
			b.rdb.LPush(ctx, auditQueueKey, data)
			return
		}
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./...
```

Expected: Clean compilation.

- [ ] **Step 3: Commit**

```bash
git add internal/infrastructure/audit/buffered_logger.go
git commit -m "feat: add Redis-buffered audit logger"
```

---

## Task 5: Audit Query Use Cases + Tests

**Files:**
- Create: `backend/internal/usecase/audit/list_audit_logs.go`
- Create: `backend/internal/usecase/audit/list_landlord_audit_logs.go`
- Create: `backend/tests/unit/audit_usecase_test.go`

- [ ] **Step 1: Implement ListAuditLogsUseCase**

Create `backend/internal/usecase/audit/list_audit_logs.go`:

```go
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
```

- [ ] **Step 2: Implement ListLandlordAuditLogsUseCase**

Create `backend/internal/usecase/audit/list_landlord_audit_logs.go`:

```go
package audit

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

type ListLandlordAuditLogsUseCase struct {
	repo repository.AuditRepository
}

func NewListLandlordAuditLogsUseCase(repo repository.AuditRepository) *ListLandlordAuditLogsUseCase {
	return &ListLandlordAuditLogsUseCase{repo: repo}
}

func (uc *ListLandlordAuditLogsUseCase) Execute(ctx context.Context, landlordID uuid.UUID, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
	filter.LandlordID = &landlordID

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
```

- [ ] **Step 3: Write tests**

Create `backend/tests/unit/audit_usecase_test.go`:

```go
package unit

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newTestAuditEntry() *repository.AuditEntry {
	return &repository.AuditEntry{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		UserEmail:    "admin@test.com",
		UserRole:     "admin",
		Action:       "CREATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   uuid.New(),
		Success:      true,
		StatusCode:   200,
		Timestamp:    time.Now().UTC(),
	}
}

func TestListAuditLogs_Success(t *testing.T) {
	entries := []*repository.AuditEntry{newTestAuditEntry(), newTestAuditEntry()}
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			return entries, 2, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	result, total, err := uc.Execute(context.Background(), repository.AuditFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
}

func TestListAuditLogs_DefaultLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.AuditFilter{Limit: 0})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}

func TestListAuditLogs_MaxLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), repository.AuditFilter{Limit: 500})
	if capturedFilter.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", capturedFilter.Limit)
	}
}

func TestListLandlordAuditLogs_ScopesFilter(t *testing.T) {
	landlordID := uuid.New()
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListLandlordAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), landlordID, repository.AuditFilter{})
	if capturedFilter.LandlordID == nil {
		t.Fatal("expected LandlordID to be set")
	}
	if *capturedFilter.LandlordID != landlordID {
		t.Errorf("expected LandlordID %s, got %s", landlordID, *capturedFilter.LandlordID)
	}
}

func TestListLandlordAuditLogs_DefaultLimit(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	uc := audituc.NewListLandlordAuditLogsUseCase(repo)
	_, _, _ = uc.Execute(context.Background(), uuid.New(), repository.AuditFilter{Limit: 0})
	if capturedFilter.Limit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedFilter.Limit)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./tests/unit/ -run "TestListAuditLogs|TestListLandlordAuditLogs" -v
```

Expected: 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/usecase/audit/ tests/unit/audit_usecase_test.go
git commit -m "feat: add audit query use cases with tests"
```

---

## Task 6: Audit DTOs + Handler + Tests

**Files:**
- Create: `backend/internal/delivery/http/dto/audit_dto.go`
- Create: `backend/internal/delivery/http/handler/audit_handler.go`
- Create: `backend/tests/unit/audit_handler_test.go`

- [ ] **Step 1: Create audit DTOs**

Create `backend/internal/delivery/http/dto/audit_dto.go`:

```go
package dto

import (
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
)

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

func NewAuditEntryResponse(e *repository.AuditEntry) AuditEntryResponse {
	return AuditEntryResponse{
		ID:           e.ID.String(),
		UserID:       e.UserID.String(),
		UserEmail:    e.UserEmail,
		UserRole:     e.UserRole,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID.String(),
		IPAddress:    e.IPAddress,
		Endpoint:     e.Endpoint,
		Method:       e.Method,
		StatusCode:   e.StatusCode,
		Success:      e.Success,
		ErrorMessage: e.ErrorMessage,
		Metadata:     e.Metadata,
		Timestamp:    e.Timestamp.Format("2006-01-02T15:04:05Z"),
	}
}

func NewAuditListResponse(entries []*repository.AuditEntry, total, limit, offset int) AuditListResponse {
	data := make([]AuditEntryResponse, 0, len(entries))
	for _, e := range entries {
		data = append(data, NewAuditEntryResponse(e))
	}
	return AuditListResponse{Data: data, Total: total, Limit: limit, Offset: offset}
}
```

- [ ] **Step 2: Create audit handler**

Create `backend/internal/delivery/http/handler/audit_handler.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type AuditHandler struct {
	listUC         *audituc.ListAuditLogsUseCase
	listLandlordUC *audituc.ListLandlordAuditLogsUseCase
}

func NewAuditHandler(listUC *audituc.ListAuditLogsUseCase, listLandlordUC *audituc.ListLandlordAuditLogsUseCase) *AuditHandler {
	return &AuditHandler{listUC: listUC, listLandlordUC: listLandlordUC}
}

// List godoc
// @Summary      List audit logs
// @Description  Returns audit log entries with optional filters (admin/auditor only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        user_id       query  string  false  "Filter by user ID"
// @Param        action        query  string  false  "Filter by action"
// @Param        resource_type query  string  false  "Filter by resource type"
// @Param        success       query  bool    false  "Filter by success"
// @Param        from_date     query  string  false  "Filter from date (RFC3339)"
// @Param        to_date       query  string  false  "Filter to date (RFC3339)"
// @Param        limit         query  int     false  "Limit (default 20, max 100)"
// @Param        offset        query  int     false  "Offset"
// @Success      200  {object}  dto.AuditListResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Router       /audit [get]
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseAuditFilter(r)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	entries, total, err := h.listUC.Execute(r.Context(), filter)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	resp := dto.NewAuditListResponse(entries, total, filter.Limit, filter.Offset)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// LandlordList godoc
// @Summary      List landlord-scoped audit logs
// @Description  Returns audit log entries scoped to the authenticated landlord
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        action        query  string  false  "Filter by action"
// @Param        resource_type query  string  false  "Filter by resource type"
// @Param        from_date     query  string  false  "Filter from date (RFC3339)"
// @Param        to_date       query  string  false  "Filter to date (RFC3339)"
// @Param        limit         query  int     false  "Limit (default 20, max 100)"
// @Param        offset        query  int     false  "Offset"
// @Success      200  {object}  dto.AuditListResponse
// @Failure      400  {object}  apperror.ErrorResponse
// @Router       /audit/landlord [get]
func (h *AuditHandler) LandlordList(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.AuthUserIDKey).(string)
	if !ok || userIDStr == "" {
		apperror.WriteError(w, apperror.NewUnauthorized("Missing user ID"))
		return
	}
	landlordID, err := uuid.Parse(userIDStr)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest("Invalid user ID"))
		return
	}

	filter, err := parseAuditFilter(r)
	if err != nil {
		apperror.WriteError(w, apperror.NewBadRequest(err.Error()))
		return
	}

	entries, total, err := h.listLandlordUC.Execute(r.Context(), landlordID, filter)
	if err != nil {
		apperror.WriteError(w, apperror.NewInternal(err))
		return
	}

	resp := dto.NewAuditListResponse(entries, total, filter.Limit, filter.Offset)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseAuditFilter(r *http.Request) (repository.AuditFilter, error) {
	filter := repository.AuditFilter{}

	if s := r.URL.Query().Get("user_id"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return filter, err
		}
		filter.UserID = &id
	}
	if s := r.URL.Query().Get("action"); s != "" {
		filter.Action = &s
	}
	if s := r.URL.Query().Get("resource_type"); s != "" {
		filter.ResourceType = &s
	}
	if s := r.URL.Query().Get("success"); s != "" {
		v := s == "true"
		filter.Success = &v
	}
	if s := r.URL.Query().Get("from_date"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.FromDate = &t
	}
	if s := r.URL.Query().Get("to_date"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return filter, err
		}
		filter.ToDate = &t
	}
	if s := r.URL.Query().Get("limit"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Limit = v
	}
	if s := r.URL.Query().Get("offset"); s != "" {
		v, _ := strconv.Atoi(s)
		filter.Offset = v
	}

	return filter, nil
}
```

- [ ] **Step 3: Write handler tests**

Create `backend/tests/unit/audit_handler_test.go`:

```go
package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/dto"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/handler"
	"github.com/ISubamariner/guimba-go/backend/internal/delivery/http/middleware"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
	"github.com/ISubamariner/guimba-go/backend/tests/mocks"
)

func newAuditHandler() (*handler.AuditHandler, *mocks.AuditRepositoryMock) {
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			return []*repository.AuditEntry{newTestAuditEntry()}, 1, nil
		},
	}
	listUC := audituc.NewListAuditLogsUseCase(repo)
	listLandlordUC := audituc.NewListLandlordAuditLogsUseCase(repo)
	return handler.NewAuditHandler(listUC, listLandlordUC), repo
}

func TestAuditHandler_List_Success(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?limit=10", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp dto.AuditListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 entry, got %d", len(resp.Data))
	}
}

func TestAuditHandler_List_InvalidFromDate(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?from_date=not-a-date", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestAuditHandler_LandlordList_Success(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/landlord", nil)
	userID := uuid.New().String()
	ctx := context.WithValue(req.Context(), middleware.AuthUserIDKey, userID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.LandlordList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuditHandler_LandlordList_MissingUserID(t *testing.T) {
	h, _ := newAuditHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/landlord", nil)
	rr := httptest.NewRecorder()

	h.LandlordList(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuditHandler_List_WithFilters(t *testing.T) {
	var capturedFilter repository.AuditFilter
	repo := &mocks.AuditRepositoryMock{
		ListFn: func(ctx context.Context, filter repository.AuditFilter) ([]*repository.AuditEntry, int, error) {
			capturedFilter = filter
			return nil, 0, nil
		},
	}
	listUC := audituc.NewListAuditLogsUseCase(repo)
	listLandlordUC := audituc.NewListLandlordAuditLogsUseCase(repo)
	h := handler.NewAuditHandler(listUC, listLandlordUC)

	now := time.Now().UTC().Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit?action=CREATE_TENANT&resource_type=Tenant&success=true&from_date="+now, nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedFilter.Action == nil || *capturedFilter.Action != "CREATE_TENANT" {
		t.Error("expected action filter to be set")
	}
	if capturedFilter.ResourceType == nil || *capturedFilter.ResourceType != "Tenant" {
		t.Error("expected resource_type filter to be set")
	}
	if capturedFilter.Success == nil || *capturedFilter.Success != true {
		t.Error("expected success filter to be set")
	}
	if capturedFilter.FromDate == nil {
		t.Error("expected from_date filter to be set")
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./tests/unit/ -run "TestAuditHandler" -v
```

Expected: 5 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/delivery/http/dto/audit_dto.go internal/delivery/http/handler/audit_handler.go tests/unit/audit_handler_test.go
git commit -m "feat: add audit DTOs, handler, and handler tests"
```

---

## Task 7: Inject Audit into Tenant + Property Use Cases

**Files:**
- Modify: `backend/internal/usecase/tenant/create_tenant.go`
- Modify: `backend/internal/usecase/tenant/update_tenant.go`
- Modify: `backend/internal/usecase/tenant/deactivate_tenant.go`
- Modify: `backend/internal/usecase/tenant/delete_tenant.go`
- Modify: `backend/internal/usecase/property/create_property.go`
- Modify: `backend/internal/usecase/property/update_property.go`
- Modify: `backend/internal/usecase/property/deactivate_property.go`
- Modify: `backend/internal/usecase/property/delete_property.go`
- Modify: `backend/tests/unit/tenant_usecase_test.go`
- Modify: `backend/tests/unit/property_usecase_test.go`

**Pattern for every use case modification:**
1. Add `auditRepo repository.AuditRepository` field to struct
2. Add `auditRepo` param to constructor
3. Add `uc.auditRepo.Log(ctx, ...)` call after successful mutation

- [ ] **Step 1: Modify tenant use cases**

Each tenant use case gets `auditRepo` added. Here are the complete modified files:

**`create_tenant.go`** — add `auditRepo` field + constructor param + `Log()` after `Create`:
```go
package tenant

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreateTenantUseCase struct {
	repo      repository.TenantRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
}

func NewCreateTenantUseCase(repo repository.TenantRepository, userRepo repository.UserRepository, auditRepo repository.AuditRepository) *CreateTenantUseCase {
	return &CreateTenantUseCase{repo: repo, userRepo: userRepo, auditRepo: auditRepo}
}

func (uc *CreateTenantUseCase) Execute(ctx context.Context, tenant *entity.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return err
	}

	landlord, err := uc.userRepo.GetByID(ctx, tenant.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", tenant.LandlordID)
	}

	if tenant.Email != nil && *tenant.Email != "" {
		existing, err := uc.repo.GetByEmail(ctx, *tenant.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return apperror.NewConflict(entity.ErrTenantEmailExists.Error())
		}
	}

	if err := uc.repo.Create(ctx, tenant); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   tenant.ID,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": tenant.FullName, "landlord_id": tenant.LandlordID.String()},
	})

	return nil
}
```

**`update_tenant.go`:**
```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdateTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

func NewUpdateTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *UpdateTenantUseCase) Execute(ctx context.Context, id uuid.UUID, tenant *entity.Tenant) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	tenant.ID = id
	tenant.CreatedAt = existing.CreatedAt
	tenant.LandlordID = existing.LandlordID

	if err := tenant.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, tenant); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": tenant.FullName, "landlord_id": tenant.LandlordID.String()},
	})

	return nil
}
```

**`deactivate_tenant.go`:**
```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivateTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

func NewDeactivateTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *DeactivateTenantUseCase {
	return &DeactivateTenantUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeactivateTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	existing.IsActive = false
	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DEACTIVATE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": existing.FullName, "landlord_id": existing.LandlordID.String()},
	})

	return nil
}
```

**`delete_tenant.go`:**
```go
package tenant

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeleteTenantUseCase struct {
	repo      repository.TenantRepository
	auditRepo repository.AuditRepository
}

func NewDeleteTenantUseCase(repo repository.TenantRepository, auditRepo repository.AuditRepository) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeleteTenantUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Tenant", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_TENANT",
		ResourceType: "Tenant",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"tenant_name": existing.FullName, "landlord_id": existing.LandlordID.String()},
	})

	return nil
}
```

- [ ] **Step 2: Modify property use cases**

**`create_property.go`:**
```go
package property

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreatePropertyUseCase struct {
	repo      repository.PropertyRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditRepository
}

func NewCreatePropertyUseCase(repo repository.PropertyRepository, userRepo repository.UserRepository, auditRepo repository.AuditRepository) *CreatePropertyUseCase {
	return &CreatePropertyUseCase{repo: repo, userRepo: userRepo, auditRepo: auditRepo}
}

func (uc *CreatePropertyUseCase) Execute(ctx context.Context, property *entity.Property) error {
	if err := property.Validate(); err != nil {
		return err
	}

	owner, err := uc.userRepo.GetByID(ctx, property.OwnerID)
	if err != nil {
		return err
	}
	if owner == nil {
		return apperror.NewNotFound("User", property.OwnerID)
	}

	existing, err := uc.repo.GetByPropertyCode(ctx, property.PropertyCode)
	if err != nil {
		return err
	}
	if existing != nil {
		return apperror.NewConflict(entity.ErrPropertyCodeExists.Error())
	}

	if err := uc.repo.Create(ctx, property); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   property.ID,
		Success:      true,
		Metadata:     map[string]any{"property_name": property.Name, "property_code": property.PropertyCode, "owner_id": property.OwnerID.String()},
	})

	return nil
}
```

**`update_property.go`:**
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdatePropertyUseCase struct {
	repo      repository.PropertyRepository
	auditRepo repository.AuditRepository
}

func NewUpdatePropertyUseCase(repo repository.PropertyRepository, auditRepo repository.AuditRepository) *UpdatePropertyUseCase {
	return &UpdatePropertyUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *UpdatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID, property *entity.Property) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	property.ID = id
	property.CreatedAt = existing.CreatedAt
	property.OwnerID = existing.OwnerID

	if err := property.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, property); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": property.Name, "owner_id": property.OwnerID.String()},
	})

	return nil
}
```

**`deactivate_property.go`:**
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeactivatePropertyUseCase struct {
	repo      repository.PropertyRepository
	debtRepo  repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewDeactivatePropertyUseCase(repo repository.PropertyRepository, debtRepo repository.DebtRepository, auditRepo repository.AuditRepository) *DeactivatePropertyUseCase {
	return &DeactivatePropertyUseCase{repo: repo, debtRepo: debtRepo, auditRepo: auditRepo}
}

func (uc *DeactivatePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	hasDebts, err := uc.debtRepo.HasActiveDebtsForProperty(ctx, id)
	if err != nil {
		return err
	}
	if hasDebts {
		return apperror.NewConflict(entity.ErrPropertyHasActiveDebts.Error())
	}

	existing.IsActive = false
	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DEACTIVATE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": existing.Name, "owner_id": existing.OwnerID.String()},
	})

	return nil
}
```

**`delete_property.go`:**
```go
package property

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeletePropertyUseCase struct {
	repo      repository.PropertyRepository
	auditRepo repository.AuditRepository
}

func NewDeletePropertyUseCase(repo repository.PropertyRepository, auditRepo repository.AuditRepository) *DeletePropertyUseCase {
	return &DeletePropertyUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeletePropertyUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Property", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_PROPERTY",
		ResourceType: "Property",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"property_name": existing.Name, "owner_id": existing.OwnerID.String()},
	})

	return nil
}
```

- [ ] **Step 3: Fix tenant and property test constructors**

In `backend/tests/unit/tenant_usecase_test.go`, update every `NewXxxUseCase` call to add `&mocks.AuditRepositoryMock{}` as the last argument.

In `backend/tests/unit/property_usecase_test.go`, update every `NewXxxUseCase` call to add `&mocks.AuditRepositoryMock{}` as the last argument.

The pattern is the same for all: add `&mocks.AuditRepositoryMock{}` as the final constructor parameter.

- [ ] **Step 4: Verify compilation and run tests**

```bash
cd backend && go build ./...
cd backend && go test ./tests/unit/ -run "TestCreateTenant|TestUpdateTenant|TestDeactivateTenant|TestDeleteTenant|TestCreateProperty|TestUpdateProperty|TestDeactivateProperty|TestDeleteProperty" -v
```

Expected: All tenant and property tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/usecase/tenant/ internal/usecase/property/ tests/unit/tenant_usecase_test.go tests/unit/property_usecase_test.go
git commit -m "feat: inject audit logging into tenant and property use cases"
```

---

## Task 8: Inject Audit into Debt + Transaction Use Cases

**Files:**
- Modify: `backend/internal/usecase/debt/create_debt.go`
- Modify: `backend/internal/usecase/debt/update_debt.go`
- Modify: `backend/internal/usecase/debt/cancel_debt.go`
- Modify: `backend/internal/usecase/debt/mark_debt_paid.go`
- Modify: `backend/internal/usecase/debt/delete_debt.go`
- Modify: `backend/internal/usecase/transaction/record_payment.go`
- Modify: `backend/internal/usecase/transaction/record_refund.go`
- Modify: `backend/internal/usecase/transaction/verify_transaction.go`
- Modify: `backend/tests/unit/debt_usecase_test.go`
- Modify: `backend/tests/unit/transaction_usecase_test.go`

Same pattern as Task 7. Add `auditRepo repository.AuditRepository` to struct, constructor, and `Log()` call after success.

- [ ] **Step 1: Modify debt use cases**

**`create_debt.go`** — add `auditRepo` field, constructor param, `Log()` after Create:
```go
package debt

import (
	"context"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CreateDebtUseCase struct {
	repo       repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	propRepo   repository.PropertyRepository
	auditRepo  repository.AuditRepository
}

func NewCreateDebtUseCase(repo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, propRepo repository.PropertyRepository, auditRepo repository.AuditRepository) *CreateDebtUseCase {
	return &CreateDebtUseCase{repo: repo, userRepo: userRepo, tenantRepo: tenantRepo, propRepo: propRepo, auditRepo: auditRepo}
}

func (uc *CreateDebtUseCase) Execute(ctx context.Context, d *entity.Debt) error {
	if err := d.Validate(); err != nil {
		return err
	}

	landlord, err := uc.userRepo.GetByID(ctx, d.LandlordID)
	if err != nil {
		return err
	}
	if landlord == nil {
		return apperror.NewNotFound("User", d.LandlordID)
	}

	tenant, err := uc.tenantRepo.GetByID(ctx, d.TenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return apperror.NewNotFound("Tenant", d.TenantID)
	}
	if tenant.LandlordID != d.LandlordID {
		return apperror.NewForbidden("tenant does not belong to this landlord")
	}

	if d.PropertyID != nil {
		prop, err := uc.propRepo.GetByID(ctx, *d.PropertyID)
		if err != nil {
			return err
		}
		if prop == nil {
			return apperror.NewNotFound("Property", *d.PropertyID)
		}
		if prop.OwnerID != d.LandlordID {
			return apperror.NewForbidden("property does not belong to this landlord")
		}
	}

	if err := uc.repo.Create(ctx, d); err != nil {
		return err
	}

	metadata := map[string]any{
		"landlord_id":  d.LandlordID.String(),
		"tenant_id":    d.TenantID.String(),
		"amount":       d.OriginalAmount.Amount.String(),
		"currency":     string(d.OriginalAmount.Currency),
		"debt_type":    string(d.DebtType),
		"description":  d.Description,
	}
	if d.PropertyID != nil {
		metadata["property_id"] = d.PropertyID.String()
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CREATE_DEBT",
		ResourceType: "Debt",
		ResourceID:   d.ID,
		Success:      true,
		Metadata:     metadata,
	})

	return nil
}
```

**`update_debt.go`:**
```go
package debt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type UpdateDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewUpdateDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *UpdateDebtUseCase {
	return &UpdateDebtUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *UpdateDebtUseCase) Execute(ctx context.Context, id uuid.UUID, updates *entity.Debt) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	existing.Description = updates.Description
	existing.DebtType = updates.DebtType
	existing.DueDate = updates.DueDate
	existing.PropertyID = updates.PropertyID
	existing.Notes = updates.Notes
	existing.UpdatedAt = time.Now().UTC()

	if err := existing.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "UPDATE_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": existing.LandlordID.String(), "description": existing.Description},
	})

	return nil
}
```

**`cancel_debt.go`:**
```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type CancelDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewCancelDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *CancelDebtUseCase {
	return &CancelDebtUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *CancelDebtUseCase) Execute(ctx context.Context, id uuid.UUID, reason *string) error {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return apperror.NewNotFound("Debt", id)
	}

	if err := d.Cancel(reason); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, d); err != nil {
		return err
	}

	metadata := map[string]any{"landlord_id": d.LandlordID.String(), "tenant_id": d.TenantID.String()}
	if reason != nil {
		metadata["reason"] = *reason
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "CANCEL_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     metadata,
	})

	return nil
}
```

**`mark_debt_paid.go`:**
```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type MarkDebtPaidUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewMarkDebtPaidUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *MarkDebtPaidUseCase {
	return &MarkDebtPaidUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *MarkDebtPaidUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	d, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if d == nil {
		return apperror.NewNotFound("Debt", id)
	}

	balance := d.GetBalance()
	if err := d.RecordPayment(balance); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, d); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "MARK_DEBT_PAID",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": d.LandlordID.String(), "tenant_id": d.TenantID.String(), "amount": d.OriginalAmount.Amount.String()},
	})

	return nil
}
```

**`delete_debt.go`:**
```go
package debt

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type DeleteDebtUseCase struct {
	repo      repository.DebtRepository
	auditRepo repository.AuditRepository
}

func NewDeleteDebtUseCase(repo repository.DebtRepository, auditRepo repository.AuditRepository) *DeleteDebtUseCase {
	return &DeleteDebtUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *DeleteDebtUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NewNotFound("Debt", id)
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "DELETE_DEBT",
		ResourceType: "Debt",
		ResourceID:   id,
		Success:      true,
		Metadata:     map[string]any{"landlord_id": existing.LandlordID.String(), "tenant_id": existing.TenantID.String()},
	})

	return nil
}
```

- [ ] **Step 2: Modify transaction use cases**

**`record_payment.go`** — add `auditRepo`, log after success:
```go
package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type RecordPaymentUseCase struct {
	txRepo     repository.TransactionRepository
	debtRepo   repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	auditRepo  repository.AuditRepository
}

func NewRecordPaymentUseCase(txRepo repository.TransactionRepository, debtRepo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, auditRepo repository.AuditRepository) *RecordPaymentUseCase {
	return &RecordPaymentUseCase{txRepo: txRepo, debtRepo: debtRepo, userRepo: userRepo, tenantRepo: tenantRepo, auditRepo: auditRepo}
}

func (uc *RecordPaymentUseCase) Execute(ctx context.Context, debtID, tenantID uuid.UUID, recordedBy *uuid.UUID, amount entity.Money, method entity.PaymentMethod, txDate time.Time, description string, receipt, reference *string) (*entity.Transaction, error) {
	d, err := uc.debtRepo.GetByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", debtID)
	}

	if reference != nil {
		exists, err := uc.txRepo.ExistsByReferenceNumber(ctx, debtID, *reference)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, entity.ErrTransactionDuplicateReference
		}
	}

	balanceBefore := d.GetBalance()
	if err := d.RecordPayment(amount); err != nil {
		return nil, err
	}
	balanceAfter := d.GetBalance()

	tx, err := entity.NewTransaction(debtID, d.TenantID, d.LandlordID, recordedBy, entity.TransactionTypePayment, amount, method, txDate, description, receipt, reference)
	if err != nil {
		return nil, err
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if err := uc.debtRepo.Update(ctx, d); err != nil {
		return nil, err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "APPLY_PAYMENT",
		ResourceType: "Transaction",
		ResourceID:   tx.ID,
		Success:      true,
		Metadata: map[string]any{
			"landlord_id":    d.LandlordID.String(),
			"tenant_id":     d.TenantID.String(),
			"payment_amount": amount.Amount.String(),
			"currency":      string(amount.Currency),
			"balance_before": balanceBefore.Amount.String(),
			"balance_after":  balanceAfter.Amount.String(),
			"debt_type":     string(d.DebtType),
		},
	})

	return tx, nil
}
```

**`record_refund.go`:**
```go
package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/entity"
	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type RecordRefundUseCase struct {
	txRepo     repository.TransactionRepository
	debtRepo   repository.DebtRepository
	userRepo   repository.UserRepository
	tenantRepo repository.TenantRepository
	auditRepo  repository.AuditRepository
}

func NewRecordRefundUseCase(txRepo repository.TransactionRepository, debtRepo repository.DebtRepository, userRepo repository.UserRepository, tenantRepo repository.TenantRepository, auditRepo repository.AuditRepository) *RecordRefundUseCase {
	return &RecordRefundUseCase{txRepo: txRepo, debtRepo: debtRepo, userRepo: userRepo, tenantRepo: tenantRepo, auditRepo: auditRepo}
}

func (uc *RecordRefundUseCase) Execute(ctx context.Context, debtID, tenantID uuid.UUID, recordedBy *uuid.UUID, amount entity.Money, method entity.PaymentMethod, refundDate time.Time, description string, reference *string) (*entity.Transaction, error) {
	d, err := uc.debtRepo.GetByID(ctx, debtID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperror.NewNotFound("Debt", debtID)
	}

	if err := d.ReversePayment(amount); err != nil {
		return nil, err
	}

	tx, err := entity.NewTransaction(debtID, d.TenantID, d.LandlordID, recordedBy, entity.TransactionTypeRefund, amount, method, refundDate, description, nil, reference)
	if err != nil {
		return nil, err
	}

	if err := uc.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if err := uc.debtRepo.Update(ctx, d); err != nil {
		return nil, err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "APPLY_REFUND",
		ResourceType: "Transaction",
		ResourceID:   tx.ID,
		Success:      true,
		Metadata: map[string]any{
			"landlord_id":   d.LandlordID.String(),
			"tenant_id":    d.TenantID.String(),
			"refund_amount": amount.Amount.String(),
			"currency":     string(amount.Currency),
		},
	})

	return tx, nil
}
```

**`verify_transaction.go`:**
```go
package transaction

import (
	"context"

	"github.com/google/uuid"

	"github.com/ISubamariner/guimba-go/backend/internal/domain/repository"
	"github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

type VerifyTransactionUseCase struct {
	repo      repository.TransactionRepository
	auditRepo repository.AuditRepository
}

func NewVerifyTransactionUseCase(repo repository.TransactionRepository, auditRepo repository.AuditRepository) *VerifyTransactionUseCase {
	return &VerifyTransactionUseCase{repo: repo, auditRepo: auditRepo}
}

func (uc *VerifyTransactionUseCase) Execute(ctx context.Context, id, verifierID uuid.UUID) error {
	tx, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tx == nil {
		return apperror.NewNotFound("Transaction", id)
	}

	if err := tx.Verify(verifierID); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, tx); err != nil {
		return err
	}

	uc.auditRepo.Log(ctx, &repository.AuditEntry{
		Action:       "VERIFY_TRANSACTION",
		ResourceType: "Transaction",
		ResourceID:   id,
		Success:      true,
		Metadata: map[string]any{
			"verified_by_user_id": verifierID.String(),
			"transaction_type":   string(tx.TransactionType),
			"amount":             tx.Amount.Amount.String(),
			"currency":           string(tx.Amount.Currency),
			"landlord_id":        tx.LandlordID.String(),
			"tenant_id":          tx.TenantID.String(),
		},
	})

	return nil
}
```

- [ ] **Step 3: Fix debt and transaction test constructors**

In `backend/tests/unit/debt_usecase_test.go`, update every `NewXxxUseCase` call to add `&mocks.AuditRepositoryMock{}` as the last argument.

In `backend/tests/unit/transaction_usecase_test.go`, update every `NewXxxUseCase` call to add `&mocks.AuditRepositoryMock{}` as the last argument.

- [ ] **Step 4: Verify compilation and run tests**

```bash
cd backend && go build ./...
cd backend && go test ./tests/unit/ -run "TestCreateDebt|TestUpdateDebt|TestCancelDebt|TestMarkDebtPaid|TestDeleteDebt|TestRecordPayment|TestRecordRefund|TestVerifyTransaction" -v
```

Expected: All debt and transaction tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/usecase/debt/ internal/usecase/transaction/ tests/unit/debt_usecase_test.go tests/unit/transaction_usecase_test.go
git commit -m "feat: inject audit logging into debt and transaction use cases"
```

---

## Task 9: Inject Audit into Beneficiary + User + Auth Use Cases

**Files:**
- Modify: `backend/internal/usecase/beneficiary/create_beneficiary.go`
- Modify: `backend/internal/usecase/beneficiary/update_beneficiary.go`
- Modify: `backend/internal/usecase/beneficiary/delete_beneficiary.go`
- Modify: `backend/internal/usecase/beneficiary/enroll_in_program.go`
- Modify: `backend/internal/usecase/beneficiary/remove_from_program.go`
- Modify: `backend/internal/usecase/user/update_user.go`
- Modify: `backend/internal/usecase/user/delete_user.go`
- Modify: `backend/internal/usecase/user/assign_role.go` (contains both AssignRole and RemoveRole)
- Modify: `backend/internal/usecase/auth/register.go`
- Modify: `backend/tests/unit/beneficiary_usecase_test.go`
- Modify: `backend/tests/unit/user_usecase_test.go`
- Modify: `backend/tests/unit/user_auth_test.go`

Same pattern. Add `auditRepo` to struct, constructor, `Log()` after success. For brevity, showing the audit-specific additions — the full file content follows the same pattern as Tasks 7-8 (add `auditRepo repository.AuditRepository` to struct and constructor, add `Log()` call after successful mutation).

- [ ] **Step 1: Modify beneficiary use cases**

Add `auditRepo` to all 5 beneficiary use cases. Actions and metadata:

| File | Action | Metadata |
|------|--------|----------|
| `create_beneficiary.go` | `CREATE_BENEFICIARY` | `beneficiary_name` |
| `update_beneficiary.go` | `UPDATE_BENEFICIARY` | `beneficiary_name` |
| `delete_beneficiary.go` | `DELETE_BENEFICIARY` | `beneficiary_name` |
| `enroll_in_program.go` | `ENROLL_BENEFICIARY` | `beneficiary_id`, `program_id` |
| `remove_from_program.go` | `REMOVE_BENEFICIARY` | `beneficiary_id`, `program_id` |

For `enroll_in_program.go`, note the struct has `beneficiaryRepo` and `programRepo` fields — add `auditRepo` as a third field and constructor param.

Each file follows the exact same pattern shown in Tasks 7-8:
1. Add `auditRepo repository.AuditRepository` to the struct
2. Add `auditRepo` parameter to the constructor
3. Change the final `return uc.repo.XxxMethod(...)` to capture error, check it, then call `uc.auditRepo.Log(...)`, then `return nil`

- [ ] **Step 2: Modify user use cases**

Add `auditRepo` to all 4 user use cases (update, delete, assign_role which has both AssignRole and RemoveRole).

| File | Action | Metadata |
|------|--------|----------|
| `update_user.go` | `UPDATE_USER` | `full_name`, `is_active` |
| `delete_user.go` | `DELETE_USER` | `user_email` |
| `assign_role.go` (AssignRoleUseCase) | `ASSIGN_ROLE` | `role_id`, `target_user_email` |
| `assign_role.go` (RemoveRoleUseCase) | `REMOVE_ROLE` | `role_id`, `target_user_email` |

Note: `assign_role.go` contains TWO use case types. Both get `auditRepo`.

- [ ] **Step 3: Modify register use case**

Add `auditRepo` to `RegisterUseCase`. Action: `REGISTER_USER`. Metadata: `user_email`, `full_name`. Log after successful user creation (before role assignment).

- [ ] **Step 4: Fix test constructors**

In `backend/tests/unit/beneficiary_usecase_test.go`, update all `NewXxxUseCase` calls to add `&mocks.AuditRepositoryMock{}`.

In `backend/tests/unit/user_usecase_test.go`, update all `NewXxxUseCase` calls to add `&mocks.AuditRepositoryMock{}`.

In `backend/tests/unit/user_auth_test.go`, update `NewRegisterUseCase` call to add `&mocks.AuditRepositoryMock{}`.

- [ ] **Step 5: Verify compilation and run all tests**

```bash
cd backend && go build ./...
cd backend && go test ./tests/... -count=1
```

Expected: All tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/usecase/beneficiary/ internal/usecase/user/ internal/usecase/auth/register.go tests/unit/beneficiary_usecase_test.go tests/unit/user_usecase_test.go tests/unit/user_auth_test.go
git commit -m "feat: inject audit logging into beneficiary, user, and auth use cases"
```

---

## Task 10: Router + Main.go Wiring

**Files:**
- Modify: `backend/internal/delivery/http/router/router.go`
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: Update router.go**

Add `Audit` field to the `Handlers` struct:

```go
Audit       *handler.AuditHandler
```

Add the audit route group after the transaction routes:

```go
// Audit logs
r.Route("/audit", func(r chi.Router) {
    r.Use(requireAuth)
    r.With(middleware.RequireRole("admin", "auditor")).Get("/", h.Audit.List)
    r.With(middleware.RequireRole("admin", "landlord")).Get("/landlord", h.Audit.LandlordList)
})
```

Add `middleware.AuditContext` to the global middleware stack (after CORS, before routes):

```go
r.Use(middleware.AuditContext)
```

- [ ] **Step 2: Update main.go**

Add imports:
```go
auditinfra "github.com/ISubamariner/guimba-go/backend/internal/infrastructure/audit"
mongorepo "github.com/ISubamariner/guimba-go/backend/internal/infrastructure/persistence/mongo"
audituc "github.com/ISubamariner/guimba-go/backend/internal/usecase/audit"
```

After Redis connection block, add audit wiring:
```go
// Wire Audit system
mongoAuditRepo := mongorepo.NewAuditRepoMongo(mongoClient, cfg.Mongo.DB)
if err := mongoAuditRepo.EnsureIndexes(ctx); err != nil {
    slog.Warn("failed to create audit indexes", "error", err)
}
auditLogger := auditinfra.NewBufferedAuditLogger(mongoAuditRepo, redisClient)
auditCtx, auditCancel := context.WithCancel(ctx)
go auditLogger.Start(auditCtx)
```

Update ALL mutating use case constructors to pass `auditLogger` as the last argument. Example:
```go
createTenantUC := tenantuc.NewCreateTenantUseCase(tenantRepo, userRepo, auditLogger)
```

Wire audit query use cases and handler:
```go
listAuditLogsUC := audituc.NewListAuditLogsUseCase(auditLogger)
listLandlordAuditLogsUC := audituc.NewListLandlordAuditLogsUseCase(auditLogger)
auditHandler := handler.NewAuditHandler(listAuditLogsUC, listLandlordAuditLogsUC)
```

Add `Audit: auditHandler` to the `router.Handlers` struct initialization.

Update shutdown sequence — cancel audit AFTER server shutdown:
```go
if err := srv.Shutdown(shutdownCtx); err != nil {
    slog.Error("server forced to shutdown", "error", err)
}
auditCancel()
auditLogger.Stop()
slog.Info("server stopped")
```

- [ ] **Step 3: Verify compilation**

```bash
cd backend && go build ./cmd/server/main.go
```

Expected: Clean compilation.

- [ ] **Step 4: Run full test suite**

```bash
cd backend && go test ./tests/... -count=1
```

Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/delivery/http/router/router.go cmd/server/main.go
git commit -m "feat: wire audit system in router and main.go"
```

---

## Task 11: CLAUDE.md Update + Final Verification

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md routes table**

Add audit routes to the API Routes table:

```markdown
| `/audit` | GET | Required | admin, auditor |
| `/audit/landlord` | GET | Required | admin, landlord |
```

- [ ] **Step 2: Update Current Modules**

Update the "Current Modules" section to include Audit:

```markdown
Programs, Users & Auth (JWT + RBAC), Beneficiaries (with program enrollment), Tenants (landlord-scoped CRUD with Address value object, deactivation), Properties (landlord-scoped with deactivation), Debts & Transactions (Money value object, debt state machine, payment/refund orchestration, lazy overdue detection, transaction verification), Audit (MongoDB audit logs, Redis-buffered durability, cross-cutting mutation logging).
```

- [ ] **Step 3: Run final test suite**

```bash
cd backend && go test ./tests/... -v -count=1 2>&1 | tail -20
```

Expected: All tests PASS.

- [ ] **Step 4: Verify build**

```bash
cd backend && go build ./cmd/server/main.go
```

Expected: Clean build.

- [ ] **Step 5: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with audit system module"
```
