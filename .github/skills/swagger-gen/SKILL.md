---
name: swagger-gen
description: "Generates Swagger/OpenAPI documentation from Go handler comments using swaggo/swag. Use when user says 'generate swagger', 'create API docs', 'update OpenAPI spec', 'add swagger annotations', or when working with handler/*.go files."
---

# Swagger Documentation Generator

Generates and maintains Swagger/OpenAPI docs from Go code annotations.

## Prerequisites
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Workflow

### Step 1: Add General API Info
In `backend/cmd/server/main.go`, add top-level annotations:
```go
// @title Guimba API
// @version 1.0
// @description Guimba Batangan Debt Management System API
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
```

### Step 2: Annotate Handlers
Add annotations above each handler function:
```go
// GetProgram godoc
// @Summary Get a social program by ID
// @Description Returns a single social program
// @Tags programs
// @Accept json
// @Produce json
// @Param id path string true "Program ID" format(uuid)
// @Success 200 {object} models.Program
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /programs/{id} [get]
func (h *ProgramHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
```

### Step 3: Generate Docs
```bash
cd backend
swag init -g cmd/server/main.go -o docs/
```
This creates `docs/swagger.json`, `docs/swagger.yaml`, and `docs/docs.go`.

### Step 4: Serve Swagger UI
Use `httpSwagger` middleware to serve the UI:
```go
import httpSwagger "github.com/swaggo/http-swagger"

r.Get("/swagger/*", httpSwagger.WrapHandler)
```
Access at: `http://localhost:8080/swagger/index.html`

## Common Annotations Reference
See `references/annotation-guide.md` for the full list.

## Troubleshooting

### "swag: command not found"
Ensure `$GOPATH/bin` is in PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Swagger UI Shows Stale Docs
Re-run `swag init` — the generated files are static and must be regenerated after changes.

### Annotation Parse Errors
Check that annotations are directly above the handler function with no blank lines between the comment block and `func`.
