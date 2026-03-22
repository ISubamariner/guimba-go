---
name: error-handling
description: "Standardizes error handling across all Go layers using pkg/apperror/. Use when user says 'handle error', 'error response', 'error code', 'custom error', 'error mapping', 'status code', or when working with pkg/apperror/, handler files, or use case files."
---

# Standardized Error Handling

Manages error handling patterns across all Go layers using `pkg/apperror/`.

## Error Code Taxonomy

| Code | HTTP Status | Meaning |
|:---|:---|:---|
| `VALIDATION_ERROR` | 400 | Input failed validation rules |
| `UNAUTHORIZED` | 401 | Missing or invalid auth credentials |
| `FORBIDDEN` | 403 | Authenticated but lacking permissions |
| `NOT_FOUND` | 404 | Requested resource does not exist |
| `CONFLICT` | 409 | Resource already exists or state conflict |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

## AppError Struct (`pkg/apperror/`)

```go
// pkg/apperror/error.go
type AppError struct {
    Code    string   `json:"code"`
    Message string   `json:"message"`
    Details []string `json:"details,omitempty"`
    Err     error    `json:"-"` // internal error, never exposed
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}
```

## Constructor Functions

```go
func NewValidation(message string, details ...string) *AppError {
    return &AppError{Code: "VALIDATION_ERROR", Message: message, Details: details}
}

func NewUnauthorized(message string) *AppError {
    return &AppError{Code: "UNAUTHORIZED", Message: message}
}

func NewForbidden(message string) *AppError {
    return &AppError{Code: "FORBIDDEN", Message: message}
}

func NewNotFound(message string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: message}
}

func NewConflict(message string) *AppError {
    return &AppError{Code: "CONFLICT", Message: message}
}

func NewInternal(message string, err error) *AppError {
    return &AppError{Code: "INTERNAL_ERROR", Message: message, Err: err}
}
```

## Error Propagation Flow

```
Infrastructure (DB error)
  → wraps into AppError (NewNotFound, NewInternal, etc.)
  → Use case receives AppError, may add context
  → Handler checks AppError.Code → maps to HTTP status
  → Returns structured JSON response
```

### Infrastructure Layer
```go
func (r *DebtRepoPg) GetByID(ctx context.Context, id string) (*domain.Debt, error) {
    var debt domain.Debt
    err := r.db.QueryRow(ctx, query, id).Scan(...)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperror.NewNotFound("debt not found")
        }
        return nil, apperror.NewInternal("failed to query debt", err)
    }
    return &debt, nil
}
```

### Use Case Layer
```go
func (uc *GetDebtUseCase) Execute(ctx context.Context, id string) (*domain.Debt, error) {
    debt, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err // pass through — already an AppError
    }
    return debt, nil
}
```

### Handler Layer
```go
func (h *DebtHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    debt, err := h.getDebtUC.Execute(r.Context(), id)
    if err != nil {
        respondError(w, err)
        return
    }
    respondJSON(w, http.StatusOK, toDebtDTO(debt))
}
```

## Handler Helper: `respondError`

```go
func respondError(w http.ResponseWriter, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        status := mapCodeToStatus(appErr.Code)
        // Log full error server-side
        slog.Error("request error",
            "code", appErr.Code,
            "message", appErr.Message,
            "internal", appErr.Err,
        )
        respondJSON(w, status, map[string]any{
            "error": map[string]any{
                "code":    appErr.Code,
                "message": appErr.Message,
                "details": appErr.Details,
            },
        })
        return
    }
    // Unknown error — treat as internal
    slog.Error("unexpected error", "error", err)
    respondJSON(w, http.StatusInternalServerError, map[string]any{
        "error": map[string]any{
            "code":    "INTERNAL_ERROR",
            "message": "An unexpected error occurred",
        },
    })
}

func mapCodeToStatus(code string) int {
    switch code {
    case "VALIDATION_ERROR":
        return http.StatusBadRequest
    case "UNAUTHORIZED":
        return http.StatusUnauthorized
    case "FORBIDDEN":
        return http.StatusForbidden
    case "NOT_FOUND":
        return http.StatusNotFound
    case "CONFLICT":
        return http.StatusConflict
    default:
        return http.StatusInternalServerError
    }
}
```

## Rules

- **Use cases** return domain-level errors (never HTTP concepts like status codes)
- **Handlers** are the only layer that maps errors to HTTP responses
- **Infrastructure** wraps DB/external errors into AppErrors before returning
- Never expose internal error details to the client
- Always log the full error (including wrapped cause) server-side
- Validation errors should include a `details` array listing each invalid field
