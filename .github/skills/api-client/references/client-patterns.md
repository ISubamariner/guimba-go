# API Client Patterns Reference

## Full `api.ts` Template

```typescript
// src/lib/api.ts
import { ApiError } from './api-error';

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  accessToken = token;
}

export function getAccessToken(): string | null {
  return accessToken;
}

async function attemptTokenRefresh(): Promise<boolean> {
  try {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      credentials: 'include', // sends httpOnly cookie
    });
    if (!res.ok) return false;
    const data = await res.json();
    setAccessToken(data.access_token);
    return true;
  } catch {
    return false;
  }
}

function redirectToLogin() {
  if (typeof window !== 'undefined') {
    window.location.href = '/login';
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${BASE_URL}${endpoint}`;
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string>),
  };

  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (response.status === 401) {
    const refreshed = await attemptTokenRefresh();
    if (refreshed) {
      return request<T>(endpoint, options);
    }
    redirectToLogin();
    throw new ApiError('UNAUTHORIZED', 'Session expired');
  }

  if (!response.ok) {
    const errorBody = await response.json().catch(() => null);
    throw ApiError.fromResponse(response.status, errorBody);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

// Typed helpers
export const apiGet = <T>(endpoint: string) =>
  request<T>(endpoint, { method: 'GET' });

export const apiPost = <T>(endpoint: string, data?: unknown) =>
  request<T>(endpoint, { method: 'POST', body: data ? JSON.stringify(data) : undefined });

export const apiPut = <T>(endpoint: string, data: unknown) =>
  request<T>(endpoint, { method: 'PUT', body: JSON.stringify(data) });

export const apiPatch = <T>(endpoint: string, data: unknown) =>
  request<T>(endpoint, { method: 'PATCH', body: JSON.stringify(data) });

export const apiDelete = <T>(endpoint: string) =>
  request<T>(endpoint, { method: 'DELETE' });
```

## Error Type Definitions

```typescript
// src/lib/api-error.ts
export class ApiError extends Error {
  code: string;
  details: unknown[];
  status?: number;

  constructor(code: string, message: string, details: unknown[] = []) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.details = details;
  }

  static fromResponse(status: number, body: any): ApiError {
    const err = body?.error;
    const apiError = new ApiError(
      err?.code || 'UNKNOWN_ERROR',
      err?.message || `Request failed with status ${status}`,
      err?.details || []
    );
    apiError.status = status;
    return apiError;
  }

  get isValidation(): boolean {
    return this.code === 'VALIDATION_ERROR';
  }

  get isNotFound(): boolean {
    return this.code === 'NOT_FOUND';
  }

  get isUnauthorized(): boolean {
    return this.code === 'UNAUTHORIZED';
  }

  get isForbidden(): boolean {
    return this.code === 'FORBIDDEN';
  }
}
```

## Example Typed API Function

```typescript
// src/lib/api/debts.ts
import { apiGet, apiPost, apiPut, apiDelete } from '../api';
import type { DebtDTO, CreateDebtRequest, UpdateDebtRequest } from '@/types/debt';
import type { PaginatedResponse } from '@/types/pagination';

export const debtApi = {
  list: (page = 1, perPage = 20) =>
    apiGet<PaginatedResponse<DebtDTO>>(`/api/v1/debts?page=${page}&per_page=${perPage}`),

  getById: (id: string) =>
    apiGet<DebtDTO>(`/api/v1/debts/${id}`),

  create: (data: CreateDebtRequest) =>
    apiPost<DebtDTO>('/api/v1/debts', data),

  update: (id: string, data: UpdateDebtRequest) =>
    apiPut<DebtDTO>(`/api/v1/debts/${id}`, data),

  delete: (id: string) =>
    apiDelete<void>(`/api/v1/debts/${id}`),
};
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

export interface PaginationParams {
  page?: number;
  per_page?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export function buildQueryString(params: Record<string, string | number | undefined>): string {
  const entries = Object.entries(params).filter(([, v]) => v !== undefined);
  if (entries.length === 0) return '';
  return '?' + entries.map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&');
}
```
