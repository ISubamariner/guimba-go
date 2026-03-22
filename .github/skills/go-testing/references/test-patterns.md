# Go Testing Patterns Reference

## Test Helpers

### Setup/Teardown with TestMain
```go
func TestMain(m *testing.M) {
    // Setup
    code := m.Run()
    // Teardown
    os.Exit(code)
}
```

### Helper Functions
```go
func newTestService(t *testing.T) *ProgramService {
    t.Helper()
    repo := &mockProgramRepo{
        getByIDFn: func(ctx context.Context, id string) (*models.Program, error) {
            if id == "valid-id" {
                return &models.Program{ID: id, Name: "Test"}, nil
            }
            return nil, ErrNotFound
        },
    }
    return NewProgramService(repo)
}
```

## Assertion Patterns

### Using testify (if added)
```go
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.ErrorIs(t, err, ErrNotFound)
```

### Without testify (stdlib only)
```go
if got != want {
    t.Errorf("got %v, want %v", got, want)
}
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

## Integration Test Pattern
```go
//go:build integration

func TestProgramRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Use real DB connection from test container
}
```

## Benchmark Tests
```go
func BenchmarkGetProgram(b *testing.B) {
    svc := newTestService(b)
    for i := 0; i < b.N; i++ {
        svc.GetProgram(context.Background(), "valid-id")
    }
}
```
