---
applyTo: "**/*.sql,backend/migrations/**,backend/internal/infrastructure/persistence/**"
---

# Database Instructions

## Polyglot Persistence

This project uses **two databases**. Domain repository interfaces are database-agnostic — implementations decide which DB to use.

| Database | Location | Used For |
|:---|:---|:---|
| **PostgreSQL 16+** | `persistence/pg/` | Relational data: users, programs, beneficiaries, roles |
| **MongoDB 7+** | `persistence/mongo/` | Audit logs, document storage, CQRS read models |

## PostgreSQL — Migration Rules
- Use `golang-migrate` format: `{timestamp}_{description}.up.sql` / `.down.sql`
- Every `.up.sql` must have a corresponding `.down.sql`
- Down migrations must be reversible (no data loss if possible)
- Never use `DROP TABLE` in up migrations without a backup strategy

## SQL Standards
- Table names: `snake_case`, plural (`social_programs`, `user_roles`)
- Column names: `snake_case` (`created_at`, `first_name`)
- Primary keys: `id` (UUID preferred, or `BIGSERIAL`)
- Timestamps: always include `created_at` and `updated_at`
- Soft deletes: use `deleted_at TIMESTAMPTZ NULL` where appropriate
- Foreign keys: always named explicitly (`fk_<table>_<column>`)
- Indexes: named `idx_<table>_<columns>`

## Query Safety
- Always use parameterized queries (`$1`, `$2`)
- Never concatenate user input into SQL strings
- Use `BEGIN`/`COMMIT`/`ROLLBACK` for multi-statement operations

## MongoDB — Collection Standards
- Collection names: `snake_case`, plural (`audit_logs`, `documents`)
- Document IDs: use MongoDB's default `_id` (ObjectID) or application-generated UUIDs
- Timestamps: always include `created_at` and `updated_at` fields
- Use `bson` struct tags on Go structs: `bson:"field_name"`
- Indexes: create via Go code in `database/mongodb.go` init function, not manually

## MongoDB — Query Standards
- Use the official `mongo-go-driver` (`go.mongodb.org/mongo-driver/v2`)
- Always set context with timeout for operations: `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)`
- Use `bson.M{}` for simple filters, `bson.D{}` when order matters
- Use `options.Find().SetLimit().SetSkip()` for pagination
- Never store sensitive data unencrypted in MongoDB

## MCP Server Integration
Before writing persistence code, use the available MCP servers to verify state:
- **`postgres` MCP**: Query schemas, verify migrations, test queries before embedding in Go
- **`mongodb` MCP**: Inspect collections, verify document schemas and indexes (read-only)
- **`redis` MCP**: Check cached data, inspect TTLs, verify blocklist entries

## Choosing Which Database
| If the data... | Use |
|:---|:---|
| Has relationships / needs JOINs / needs ACID transactions | PostgreSQL |
| Is append-only / audit trail | MongoDB |
| Has flexible/evolving schema | MongoDB |
| Needs full-text search on documents | MongoDB |
| Is a denormalized read model (CQRS) | MongoDB |
| Needs strong consistency + foreign key enforcement | PostgreSQL |
