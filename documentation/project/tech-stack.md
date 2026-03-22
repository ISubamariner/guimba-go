# Tech Stack

## Backend

| Component | Technology | Rationale |
|:---|:---|:---|
| Language | **Go 1.24+** | Performance, simplicity, strong stdlib, excellent concurrency |
| HTTP Router | **Chi** | Lightweight, idiomatic, middleware-compatible, stdlib `net/http` compatible |
| Database | **PostgreSQL 16+** (via Docker) | Robust relational DB, ACID transactions, foreign keys, complex joins |
| DB Driver (PG) | **pgx** (jackc/pgx) | Best pure-Go Postgres driver, connection pooling built-in |
| Database | **MongoDB 7+** (via Docker) | Document store, audit logs, CQRS read models, flexible schema |
| DB Driver (Mongo) | **mongo-go-driver** | Official MongoDB Go driver |
| Cache | **Redis 7+** (via Docker) | Fast in-memory cache, pub/sub capable |
| Redis Client | **go-redis** | Most popular, full feature set |
| Migrations | **golang-migrate** | SQL-file based, CLI + library, reversible migrations |
| API Docs | **swaggo/swag** | Generates Swagger/OpenAPI from Go comments |
| Config | **viper** | Reads env vars, config files, supports multiple formats |
| Auth | **golang-jwt/jwt** | Standard JWT library for Go |
| Validation | **go-playground/validator** | Struct tag-based validation |
| Logging | **slog** (stdlib) | Structured logging, built into Go 1.21+ |

## Frontend

| Component | Technology | Rationale |
|:---|:---|:---|
| Framework | **Next.js 15+** | React-based, SSR/SSG, App Router, excellent DX |
| Language | **TypeScript** (strict mode) | Type safety, better IDE support, catches bugs at compile time |
| Styling | **Tailwind CSS** | Utility-first, consolidated via design system tokens |
| Design System | **Custom** (tokens.css + components.css + ui/ primitives) | Single source of truth for all visual decisions |
| Class Utility | **clsx + tailwind-merge** (`cn()`) | Conditional class composition without conflicts |

## Testing

| Layer | Technology | Scope |
|:---|:---|:---|
| Go Unit Tests | **stdlib testing** + table-driven | Domain logic, use cases, handlers (mocked) |
| Go Integration | **testcontainers-go** | Repository implementations against real PostgreSQL |
| Browser E2E | **Playwright** | UI flows, full-stack validation, visual regression |
| Frontend Unit | **Jest/Vitest** | React components, hooks, API client |

## Infrastructure

| Component | Technology | Rationale |
|:---|:---|:---|
| Containers | **Docker + Docker Compose** | Reproducible local dev, same as CI |
| CI/CD | **GitHub Actions** | Native to GitHub, free for open source |
| Hot Reload (Go) | **air** (cosmtrek/air) | Live-reload Go server during development |
