---
name: env-config
description: "Manages environment configuration, .env files, config validation, and secret handling across environments. Use when user says 'add env variable', 'configure', 'environment', '.env', 'config', 'secret', 'production config', or when working with infrastructure/config/ or .env files."
---

# Environment Configuration

Manages configuration loading, validation, and secret handling.

## Config Struct (`infrastructure/config/config.go`)

```go
type Config struct {
    // Application
    AppPort string `env:"APP_PORT" default:"8080"`
    AppEnv  string `env:"APP_ENV" default:"development"`

    // PostgreSQL
    PostgresDSN string `env:"POSTGRES_DSN" required:"true"`

    // MongoDB
    MongoURI string `env:"MONGO_URI" required:"true"`
    MongoDB  string `env:"MONGO_DB" default:"guimba_db"`

    // Redis
    RedisAddr     string `env:"REDIS_ADDR" default:"localhost:6380"`
    RedisPassword string `env:"REDIS_PASSWORD" required:"true"`

    // Auth
    JWTSecret          string `env:"JWT_SECRET" required:"true"`
    AccessTokenExpiry   time.Duration `env:"ACCESS_TOKEN_EXPIRY" default:"15m"`
    RefreshTokenExpiry  time.Duration `env:"REFRESH_TOKEN_EXPIRY" default:"168h"` // 7 days

    // Logging
    LogLevel string `env:"LOG_LEVEL" default:"info"`

    // Frontend
    FrontendURL string `env:"FRONTEND_URL" default:"http://localhost:3000"`
}
```

## Config Loading

```go
func Load() (*Config, error) {
    // 1. Load .env file (if exists)
    godotenv.Load() // doesn't error if file missing

    // 2. Read from environment
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("config: %w", err)
    }

    // 3. Validate required fields
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}
```

## Config Validation (Fail Fast)

```go
func (c *Config) Validate() error {
    var errs []string
    if c.PostgresDSN == "" {
        errs = append(errs, "POSTGRES_DSN is required")
    }
    if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
        errs = append(errs, "JWT_SECRET must be at least 32 characters")
    }
    if len(errs) > 0 {
        return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }
    return nil
}
```

If required vars are missing, the app crashes on startup with a clear error message.

## `.env.example` Rules

- Every config field must have a corresponding entry in `.env.example`
- Use realistic placeholder values (not actual secrets)
- Add comments explaining each variable
- Keep in sync with the Config struct

## Environment Hierarchy

| File | Purpose | Loaded |
|:---|:---|:---|
| `.env` | Local development defaults | Always (via `godotenv.Load()`) |
| `.env.test` | Test environment overrides | In test setup |
| Environment vars | Production values | Set by deployment platform |

**Priority**: Environment variables > `.env` file > struct defaults

## Secret Handling Rules

- **Never commit `.env`** — only `.env.example`
- **Never log secrets** — mask with `*****` in debug output
- Use `*****` when printing config for debugging:
  ```go
  func (c *Config) LogSafe() {
      slog.Info("config loaded",
          "app_port", c.AppPort,
          "app_env", c.AppEnv,
          "postgres_dsn", "****",
          "jwt_secret", "****",
      )
  }
  ```
- Rotate secrets regularly in production
- Use environment-scoped secrets (different per environment)

## Frontend Environment

Next.js uses `NEXT_PUBLIC_` prefix for client-exposed vars:

```env
# .env.local (frontend)
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Guimba
```

**Rule**: Never put secrets in `NEXT_PUBLIC_*` variables — they're embedded in the client bundle.

## Docker Compose Integration

```yaml
services:
  backend:
    env_file: .env
    environment:
      POSTGRES_DSN: postgres://guimba:guimba_secret@postgres:5432/guimba_db?sslmode=disable
      MONGO_URI: mongodb://guimba:guimba_secret@mongodb:27017/guimba_db?authSource=admin
      REDIS_ADDR: redis:6379
```

## Rules

- Every new service connection needs a Config struct field
- Every Config field needs an `.env.example` entry
- Fail fast on missing required config at startup
- Never hardcode connection strings — always use config
- Use meaningful defaults for development, require explicit values for production
