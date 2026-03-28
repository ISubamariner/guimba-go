package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
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
	mu     sync.Mutex
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
	b.mu.Lock()
	b.cancel = cancel
	b.mu.Unlock()

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
	b.mu.Lock()
	cancel := b.cancel
	b.mu.Unlock()
	if cancel != nil {
		cancel()
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
