# Security Checklist

## OWASP Top 10 — Project-Specific Mitigations

| # | OWASP Risk | Guimba-GO Mitigation |
|:---|:---|:---|
| A01 | Broken Access Control | RBAC middleware (`RequireRole`), JWT auth on all protected routes |
| A02 | Cryptographic Failures | bcrypt for passwords, JWT with strong secret, HTTPS enforced |
| A03 | Injection | Parameterized queries (pgx `$1`), `go-playground/validator` on all DTOs |
| A04 | Insecure Design | Clean Architecture with domain validation, input never trusted |
| A05 | Security Misconfiguration | Secure headers middleware, CORS locked to known origins, no debug in prod |
| A06 | Vulnerable Components | Dependabot enabled, `go mod tidy`, `npm audit` in CI |
| A07 | Auth Failures | Rate limiting on auth endpoints, token expiry, refresh rotation |
| A08 | Data Integrity Failures | Signed JWTs, migration-only DDL, no dynamic SQL |
| A09 | Logging & Monitoring | `slog` structured logging, audit logs in MongoDB, request IDs |
| A10 | SSRF | No user-controlled URLs in backend fetch; validate all external calls |

## Middleware Stack Order

```go
r := chi.NewRouter()

// 1. Security headers (always first — sets response headers)
r.Use(middleware.SecureHeaders)

// 2. CORS (must be before auth to handle preflight)
r.Use(cors.Handler(corsOptions))

// 3. Rate limiting (before any processing)
r.Use(httprate.LimitByIP(100, time.Minute))

// 4. Request ID (for tracing)
r.Use(chimiddleware.RequestID)

// 5. Structured logging
r.Use(chimiddleware.Logger)

// 6. Panic recovery
r.Use(chimiddleware.Recoverer)

// 7. Request size limit
r.Use(chimiddleware.RequestSize(1 << 20))

// Protected routes add:
// 8. Auth middleware
// 9. Role middleware
```

## Content Security Policy Template for Next.js

```typescript
// next.config.ts
const securityHeaders = [
  {
    key: 'Content-Security-Policy',
    value: [
      "default-src 'self'",
      "script-src 'self' 'unsafe-eval' 'unsafe-inline'",
      "style-src 'self' 'unsafe-inline'",
      "img-src 'self' blob: data:",
      "font-src 'self'",
      `connect-src 'self' ${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}`,
      "frame-ancestors 'none'",
      "base-uri 'self'",
      "form-action 'self'",
    ].join('; '),
  },
  { key: 'X-Frame-Options', value: 'DENY' },
  { key: 'X-Content-Type-Options', value: 'nosniff' },
  { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
  { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
];

module.exports = {
  async headers() {
    return [{ source: '/(.*)', headers: securityHeaders }];
  },
};
```

## Pre-Deployment Checklist

- [ ] All routes require auth unless explicitly public
- [ ] CORS allows only known frontend origins
- [ ] Security headers middleware is first in chain
- [ ] Rate limiting enabled on auth endpoints
- [ ] No secrets in code or config files committed
- [ ] `JWT_SECRET` is strong (32+ chars) and rotated
- [ ] Passwords hashed with bcrypt (cost ≥ 10)
- [ ] No `console.log` or `fmt.Println` with sensitive data
- [ ] Docker images run as non-root user
- [ ] Dependencies scanned (`go mod tidy`, `npm audit`)
