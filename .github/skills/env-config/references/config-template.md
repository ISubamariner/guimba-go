# Config Templates Reference

## Go Config Struct Template

```go
// internal/infrastructure/config/config.go
package config

import (
    "fmt"
    "log/slog"
    "strings"
    "time"

    "github.com/joho/godotenv"
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    // Application
    AppPort  string `envconfig:"APP_PORT" default:"8080"`
    AppEnv   string `envconfig:"APP_ENV" default:"development"`

    // PostgreSQL
    PostgresDSN string `envconfig:"POSTGRES_DSN" required:"true"`

    // MongoDB
    MongoURI string `envconfig:"MONGO_URI" required:"true"`
    MongoDB  string `envconfig:"MONGO_DB" default:"guimba_db"`

    // Redis
    RedisAddr     string `envconfig:"REDIS_ADDR" default:"localhost:6380"`
    RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`

    // JWT Auth
    JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
    AccessTokenExpiry   time.Duration `envconfig:"ACCESS_TOKEN_EXPIRY" default:"15m"`
    RefreshTokenExpiry  time.Duration `envconfig:"REFRESH_TOKEN_EXPIRY" default:"168h"`

    // Logging
    LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

    // Frontend (for CORS)
    FrontendURL string `envconfig:"FRONTEND_URL" default:"http://localhost:3000"`
}

func (c *Config) IsProd() bool {
    return c.AppEnv == "production"
}

func (c *Config) IsDev() bool {
    return c.AppEnv == "development"
}
```

## Config Loader Function Template

```go
func Load() (*Config, error) {
    // Load .env file (non-fatal if missing)
    _ = godotenv.Load()

    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }

    if err := validate(&cfg); err != nil {
        return nil, err
    }

    cfg.logSafe()
    return &cfg, nil
}

func validate(cfg *Config) error {
    var errs []string

    if cfg.PostgresDSN == "" {
        errs = append(errs, "POSTGRES_DSN is required")
    }
    if cfg.MongoURI == "" {
        errs = append(errs, "MONGO_URI is required")
    }
    if cfg.RedisPassword == "" {
        errs = append(errs, "REDIS_PASSWORD is required")
    }
    if len(cfg.JWTSecret) < 32 {
        errs = append(errs, "JWT_SECRET must be at least 32 characters")
    }

    if len(errs) > 0 {
        return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
    }
    return nil
}

func (c *Config) logSafe() {
    slog.Info("configuration loaded",
        "app_port", c.AppPort,
        "app_env", c.AppEnv,
        "log_level", c.LogLevel,
        "postgres_dsn", maskDSN(c.PostgresDSN),
        "mongo_uri", maskDSN(c.MongoURI),
        "redis_addr", c.RedisAddr,
        "frontend_url", c.FrontendURL,
    )
}

func maskDSN(dsn string) string {
    // Show host, hide credentials
    if idx := strings.Index(dsn, "@"); idx != -1 {
        return "****@" + dsn[idx+1:]
    }
    return "****"
}
```

## `.env.example` Annotated Template

```env
# ===========================================
# Guimba-GO Environment Configuration
# ===========================================
# Copy this file to .env and fill in actual values.
# NEVER commit .env to version control.

# --- Application ---
APP_PORT=8080                          # HTTP server port
APP_ENV=development                    # development | test | production

# --- PostgreSQL ---
POSTGRES_USER=guimba                   # DB username
POSTGRES_PASSWORD=guimba_secret        # DB password (change in production!)
POSTGRES_DB=guimba_db                  # Database name
POSTGRES_PORT=5432                     # DB port
POSTGRES_HOST=localhost                # DB host (use 'postgres' in Docker)
POSTGRES_DSN=postgres://guimba:guimba_secret@localhost:5432/guimba_db?sslmode=disable

# --- MongoDB ---
MONGO_USER=guimba                      # MongoDB username
MONGO_PASSWORD=guimba_secret           # MongoDB password (change in production!)
MONGO_DB=guimba_db                     # MongoDB database name
MONGO_PORT=27017                       # MongoDB port
MONGO_HOST=localhost                   # MongoDB host (use 'mongodb' in Docker)
MONGO_URI=mongodb://guimba:guimba_secret@localhost:27017/guimba_db?authSource=admin

# --- Redis ---
REDIS_PASSWORD=guimba_secret           # Redis password (change in production!)
REDIS_PORT=6380                        # Redis port (mapped from container's 6379)
REDIS_HOST=localhost                   # Redis host (use 'redis' in Docker)
REDIS_ADDR=localhost:6380              # Redis address for Go client

# --- Auth ---
JWT_SECRET=change-me-in-production     # Must be 32+ chars in production
ACCESS_TOKEN_EXPIRY=15m                # Access token lifetime
REFRESH_TOKEN_EXPIRY=168h              # Refresh token lifetime (7 days)

# --- Logging ---
LOG_LEVEL=debug                        # debug | info | warn | error

# --- Frontend ---
FRONTEND_PORT=3000                     # Next.js dev server port
FRONTEND_URL=http://localhost:3000     # For CORS allowed origins
NEXT_PUBLIC_API_URL=http://localhost:8080  # API URL exposed to browser
```
