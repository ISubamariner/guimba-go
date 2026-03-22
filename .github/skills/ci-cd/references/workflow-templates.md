# CI/CD Workflow Templates

## Go CI Workflow

```yaml
# .github/workflows/ci-backend.yml
name: Backend CI

on:
  push:
    branches: [develop, main]
    paths: ['backend/**']
  pull_request:
    branches: [main]
    paths: ['backend/**']

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: backend
          version: latest

  test:
    runs-on: ubuntu-latest
    needs: lint
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: guimba
          POSTGRES_PASSWORD: guimba_secret
          POSTGRES_DB: guimba_db
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('backend/go.sum') }}
      - name: Run tests
        working-directory: backend
        run: go test -race -coverprofile=coverage.out ./...
        env:
          POSTGRES_DSN: postgres://guimba:guimba_secret@localhost:5432/guimba_db?sslmode=disable
      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage
          path: backend/coverage.out

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - name: Build
        working-directory: backend
        run: go build -o server cmd/server/main.go
```

## Frontend CI Workflow

```yaml
# .github/workflows/ci-frontend.yml
name: Frontend CI

on:
  push:
    branches: [develop, main]
    paths: ['frontend/**']
  pull_request:
    branches: [main]
    paths: ['frontend/**']

jobs:
  lint-and-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '24'
      - uses: actions/cache@v4
        with:
          path: frontend/node_modules
          key: ${{ runner.os }}-npm-${{ hashFiles('frontend/package-lock.json') }}
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      - name: Lint
        working-directory: frontend
        run: npm run lint
      - name: Type check
        working-directory: frontend
        run: npx tsc --noEmit
      - name: Test
        working-directory: frontend
        run: npm test
      - name: Build
        working-directory: frontend
        run: npm run build
```

## Docker Build & Push Workflow

```yaml
# .github/workflows/docker-build.yml
name: Docker Build

on:
  push:
    branches: [main]

jobs:
  build-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: true
          tags: ghcr.io/${{ github.repository }}/backend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  build-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: true
          tags: ghcr.io/${{ github.repository }}/frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

## Migration Workflow

```yaml
# .github/workflows/migrate.yml
name: Run Migrations

on:
  workflow_dispatch:
    inputs:
      direction:
        description: 'Migration direction'
        required: true
        default: 'up'
        type: choice
        options: [up, down]
      steps:
        description: 'Number of steps (down only)'
        required: false
        type: number

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install migrate
        run: |
          curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
          sudo mv migrate /usr/local/bin/
      - name: Run migration
        run: |
          if [ "${{ inputs.direction }}" = "up" ]; then
            migrate -path backend/migrations -database "${{ secrets.DATABASE_URL }}" up
          else
            migrate -path backend/migrations -database "${{ secrets.DATABASE_URL }}" down ${{ inputs.steps }}
          fi
```
