# Docker Compose Patterns

## Health Checks
Always add health checks so dependent services wait for readiness:

```yaml
postgres:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U spmis"]
    interval: 5s
    timeout: 5s
    retries: 5

mongodb:
  healthcheck:
    test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
    interval: 5s
    timeout: 5s
    retries: 5

redis:
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s
    timeout: 5s
    retries: 5
```

## Depends On with Condition
```yaml
backend:
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
```

## Named Volumes for Persistence
```yaml
volumes:
  postgres_data:
  redis_data:
```

## Environment Variable Patterns
Use `.env` file with `env_file` directive:
```yaml
services:
  postgres:
    env_file: .env
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
```
