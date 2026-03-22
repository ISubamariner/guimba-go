---
name: go-testing
description: "Go testing patterns including table-driven tests, mocking, and coverage. Use when user says 'write tests', 'add test', 'test coverage', 'create mock', 'run tests', 'table-driven test', or when working with *_test.go files."
---

# Go Testing Skill

Provides patterns and workflows for writing and running Go tests.

## Before Debugging Any Test Failure
**MANDATORY**: Search `.github/skills/bug-tracker/references/bug-log.md` for related keywords before investigating a new test failure. If a matching root cause exists, apply the documented resolution first.

## Table-Driven Test Pattern
The standard pattern for all service/handler tests:

```go
func TestGetProgram(t *testing.T) {
    tests := []struct {
        name    string
        id      string
        want    *models.Program
        wantErr bool
    }{
        {
            name: "valid program",
            id:   "uuid-123",
            want: &models.Program{ID: "uuid-123", Name: "Test"},
        },
        {
            name:    "not found",
            id:      "nonexistent",
            wantErr: true,
        },
        {
            name:    "empty id",
            id:      "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := svc.GetProgram(context.Background(), tt.id)
            if (err != nil) != tt.wantErr {
                t.Errorf("GetProgram() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got.ID != tt.want.ID {
                t.Errorf("GetProgram() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Mocking with Interfaces
Define repository interfaces in `domain/repository/` so use cases can be tested with mocks from `backend/tests/mocks/`:

```go
// In domain/repository/program_repository.go
type ProgramRepository interface {
    GetByID(ctx context.Context, id string) (*entity.Program, error)
    List(ctx context.Context, limit, offset int) ([]entity.Program, error)
}

// In backend/tests/mocks/mock_program_repository.go
type MockProgramRepo struct {
    GetByIDFn func(ctx context.Context, id string) (*entity.Program, error)
}

func (m *MockProgramRepo) GetByID(ctx context.Context, id string) (*entity.Program, error) {
    return m.GetByIDFn(ctx, id)
}
```

## Running Tests

### Run All Backend Tests
```bash
cd backend && go test ./tests/...
```

### Run Only Unit Tests
```bash
go test ./tests/unit/...
```

### Run Only Integration Tests
```bash
go test -tags=integration ./tests/integration/...
```

### Run with Coverage
```bash
go test -coverprofile=coverage.out ./tests/... && go tool cover -html=coverage.out
```

### Run Specific Test
```bash
go test -run TestCreateProgram ./tests/unit/usecase/...
```

### Verbose Output
```bash
go test -v ./tests/...
```

## HTTP Handler Testing
```go
// In backend/tests/unit/delivery/program_handler_test.go
func TestGetProgramHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/v1/programs/uuid-123", nil)
    rctx := chi.NewRouteContext()
    rctx.URLParams.Add("id", "uuid-123")
    req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

    rr := httptest.NewRecorder()
    handler.GetProgram(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rr.Code)
    }
}
```

## Troubleshooting

### Tests Pass Individually But Fail Together
**Cause**: Shared mutable state between tests
**Fix**: Initialize fresh test data in each `t.Run()` block

### "context deadline exceeded" in Tests
**Cause**: Test hitting real database/network
**Fix**: Use mocks for external dependencies; ensure no real connections in unit tests
