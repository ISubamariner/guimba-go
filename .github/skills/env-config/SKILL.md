---
name: env-config
description: "Manages environment configuration, .env files, config validation, and secret handling across environments. Use when user says 'add env variable', 'configure', 'environment', '.env', 'config', 'secret', 'production config', or when working with infrastructure/config/ or .env files."
---

# Environment Configuration

Manages configuration loading, validation, and secret handling.

## Config Struct (`infrastructure/config/config.go`)

The actual implementation uses **Viper** for configuration loading:

```go
type Config struct {
    App      AppConfig
    Postgres PostgresConfig
    Mongo    MongoConfig
    Redis    RedisConfig
    JWT      JWTConfig
}

type AppConfig struct {
    Port        string `mapstructure:"APP_PORT"`
    Env         string `mapstructure:"APP_ENV"`
    LogLevel    string `mapstructure:"LOG_LEVEL"`
    FrontendURL string `mapstructure:"FRONTEND_URL"`
}

type PostgresConfig struct {
    Host     string `mapstructure:"POSTGRES_HOST"`
    Port     string `mapstructure:"POSTGRES_PORT"`
    User     string `mapstructure:"POSTGRES_USER"`
    Password string `mapstructure:"POSTGRES_PASSWORD"`
    DB       string `mapstructure:"POSTGRES_DB"`
    DSN      string `mapstructure:"POSTGRES_DSN"`
}

type MongoConfig struct {
    URI string `mapstructure:"MONGO_URI"`
    DB  string `mapstructure:"MONGO_DB"`
}

type RedisConfig struct {
    Addr     string `mapstructure:"REDIS_ADDR"`
    Password string `mapstructure:"REDIS_PASSWORD"`
}

type JWTConfig struct {
    Secret     string        `mapstructure:"JWT_SECRET"`
    Expiration time.Duration `mapstructure:"JWT_EXPIRATION"`
}
```

## Config Loading

```go
func Load() (*Config, error) {
    v := viper.New()
    v.SetConfigFile(".env")
    v.SetConfigType("env")
    v.AddConfigPath(".")
    v.AutomaticEnv()

    // Read .env file (ignore error if file doesn't exist)
    _ = v.ReadInConfig()

    setDefaults(v)

    // Build config from Viper values
    cfg := &Config{...}

    // Auto-construct DSN/URI if not explicitly provided
    if cfg.Postgres.DSN == "" {
        cfg.Postgres.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", ...)
    }

    // Validate required fields
    if err := validate(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

## Config Validation (Fail Fast)

```go
func validate(cfg *Config) error {
    var errs []string
    if cfg.App.Port == "" {
        errs = append(errs, "APP_PORT is required")
    }
    if cfg.Postgres.DSN == "" {
        errs = append(errs, "POSTGRES_DSN is required")
    }
    if cfg.JWT.Secret == "change-me-in-production" && cfg.App.Env == "production" {
        errs = append(errs, "JWT_SECRET must be set in production")
    }
    if len(errs) > 0 {
        return fmt.Errorf("%s", strings.Join(errs, "; "))
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
| `.env` | Local development defaults | Always (via Viper's `ReadInConfig()`) |
| `.env.test` | Test environment overrides | In test setup |
| Environment vars | Production values | Set by deployment platform |

**Priority**: Environment variables > `.env` file > Viper defaults (set via `SetDefault()`)

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
