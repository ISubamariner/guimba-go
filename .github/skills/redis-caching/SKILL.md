---
name: redis-caching
description: "Manages Redis caching patterns: cache-aside, TTL strategy, key naming, invalidation, and session storage. Use when user says 'add cache', 'cache this', 'Redis', 'invalidate cache', 'cache strategy', 'TTL', 'session store', or when working with infrastructure/cache/."
---

# Redis Caching

Manages Redis caching patterns for the Guimba-GO backend.

## Cache-Aside Pattern

```
Read:  Check cache → hit? return cached → miss? query DB → store in cache → return
Write: Write to DB → invalidate cache key
```

## Key Naming Convention

| Pattern | Example | Use |
|:---|:---|:---|
| `{entity}:{id}` | `debt:550e8400-...` | Single entity by ID |
| `{entity}:list:{hash}` | `debt:list:a1b2c3` | Paginated/filtered list (hash of query params) |
| `{entity}:count` | `debt:count` | Total count cache |
| `user:session:{token}` | `user:session:abc123` | Session/token storage |
| `token_blocklist:{jti}` | `token_blocklist:abc-def` | JWT blocklist (logout/rotation) |

## TTL Strategy

| Category | TTL | Examples |
|:---|:---|:---|
| Short | 1–5 min | List queries, search results, counts |
| Medium | 15–30 min | Single entity by ID |
| Long | 1–24 hours | Config values, role definitions, static lookups |
| Token-bound | remaining token lifetime | JWT blocklist entries (auto-expire with token) |

## Cache Wrapper Repository Pattern

```go
// internal/infrastructure/cache/debt_cache_repo.go
type DebtCacheRepo struct {
    inner  domain.DebtRepository  // real DB repo
    redis  *redis.Client
    ttl    time.Duration
}

func (r *DebtCacheRepo) GetByID(ctx context.Context, id string) (*domain.Debt, error) {
    key := "debt:" + id

    // Try cache
    cached, err := r.redis.Get(ctx, key).Bytes()
    if err == nil {
        var debt domain.Debt
        if json.Unmarshal(cached, &debt) == nil {
            return &debt, nil
        }
    }

    // Cache miss — query DB
    debt, err := r.inner.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache
    data, _ := json.Marshal(debt)
    r.redis.Set(ctx, key, data, r.ttl)

    return debt, nil
}
```

## Cache Invalidation

```go
func (r *DebtCacheRepo) Create(ctx context.Context, debt *domain.Debt) error {
    err := r.inner.Create(ctx, debt)
    if err != nil {
        return err
    }
    // Invalidate list caches (use pattern delete)
    r.invalidateListCaches(ctx, "debt:list:*")
    return nil
}

func (r *DebtCacheRepo) Update(ctx context.Context, debt *domain.Debt) error {
    err := r.inner.Update(ctx, debt)
    if err != nil {
        return err
    }
    // Invalidate specific + list caches
    r.redis.Del(ctx, "debt:"+debt.ID)
    r.invalidateListCaches(ctx, "debt:list:*")
    return nil
}
```

## Redis Client Setup

```go
// internal/infrastructure/cache/redis.go
func NewRedisClient(cfg *config.Config) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
        DB:       0,
        PoolSize: 10,
    })
}
```

## Error Handling Rules

- **Cache miss is NOT an error** — fall through to DB silently
- **Cache failure falls through to DB** — never fail a request because cache is down
- Log cache errors at `warn` level, not `error`
- Always set a TTL — never cache without expiry

## Serialization

- Use JSON for complex objects (`json.Marshal`/`json.Unmarshal`)
- Use raw strings for simple values (counters, flags)
- Always handle deserialization errors gracefully (treat as cache miss)
