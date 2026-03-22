# API Documentation

API documentation is auto-generated from Go handler annotations using **swaggo/swag**.

## Accessing Swagger UI

When the backend is running:
```
http://localhost:8080/swagger/index.html
```

## Generated Files

After running `swag init`, these files are created in `backend/docs/`:
- `swagger.json` — OpenAPI 3.0 spec (JSON)
- `swagger.yaml` — OpenAPI 3.0 spec (YAML)
- `docs.go` — Go file for embedding

## Regenerating Docs

```bash
cd backend
swag init -g cmd/server/main.go -o docs/
```

Run this after adding or modifying any handler annotations.

## Base URL

All API endpoints are versioned under:
```
http://localhost:8080/api/v1/
```

## Authentication

Protected endpoints require a Bearer token in the `Authorization` header:
```
Authorization: Bearer <jwt-token>
```

## Error Response Format

All errors follow this structure:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": []
  }
}
```

See the Swagger UI for the full endpoint catalog with request/response schemas.
