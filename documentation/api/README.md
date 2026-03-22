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

## Available Endpoints

### System
| Method | Path | Description |
|:---|:---|:---|
| `GET` | `/health` | Health check (Postgres, MongoDB, Redis) |

### Programs (`/api/v1/programs`)
| Method | Path | Auth | Description |
|:---|:---|:---|:---|
| `GET` | `/api/v1/programs` | Public | List programs (paginated, filterable by status/search) |
| `POST` | `/api/v1/programs` | Staff/Admin | Create a new program |
| `GET` | `/api/v1/programs/{id}` | Public | Get a program by ID |
| `PUT` | `/api/v1/programs/{id}` | Staff/Admin | Update a program |
| `DELETE` | `/api/v1/programs/{id}` | Staff/Admin | Soft-delete a program |

### Query Parameters (List Programs)
| Param | Type | Description |
|:---|:---|:---|
| `status` | string | Filter: `active`, `inactive`, `closed` |
| `search` | string | Search by name (case-insensitive) |
| `limit` | int | Page size (default 20, max 100) |
| `offset` | int | Offset (default 0) |

### Auth (`/api/v1/auth`)
| Method | Path | Auth | Description |
|:---|:---|:---|:---|
| `POST` | `/api/v1/auth/register` | Public | Register a new user (returns JWT pair) |
| `POST` | `/api/v1/auth/login` | Public | Login with email + password (returns JWT pair) |
| `POST` | `/api/v1/auth/refresh` | Public | Refresh token pair (rotates, blocklists old) |
| `GET` | `/api/v1/auth/me` | Authenticated | Get current user profile with roles |
| `POST` | `/api/v1/auth/logout` | Authenticated | Logout (blocklists current token) |

### Users (`/api/v1/users`) — Admin Only
| Method | Path | Auth | Description |
|:---|:---|:---|:---|
| `GET` | `/api/v1/users` | Admin | List all users (paginated) |
| `PUT` | `/api/v1/users/{id}` | Admin | Update a user |
| `DELETE` | `/api/v1/users/{id}` | Admin | Soft-delete a user |
| `POST` | `/api/v1/users/{id}/roles` | Admin | Assign a role to a user |

### Beneficiaries (`/api/v1/beneficiaries`)
| Method | Path | Auth | Description |
|:---|:---|:---|:---|
| `GET` | `/api/v1/beneficiaries` | Authenticated | List beneficiaries (paginated, filterable by status/program/search) |
| `POST` | `/api/v1/beneficiaries` | Staff/Admin | Create a new beneficiary |
| `GET` | `/api/v1/beneficiaries/{id}` | Authenticated | Get a beneficiary by ID (includes program enrollments) |
| `PUT` | `/api/v1/beneficiaries/{id}` | Staff/Admin | Update a beneficiary |
| `DELETE` | `/api/v1/beneficiaries/{id}` | Staff/Admin | Soft-delete a beneficiary |
| `POST` | `/api/v1/beneficiaries/{id}/programs` | Staff/Admin | Enroll beneficiary in a program |
| `DELETE` | `/api/v1/beneficiaries/{id}/programs/{programId}` | Staff/Admin | Remove beneficiary from a program |

### Query Parameters (List Beneficiaries)
| Param | Type | Description |
|:---|:---|:---|
| `status` | string | Filter: `active`, `inactive`, `suspended` |
| `program_id` | string (UUID) | Filter by program enrollment |
| `search` | string | Search by name (case-insensitive) |
| `limit` | int | Page size (default 20, max 100) |
| `offset` | int | Offset (default 0) |
