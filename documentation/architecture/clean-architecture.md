# Clean Architecture

## Overview

The Go backend follows **Clean Architecture** (Robert C. Martin) where dependencies point inward. Each layer only knows about the layer directly inside it.

```
┌──────────────────────────────────────────────┐
│          Frameworks & Drivers (outer)         │
│   HTTP handlers, DB drivers, Redis, CLI       │
├──────────────────────────────────────────────┤
│          Interface Adapters                   │
│   Controllers, Gateways, Presenters           │
├──────────────────────────────────────────────┤
│          Use Cases (Application Logic)        │
│   Application-specific business rules         │
├──────────────────────────────────────────────┤
│          Entities (Domain Core)               │
│   Enterprise-wide business rules & types      │
└──────────────────────────────────────────────┘
        Dependencies point INWARD only →
```

## Mapping to Our Folder Structure

| Layer | Directory | Contents |
|:---|:---|:---|
| **Domain** (innermost) | `internal/domain/` | `entity/` (structs), `repository/` (interfaces), `service/` (interfaces), `valueobject/` |
| **Use Cases** | `internal/usecase/` | One file per use case, grouped by aggregate (`program/`, `user/`) |
| **Infrastructure** | `internal/infrastructure/` | `persistence/pg/` (Postgres repos), `persistence/mongo/` (MongoDB repos), `database/`, `cache/`, `config/`, `external/` |
| **Delivery** (outermost) | `internal/delivery/http/` | `handler/`, `middleware/`, `router/`, `dto/` |

## Dependency Rules

```
delivery/handler → usecase → domain/repository (interface)
                                    ↑
infrastructure/persistence ─────────┘  (implements the interface)
```

| Layer | Can Import | Cannot Import |
|:---|:---|:---|
| `domain/` | Only Go stdlib | `usecase/`, `infrastructure/`, `delivery/`, any third-party package |
| `usecase/` | `domain/` | `infrastructure/`, `delivery/` |
| `infrastructure/` | `domain/` (to implement interfaces) | `usecase/`, `delivery/` |
| `delivery/` | `usecase/`, `domain/` | `infrastructure/` (injected at startup in `main.go`) |

## Dependency Injection

All wiring happens in `cmd/server/main.go`:

```go
func main() {
    // 1. Load config (Viper-based, reads .env + env vars)
    cfg, _ := config.Load()

    // 2. Create infrastructure (outermost — DB pools, cache clients)
    pgPool, _ := database.NewPostgresPool(ctx, cfg.Postgres.DSN)
    mongoClient, _ := database.NewMongoClient(ctx, cfg.Mongo.URI)
    redisClient, _ := cache.NewRedisClient(ctx, cfg.Redis.Addr, cfg.Redis.Password)

    // 3. Run migrations
    database.RunMigrations(cfg.Postgres.DSN, "migrations")

    // 4. Create handlers (wired with infrastructure)
    healthHandler := handler.NewHealthHandler(pgPool, mongoClient, redisClient)

    // 5. Wire Program module (repo → use cases → handler)
    programRepo := pg.NewProgramRepoPG(pgPool)
    createProgramUC := programuc.NewCreateProgramUseCase(programRepo)
    getProgramUC := programuc.NewGetProgramUseCase(programRepo)
    listProgramsUC := programuc.NewListProgramsUseCase(programRepo)
    updateProgramUC := programuc.NewUpdateProgramUseCase(programRepo)
    deleteProgramUC := programuc.NewDeleteProgramUseCase(programRepo)
    programHandler := handler.NewProgramHandler(createProgramUC, getProgramUC, listProgramsUC, updateProgramUC, deleteProgramUC)

    // 6. Wire Auth module (JWT manager + blocklist + repos → use cases → handler)
    jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessDuration, cfg.JWT.RefreshDuration)
    tokenBlocklist := cache.NewTokenBlocklist(redisClient)
    userRepo := pg.NewUserRepoPG(pgPool)
    roleRepo := pg.NewRoleRepoPG(pgPool)
    registerUC := authuc.NewRegisterUseCase(userRepo, roleRepo)
    loginUC := authuc.NewLoginUseCase(userRepo)
    refreshTokenUC := authuc.NewRefreshTokenUseCase(userRepo, jwtManager, tokenBlocklist)
    getProfileUC := authuc.NewGetProfileUseCase(userRepo)
    authHandler := handler.NewAuthHandler(registerUC, loginUC, refreshTokenUC, getProfileUC, jwtManager, tokenBlocklist)

    // 7. Wire User module (use cases → handler)
    listUsersUC := useruc.NewListUsersUseCase(userRepo)
    updateUserUC := useruc.NewUpdateUserUseCase(userRepo)
    deleteUserUC := useruc.NewDeleteUserUseCase(userRepo)
    assignRoleUC := useruc.NewAssignRoleUseCase(userRepo, roleRepo)
    userHandler := handler.NewUserHandler(listUsersUC, updateUserUC, deleteUserUC, assignRoleUC)

    // 8. Wire Beneficiary module (repo → use cases → handler)
    beneficiaryRepo := pg.NewBeneficiaryRepoPG(pgPool)
    createBeneficiaryUC := beneficiaryuc.NewCreateBeneficiaryUseCase(beneficiaryRepo)
    getBeneficiaryUC := beneficiaryuc.NewGetBeneficiaryUseCase(beneficiaryRepo)
    listBeneficiariesUC := beneficiaryuc.NewListBeneficiariesUseCase(beneficiaryRepo)
    updateBeneficiaryUC := beneficiaryuc.NewUpdateBeneficiaryUseCase(beneficiaryRepo)
    deleteBeneficiaryUC := beneficiaryuc.NewDeleteBeneficiaryUseCase(beneficiaryRepo)
    enrollInProgramUC := beneficiaryuc.NewEnrollInProgramUseCase(beneficiaryRepo, programRepo)
    removeFromProgramUC := beneficiaryuc.NewRemoveFromProgramUseCase(beneficiaryRepo)
    beneficiaryHandler := handler.NewBeneficiaryHandler(...)

    // 9. Wire router with auth middleware
    r := router.NewRouter(router.Handlers{..., Beneficiary: beneficiaryHandler}, jwtManager, tokenBlocklist, cfg.App.FrontendURL)

    // 10. Start server with graceful shutdown
    srv := &http.Server{Addr: ":" + cfg.App.Port, Handler: r}
}
```

> **Programs + Auth + Users + Beneficiaries** modules are fully wired, demonstrating the DI pattern: `PG repos → use cases → handlers → router`, with auth middleware injected at the router level.

## Polyglot Persistence & Clean Architecture

The domain layer defines repository interfaces without knowing which database implements them:

```go
// domain/repository/audit_log_repository.go — doesn't know about MongoDB
type AuditLogRepository interface {
    Log(ctx context.Context, entry AuditEntry) error
    FindByEntity(ctx context.Context, entityType string, entityID string) ([]AuditEntry, error)
}

// infrastructure/persistence/mongo/audit_log_repo_mongo.go — implements with MongoDB
type auditLogRepoMongo struct {
    collection *mongo.Collection
}
```

This means you can swap databases per module without touching domain or use case code.

## Why Clean Architecture?

1. **Testability** — Domain and use cases have no framework dependencies; easily tested with mocks
2. **Flexibility** — Swap PostgreSQL for another DB by implementing the same repository interface
3. **Readability** — Clear boundaries; you know where to find business rules vs. HTTP logic vs. DB queries
4. **Maintainability** — Changes in one layer don't cascade to others (e.g., changing the HTTP framework doesn't touch business logic)
