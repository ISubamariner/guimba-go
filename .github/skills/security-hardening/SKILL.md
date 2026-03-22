---
name: security-hardening
description: "Enforces security best practices: CORS, CSP, rate limiting, input sanitization, secure headers, and OWASP Top 10 prevention. Use when user says 'add CORS', 'secure headers', 'rate limit', 'CSP', 'security', 'harden', 'OWASP', 'XSS', 'CSRF', 'injection', or when working with middleware/."
---

# Security Hardening

Enforces security best practices across the entire stack.

## CORS Configuration (Chi)

```go
cors.Handler(cors.Options{
    AllowedOrigins:   []string{"http://localhost:3000"}, // frontend origin
    AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
    ExposedHeaders:   []string{"X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           300, // 5 minutes
})
```

**Rules**: Never use `AllowedOrigins: ["*"]` with `AllowCredentials: true`.

## Security Headers Middleware

```go
func SecureHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "0") // modern browsers use CSP instead
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        if r.TLS != nil {
            w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        next.ServeHTTP(w, r)
    })
}
```

## Rate Limiting

Use per-IP rate limiting on auth endpoints (more aggressive) and general endpoints:

```go
// Auth endpoints: 5 requests/minute
authLimiter := httprate.NewRateLimiter(5, time.Minute)

// General endpoints: 100 requests/minute
generalLimiter := httprate.NewRateLimiter(100, time.Minute)
```

## Request Size Limits

```go
r.Use(middleware.RequestSize(1 << 20)) // 1 MB max body size
```

## Middleware Stack Order

```
1. SecureHeaders     ← sets security response headers
2. CORS              ← handles preflight and origin checks
3. RateLimiter       ← throttles abusive clients
4. RequestID         ← assigns unique request ID for tracing
5. Logger            ← logs request details
6. Recoverer         ← catches panics
7. AuthMiddleware    ← validates JWT (on protected routes)
8. RequireRole       ← checks RBAC (on restricted routes)
9. Handler           ← business logic
```

## Input Validation

- All input validated at the delivery layer using `go-playground/validator` struct tags
- Parameterized queries prevent SQL injection (`$1`, `$2` — never string concatenation)
- HTML output escaped by default in Next.js (React auto-escapes JSX)

## CSRF Protection

For cookie-based auth:
- Use `SameSite=Strict` on auth cookies
- Validate `Origin` header matches allowed origins
- Consider double-submit cookie pattern for extra protection

## Content Security Policy (Next.js)

```typescript
// next.config.ts
const cspHeader = `
  default-src 'self';
  script-src 'self' 'unsafe-eval' 'unsafe-inline';
  style-src 'self' 'unsafe-inline';
  img-src 'self' blob: data:;
  font-src 'self';
  connect-src 'self' ${process.env.NEXT_PUBLIC_API_URL};
  frame-ancestors 'none';
  base-uri 'self';
  form-action 'self';
`;
```

## Rules

- Never trust client input — always validate at the boundary
- Never expose stack traces or internal errors to the client
- Never store secrets in code or version control
- Always use HTTPS in production
- Always set security headers on every response
- Rate limit all auth endpoints aggressively
- Log all auth failures for monitoring
