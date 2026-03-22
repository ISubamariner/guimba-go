---
name: docker-compose-services
description: "Manages Docker Compose services for local development (PostgreSQL, Redis). Use when user says 'start services', 'docker compose up', 'reset database', 'check container status', 'start postgres', 'start redis', or 'local dev environment'."
---

# Docker Compose Services

Manages the local development environment via Docker Compose.

## Services

| Service | Image | Port | Purpose |
|:---|:---|:---|:---|
| PostgreSQL | `postgres:16-alpine` | `5432` | Primary relational database |
| MongoDB | `mongo:7` | `27017` | Document store, audit logs, CQRS reads |
| Redis | `redis:7-alpine` | `6379` | Caching layer |

## Common Operations

### Start All Services
```bash
docker compose up -d
```

### Stop All Services
```bash
docker compose down
```

### Reset Database (destroy data)
```bash
docker compose down -v
docker compose up -d
```

### Check Status
```bash
docker compose ps
```

### View Logs
```bash
docker compose logs -f postgres
docker compose logs -f redis
```

### Connect to PostgreSQL
```bash
docker compose exec postgres psql -U spmis -d spmis_db
```

### Connect to MongoDB
```bash
docker compose exec mongodb mongosh -u spmis -p spmis_password spmis_db
```

## Troubleshooting

### Port Already in Use
**Symptom**: `bind: address already in use`
**Fix**: Stop the conflicting service or change the port mapping in `docker-compose.yml`

### Container Won't Start
**Symptom**: Container exits immediately
**Fix**: Check logs with `docker compose logs <service>` — usually a config or permission issue

### Database Connection Refused
**Symptom**: Go app can't connect to PostgreSQL or MongoDB
**Fix**: Ensure containers are running (`docker compose ps`), check `DATABASE_URL` / `MONGODB_URI` in `.env` matches the compose config
