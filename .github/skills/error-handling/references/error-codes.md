# Error Codes Reference

## Complete Error Code Table

| Code | HTTP Status | When To Use | Example Message |
|:---|:---|:---|:---|
| `VALIDATION_ERROR` | 400 | Input fails struct tag validation or business rule | `"email is required"` |
| `BAD_REQUEST` | 400 | Malformed request (bad JSON, invalid UUID) | `"Invalid JSON request body"` |
| `UNAUTHORIZED` | 401 | No token, expired token, invalid token | `"invalid or expired token"` |
| `FORBIDDEN` | 403 | Valid token but insufficient role | `"insufficient permissions"` |
| `NOT_FOUND` | 404 | Entity doesn't exist in DB | `"debt not found"` |
| `CONFLICT` | 409 | Duplicate unique field, state mismatch | `"email already registered"` |
| `INTERNAL_ERROR` | 500 | Unexpected failures (DB down, nil pointer, etc.) | `"an unexpected error occurred"` |

## Example Error Responses

### Validation Error (400)
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      "email is required",
      "amount must be greater than 0"
    ]
  }
}
```

### Unauthorized (401)
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token"
  }
}
```

### Forbidden (403)
```json
{
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions"
  }
}
```

### Not Found (404)
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Debt not found"
  }
}
```

### Conflict (409)
```json
{
  "error": {
    "code": "CONFLICT",
    "message": "Email already registered"
  }
}
```

### Internal Error (500)
```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "An unexpected error occurred"
  }
}
```

## Handler Helper Function Template

```go
package handler

import (
    "encoding/json"
    "errors"
    "log/slog"
    "net/http"

    "github.com/ISubamariner/guimba-go/backend/pkg/apperror"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        status := mapCodeToStatus(appErr.Code)
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

## Validation Error Details from go-playground/validator

```go
func mapValidationErrors(err error) *apperror.AppError {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        details := make([]string, len(ve))
        for i, fe := range ve {
            details[i] = formatFieldError(fe)
        }
        return apperror.NewValidation("Request validation failed", details...)
    }
    return apperror.NewValidation(err.Error())
}

func formatFieldError(fe validator.FieldError) string {
    field := strings.ToLower(fe.Field())
    switch fe.Tag() {
    case "required":
        return field + " is required"
    case "email":
        return field + " must be a valid email"
    case "min":
        return field + " must be at least " + fe.Param()
    case "max":
        return field + " must be at most " + fe.Param()
    case "gt":
        return field + " must be greater than " + fe.Param()
    default:
        return field + " is invalid"
    }
}
```
