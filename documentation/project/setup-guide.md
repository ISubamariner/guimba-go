# Development Setup Guide

## Prerequisites

| Tool | Required Version | Installation |
|:---|:---|:---|
| Go | 1.24+ | https://go.dev/dl/ |
| Node.js | 24+ | Already installed (v24.11.0) |
| Docker | 28+ | Already installed (v28.0.4) |
| Git | 2.50+ | Already installed (v2.52.0) |

## Quick Start

### 1. Clone & Enter Project
```bash
git clone <repo-url>
cd Guimba-GO
```

### 2. Start Infrastructure
```bash
docker compose up -d
```
This starts PostgreSQL (port 5432), MongoDB (port 27017), and Redis (port 6379).

### 3. Set Up Environment
```bash
cp .env.example .env
# Edit .env with your local values
```

### 4. Run Backend
```bash
cd backend
go mod download
go run cmd/server/main.go
```
Backend runs on `http://localhost:8080`

### 5. Run Frontend
```bash
cd frontend
npm install
npm run dev
```
Frontend runs on `http://localhost:3000`

### 6. Run Migrations
```bash
migrate -path backend/migrations -database "$DATABASE_URL" up
```

### 7. Run Tests
```bash
# Go unit tests
cd backend && go test ./tests/unit/...

# Go integration tests
go test -tags=integration ./tests/integration/...

# Playwright E2E
cd tests/playwright && npx playwright test

# Frontend unit tests
cd frontend && npm test
```

## Useful Commands

| Command | What It Does |
|:---|:---|
| `docker compose ps` | Check running services |
| `docker compose exec mongodb mongosh` | Connect to MongoDB shell |
| `docker compose logs -f postgres` | Stream PostgreSQL logs |
| `docker compose down -v` | Stop services + destroy data |
| `swag init -g cmd/server/main.go -o docs/` | Regenerate Swagger docs |
| `npx playwright test --ui` | Playwright interactive mode |
