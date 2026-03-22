---
name: ci-cd
description: "Manages CI/CD pipelines via GitHub Actions for linting, testing, building, and deploying. Use when user says 'add CI', 'create pipeline', 'GitHub Actions', 'automate tests', 'deploy', 'build pipeline', 'continuous integration', or when working with .github/workflows/."
---

# CI/CD Pipeline Management

Manages GitHub Actions workflows for linting, testing, building, and deploying.

## Pipeline Architecture

```
Push/PR → Lint → Test → Build → Deploy
```

### Stage Breakdown

| Stage | Go Backend | Next.js Frontend |
|:---|:---|:---|
| **Lint** | `golangci-lint run` | `npm run lint` |
| **Type Check** | (compile-time) | `npx tsc --noEmit` |
| **Test** | `go test ./...` | `npm test` |
| **Build** | `go build -o server cmd/server/main.go` | `npm run build` |
| **Docker** | Build & push backend image | Build & push frontend image |

## Environment Strategy

| Trigger | Environment | Purpose |
|:---|:---|:---|
| Push to `develop` | Development | Continuous validation |
| PR to `main` | Staging | Pre-merge validation |
| Merge to `main` | Production | Deploy to production |

## Caching Strategies

### Go Modules
```yaml
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('backend/go.sum') }}
    restore-keys: ${{ runner.os }}-go-
```

### npm
```yaml
- uses: actions/cache@v4
  with:
    path: frontend/node_modules
    key: ${{ runner.os }}-npm-${{ hashFiles('frontend/package-lock.json') }}
    restore-keys: ${{ runner.os }}-npm-
```

### Docker Layers
```yaml
- uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

## Secrets Management

Required GitHub Actions secrets:
- `DOCKER_USERNAME` / `DOCKER_PASSWORD` — for Docker Hub or GHCR
- `DATABASE_URL` — for migration runner (if applicable)
- `JWT_SECRET` — for integration tests needing auth

**Rules**:
- Never hardcode secrets in workflow files
- Use `${{ secrets.NAME }}` syntax
- Rotate secrets regularly
- Use environment-scoped secrets for production

## Workflow Files

All workflows live in `.github/workflows/`:
- `ci-backend.yml` — Go lint + test + build
- `ci-frontend.yml` — Frontend lint + type check + test + build
- `docker-build.yml` — Docker image build & push
- `migrate.yml` — Run pending database migrations

See `references/workflow-templates.md` for full templates.
