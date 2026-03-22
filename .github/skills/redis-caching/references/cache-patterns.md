# Redis Cache Patterns Reference

## Cache Wrapper Repository Template

```go
package cache

import (
    "context"
    "encoding/json"
    "log/slog"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/ISubamariner/guimba-go/backend/internal/domain"
)

type DebtCacheRepo struct {
    inner domain.DebtRepository
    redis *redis.Client
    ttl   time.Duration
}

func NewDebtCacheRepo(inner domain.DebtRepository, redis *redis.Client) *DebtCacheRepo {
    return &DebtCacheRepo{
        inner: inner,
        redis: redis,
        ttl:   15 * time.Minute,
    }
}

func (r *DebtCacheRepo) GetByID(ctx context.Context, id string) (*domain.Debt, error) {
    key := "debt:" + id

    cached, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var debt domain.Debt
        if json.Unmarshal(cached, &debt) == nil {
            return &debt, nil
        }
        slog.Warn("cache: failed to unmarshal", "key", key)
    }

    debt, err := r.inner.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if data, mErr := json.Marshal(debt); mErr == nil {
        if sErr := r.redis.Set(ctx, key, data, r.ttl).Err(); sErr != nil {
            slog.Warn("cache: failed to set", "key", key, "error", sErr)
        }
    }

    return debt, nil
}

func (r *DebtCacheRepo) Create(ctx context.Context, debt *domain.Debt) error {
    if err := r.inner.Create(ctx, debt); err != nil {
        return err
    }
    r.invalidatePattern(ctx, "debt:list:*")
    return nil
}

func (r *DebtCacheRepo) Update(ctx context.Context, debt *domain.Debt) error {
    if err := r.inner.Update(ctx, debt); err != nil {
        return err
    }
    r.redis.Del(ctx, "debt:"+debt.ID)
    r.invalidatePattern(ctx, "debt:list:*")
    return nil
}

func (r *DebtCacheRepo) Delete(ctx context.Context, id string) error {
    if err := r.inner.Delete(ctx, id); err != nil {
        return err
    }
    r.redis.Del(ctx, "debt:"+id)
    r.invalidatePattern(ctx, "debt:list:*")
    return nil
}

func (r *DebtCacheRepo) invalidatePattern(ctx context.Context, pattern string) {
    iter := r.redis.Scan(ctx, 0, pattern, 100).Iterator()
    for iter.Next(ctx) {
        r.redis.Del(ctx, iter.Val())
    }
}
```

## Key Naming Examples

```
# Single entities
debt:550e8400-e29b-41d4-a716-446655440000
user:7c9e6679-7425-40de-944b-e07fc1f90ae7
borrower:a1b2c3d4-e5f6-7890-abcd-ef1234567890

# Lists (hash of query params)
debt:list:page=1&per_page=20&sort=created_at      → debt:list:8f14e45f
debt:list:page=2&per_page=20&status=pending        → debt:list:c4ca4238

# Counts
debt:count
debt:count:status=pending
borrower:count

# Sessions
user:session:eyJhbGciOiJIUzI1NiIs...
```

## TTL Cheat Sheet

| Data Type | TTL | Rationale |
|:---|:---|:---|
| Search/list results | 1 min | Changes frequently, many variants |
| Single entity by ID | 15 min | Moderate change rate |
| User profile | 5 min | May change (name, avatar) |
| Role definitions | 1 hour | Rarely changes |
| System config | 24 hours | Almost never changes |
| Auth session | Match token expiry | Security requirement |

## Invalidation Patterns

### On Create
```
Invalidate: {entity}:list:*, {entity}:count*
Reason: New item affects all list queries and counts
```

### On Update
```
Invalidate: {entity}:{id}, {entity}:list:*, {entity}:count*
Reason: Updated item may affect any list that includes it
```

### On Delete
```
Invalidate: {entity}:{id}, {entity}:list:*, {entity}:count*
Reason: Removed item affects all views
```

### On Bulk Operations
```
Invalidate: {entity}:* (flush all keys for that entity)
Reason: Too many individual invalidations — safer to flush
```
