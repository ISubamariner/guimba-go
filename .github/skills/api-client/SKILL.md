---
name: api-client
description: "Manages the centralized frontend API client (src/lib/api.ts), request/response typing, auth header injection, error handling, and token refresh. Use when user says 'API call', 'fetch data', 'api client', 'create api function', 'handle API error', 'token refresh', or when working with src/lib/api.ts or src/types/."
---

# Frontend API Client

Manages the centralized API client layer in `src/lib/api.ts`.

## Base Client Structure

All backend calls go through `src/lib/api.ts` — never call `fetch()` directly from components.

```typescript
// src/lib/api.ts
const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${BASE_URL}${endpoint}`;
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  // Inject auth token
  const token = getAccessToken();
  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, { ...options, headers });

  if (response.status === 401) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      return request<T>(endpoint, options); // Retry with new token
    }
    redirectToLogin();
    throw new ApiError('UNAUTHORIZED', 'Session expired');
  }

  if (!response.ok) {
    const errorBody = await response.json().catch(() => null);
    throw ApiError.fromResponse(response.status, errorBody);
  }

  return response.json();
}
```

## Typed Request Helpers

```typescript
export async function apiGet<T>(endpoint: string): Promise<T> {
  return request<T>(endpoint, { method: 'GET' });
}

export async function apiPost<T>(endpoint: string, data: unknown): Promise<T> {
  return request<T>(endpoint, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function apiPut<T>(endpoint: string, data: unknown): Promise<T> {
  return request<T>(endpoint, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function apiDelete<T>(endpoint: string): Promise<T> {
  return request<T>(endpoint, { method: 'DELETE' });
}
```

## Error Handling

### API Error Class
```typescript
// src/lib/api-error.ts
export class ApiError extends Error {
  code: string;
  details: unknown[];

  constructor(code: string, message: string, details: unknown[] = []) {
    super(message);
    this.code = code;
    this.details = details;
  }

  static fromResponse(status: number, body: any): ApiError {
    if (body?.error) {
      return new ApiError(body.error.code, body.error.message, body.error.details || []);
    }
    return new ApiError('UNKNOWN_ERROR', `Request failed with status ${status}`);
  }
}
```

### Error mapping from backend format
Backend returns:
```json
{ "error": { "code": "VALIDATION_ERROR", "message": "...", "details": [] } }
```

Frontend `ApiError` mirrors this structure for consistent handling.

## Type Definitions

Types in `src/types/` must match backend DTOs:

```typescript
// src/types/auth.ts
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  user: UserDTO;
}

export interface UserDTO {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  roles: string[];
}
```

## Pagination Response Type

```typescript
// src/types/pagination.ts
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}
```

## Example Typed API Function

```typescript
// src/lib/api.ts
import type { PaginatedResponse } from '@/types/pagination';
import type { DebtDTO } from '@/types/debt';

export async function getDebts(page = 1, perPage = 20): Promise<PaginatedResponse<DebtDTO>> {
  return apiGet(`/api/v1/debts?page=${page}&per_page=${perPage}`);
}

export async function getDebt(id: string): Promise<DebtDTO> {
  return apiGet(`/api/v1/debts/${id}`);
}

export async function createDebt(data: CreateDebtRequest): Promise<DebtDTO> {
  return apiPost('/api/v1/debts', data);
}
```

## Environment Configuration

```env
# .env.local (frontend)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

All API functions automatically use this base URL. In production, set it to the deployed backend URL.

## Rules

- Never call `fetch()` directly from components — always go through `src/lib/api.ts`
- Always type API responses with interfaces from `src/types/`
- Handle 401 globally in the base `request()` function
- Log API errors in development, suppress in production
- Never expose tokens in URL parameters
