# Phase 5: Frontend Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the frontend foundation — TypeScript types, API client, design system, UI primitives, auth flow, app shell, and dashboard — so feature pages can be built rapidly on a solid base.

**Architecture:** Next.js 16 App Router with React 19. All API calls go through a typed fetch wrapper in `src/lib/api.ts` that handles token injection/refresh. Auth state lives in a React context backed by localStorage. The app shell uses a sidebar layout with role-based navigation. Pages are server-rendered shells with client-side data fetching via custom hooks.

**Tech Stack:** Next.js 16.2.0, React 19, TypeScript strict, Tailwind CSS v4 (CSS-based config), Geist font family.

**Backend API:** `http://localhost:8080/api/v1` — all DTOs documented in `backend/internal/delivery/http/dto/`.

---

## File Structure

```
frontend/src/
├── types/
│   └── api.ts                    # All TypeScript types matching backend DTOs
├── lib/
│   ├── api.ts                    # Fetch wrapper with auth, error handling
│   └── cn.ts                     # clsx + tailwind-merge utility
├── hooks/
│   ├── use-auth.ts               # Auth context hook
│   └── use-api.ts                # Data fetching hooks (useQuery-like)
├── components/
│   ├── ui/
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── select.tsx
│   │   ├── textarea.tsx
│   │   ├── card.tsx
│   │   ├── badge.tsx
│   │   ├── modal.tsx
│   │   ├── table.tsx
│   │   ├── toast.tsx
│   │   └── index.ts              # Barrel export
│   ├── auth-provider.tsx         # Auth context provider
│   ├── sidebar.tsx               # Navigation sidebar
│   └── app-shell.tsx             # Layout wrapper (sidebar + content)
├── app/
│   ├── globals.css               # Modified: add design tokens
│   ├── layout.tsx                # Modified: wrap with AuthProvider
│   ├── page.tsx                  # Modified: redirect to /dashboard or /login
│   ├── login/
│   │   └── page.tsx              # Login page
│   ├── register/
│   │   └── page.tsx              # Registration page
│   └── (app)/
│       ├── layout.tsx            # App shell layout (sidebar + auth guard)
│       └── dashboard/
│           └── page.tsx          # Dashboard page
```

---

### Task 1: TypeScript Types

**Files:**
- Create: `frontend/src/types/api.ts`

- [ ] **Step 1: Create all TypeScript types matching backend DTOs**

```typescript
// frontend/src/types/api.ts

// ── Error Response ──────────────────────────────────────────
export type ErrorCode =
  | "NOT_FOUND"
  | "VALIDATION_ERROR"
  | "UNAUTHORIZED"
  | "FORBIDDEN"
  | "CONFLICT"
  | "INTERNAL_ERROR"
  | "BAD_REQUEST";

export interface ApiError {
  error: {
    code: ErrorCode;
    message: string;
    details?: string[];
  };
}

// ── Pagination ──────────────────────────────────────────────
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// ── Auth ────────────────────────────────────────────────────
export interface RegisterRequest {
  email: string;
  full_name: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface AuthResponse {
  user: UserResponse;
  access_token: string;
  refresh_token: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

// ── Users ───────────────────────────────────────────────────
export interface UserResponse {
  id: string;
  email: string;
  full_name: string;
  is_active: boolean;
  is_email_verified: boolean;
  last_login_at?: string;
  roles: RoleResponse[];
  created_at: string;
  updated_at: string;
}

export interface UpdateUserRequest {
  full_name: string;
  is_active: boolean;
}

export interface AssignRoleRequest {
  role_id: string;
}

export interface RoleResponse {
  id: string;
  name: string;
  display_name: string;
  description: string;
  is_system_role: boolean;
  permissions?: PermissionResponse[];
}

export interface PermissionResponse {
  id: string;
  name: string;
  display_name: string;
  category: string;
}

// ── Programs ────────────────────────────────────────────────
export interface CreateProgramRequest {
  name: string;
  description: string;
  status: "active" | "inactive" | "closed";
  start_date?: string;
  end_date?: string;
}

export interface UpdateProgramRequest {
  name: string;
  description: string;
  status: "active" | "inactive" | "closed";
  start_date?: string;
  end_date?: string;
}

export interface ProgramResponse {
  id: string;
  name: string;
  description: string;
  status: string;
  start_date?: string;
  end_date?: string;
  created_at: string;
  updated_at: string;
}

// ── Beneficiaries ───────────────────────────────────────────
export interface CreateBeneficiaryRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: "active" | "inactive" | "suspended";
  notes?: string;
}

export interface UpdateBeneficiaryRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: "active" | "inactive" | "suspended";
  notes?: string;
}

export interface EnrollProgramRequest {
  program_id: string;
}

export interface BeneficiaryResponse {
  id: string;
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: string;
  date_of_birth?: string;
  status: string;
  notes?: string;
  programs?: ProgramEnrollmentResponse[];
  created_at: string;
  updated_at: string;
}

export interface ProgramEnrollmentResponse {
  program_id: string;
  program_name: string;
  enrolled_at: string;
  status: string;
}

// ── Address ─────────────────────────────────────────────────
export interface AddressDTO {
  street: string;
  city: string;
  state_or_region: string;
  postal_code?: string;
  country?: string;
}

// ── Tenants ─────────────────────────────────────────────────
export interface CreateTenantRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  notes?: string;
}

export interface UpdateTenantRequest {
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  notes?: string;
}

export interface TenantResponse {
  id: string;
  full_name: string;
  email?: string;
  phone_number?: string;
  national_id?: string;
  address?: AddressDTO;
  landlord_id: string;
  is_active: boolean;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ── Properties ──────────────────────────────────────────────
export interface CreatePropertyRequest {
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type?: string;
  size_in_acres?: number;
  size_in_sqm: number;
  monthly_rent_amount?: number;
  description?: string;
}

export interface UpdatePropertyRequest {
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type?: string;
  size_in_acres?: number;
  size_in_sqm: number;
  is_available_for_rent?: boolean;
  monthly_rent_amount?: number;
  description?: string;
}

export interface PropertyResponse {
  id: string;
  name: string;
  property_code: string;
  address?: AddressDTO;
  geojson_coordinates?: string;
  property_type: string;
  size_in_acres?: number;
  size_in_sqm: number;
  owner_id: string;
  is_available_for_rent: boolean;
  is_active: boolean;
  monthly_rent_amount?: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

// ── Money ───────────────────────────────────────────────────
export interface MoneyDTO {
  amount: string;
  currency: string;
}

// ── Debts ───────────────────────────────────────────────────
export type DebtStatus = "PENDING" | "PARTIAL" | "PAID" | "OVERDUE" | "CANCELLED";
export type DebtType = "RENT" | "UTILITIES" | "MAINTENANCE" | "PENALTY" | "OTHER";

export interface CreateDebtRequest {
  tenant_id: string;
  property_id?: string;
  debt_type: string;
  description: string;
  original_amount: MoneyDTO;
  due_date: string;
  notes?: string;
}

export interface UpdateDebtRequest {
  description: string;
  debt_type: string;
  due_date: string;
  property_id?: string;
  notes?: string;
}

export interface CancelDebtRequest {
  reason?: string;
}

export interface DebtResponse {
  id: string;
  tenant_id: string;
  landlord_id: string;
  property_id?: string;
  debt_type: string;
  description: string;
  original_amount: MoneyDTO;
  amount_paid: MoneyDTO;
  balance: MoneyDTO;
  due_date: string;
  status: DebtStatus;
  notes?: string;
  created_at: string;
  updated_at: string;
}

// ── Transactions ────────────────────────────────────────────
export interface RecordPaymentRequest {
  debt_id: string;
  tenant_id: string;
  amount: MoneyDTO;
  payment_method: string;
  transaction_date: string;
  description: string;
  receipt_number?: string;
  reference_number?: string;
}

export interface RecordRefundRequest {
  debt_id: string;
  tenant_id: string;
  amount: MoneyDTO;
  payment_method: string;
  refund_date: string;
  description: string;
  reference_number?: string;
}

export interface TransactionResponse {
  id: string;
  debt_id: string;
  tenant_id: string;
  landlord_id: string;
  recorded_by_user_id?: string;
  transaction_type: string;
  amount: MoneyDTO;
  payment_method: string;
  transaction_date: string;
  description: string;
  receipt_number?: string;
  reference_number?: string;
  is_verified: boolean;
  verified_by_user_id?: string;
  verified_at?: string;
  created_at: string;
  updated_at: string;
}

// ── Audit ───────────────────────────────────────────────────
export interface AuditEntryResponse {
  id: string;
  user_id: string;
  user_email: string;
  user_role: string;
  action: string;
  resource_type: string;
  resource_id: string;
  ip_address: string;
  endpoint: string;
  method: string;
  status_code: number;
  success: boolean;
  error_message?: string;
  metadata?: Record<string, unknown>;
  timestamp: string;
}

// ── Dashboard ───────────────────────────────────────────────
export interface DashboardStatsResponse {
  total_tenants: number;
  total_properties: number;
  active_debts: number;
  overdue_debts: number;
}

export interface RecentActivityResponse {
  action: string;
  description: string;
  timestamp: string;
}

export interface RecentActivitiesResponse {
  data: RecentActivityResponse[];
}

// ── Health ──────────────────────────────────────────────────
export interface HealthResponse {
  status: string;
  timestamp: string;
  services: Record<string, string>;
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/api.ts
git commit -m "feat(frontend): add TypeScript types matching all backend DTOs"
```

---

### Task 2: Utility Functions

**Files:**
- Create: `frontend/src/lib/cn.ts`

- [ ] **Step 1: Install clsx and tailwind-merge**

```bash
cd frontend && npm install clsx tailwind-merge
```

- [ ] **Step 2: Create cn utility**

```typescript
// frontend/src/lib/cn.ts
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/cn.ts frontend/package.json frontend/package-lock.json
git commit -m "feat(frontend): add cn utility (clsx + tailwind-merge)"
```

---

### Task 3: API Client

**Files:**
- Create: `frontend/src/lib/api.ts`

- [ ] **Step 1: Create the API client**

The API client wraps `fetch` with:
- Base URL configuration
- Auth token injection from localStorage
- Automatic token refresh on 401
- Typed error extraction
- JSON serialization

```typescript
// frontend/src/lib/api.ts
import type { ApiError, TokenResponse } from "@/types/api";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export class ApiClientError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: string[],
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

function getAccessToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

function getRefreshToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("refresh_token");
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem("access_token", access);
  localStorage.setItem("refresh_token", refresh);
}

export function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
}

async function tryRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) return false;

    const data: TokenResponse = await res.json();
    setTokens(data.access_token, data.refresh_token);
    return true;
  } catch {
    return false;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${API_BASE}${path}`;

  const headers: Record<string, string> = {
    ...(options.headers as Record<string, string>),
  };

  // Don't set Content-Type for FormData (browser sets multipart boundary)
  if (!(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }

  const token = getAccessToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res = await fetch(url, { ...options, headers });

  // Try refresh on 401
  if (res.status === 401 && token) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${getAccessToken()}`;
      res = await fetch(url, { ...options, headers });
    }
  }

  if (!res.ok) {
    let apiErr: ApiError | null = null;
    try {
      apiErr = await res.json();
    } catch {
      // Response wasn't JSON
    }

    throw new ApiClientError(
      res.status,
      apiErr?.error?.code || "UNKNOWN",
      apiErr?.error?.message || `Request failed with status ${res.status}`,
      apiErr?.error?.details,
    );
  }

  // 204 No Content
  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

export const api = {
  get: <T>(path: string) => request<T>(path),

  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "PUT",
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(path: string) =>
    request<T>(path, { method: "DELETE" }),
};
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/lib/api.ts
git commit -m "feat(frontend): add typed API client with auth token handling"
```

---

### Task 4: Auth Context

**Files:**
- Create: `frontend/src/components/auth-provider.tsx`
- Create: `frontend/src/hooks/use-auth.ts`

- [ ] **Step 1: Create auth context and provider**

```tsx
// frontend/src/components/auth-provider.tsx
"use client";

import {
  createContext,
  useCallback,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api, clearTokens, ApiClientError } from "@/lib/api";
import type {
  AuthResponse,
  LoginRequest,
  RegisterRequest,
  UserResponse,
} from "@/types/api";

export interface AuthContextValue {
  user: UserResponse | null;
  isLoading: boolean;
  login: (data: LoginRequest) => Promise<void>;
  register: (data: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  hasRole: (role: string) => boolean;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token) {
      setIsLoading(false);
      return;
    }

    api
      .get<UserResponse>("/auth/me")
      .then(setUser)
      .catch(() => {
        clearTokens();
      })
      .finally(() => setIsLoading(false));
  }, []);

  const login = useCallback(async (data: LoginRequest) => {
    const res = await api.post<AuthResponse>("/auth/login", data);
    localStorage.setItem("access_token", res.access_token);
    localStorage.setItem("refresh_token", res.refresh_token);
    setUser(res.user);
  }, []);

  const register = useCallback(async (data: RegisterRequest) => {
    const res = await api.post<AuthResponse>("/auth/register", data);
    localStorage.setItem("access_token", res.access_token);
    localStorage.setItem("refresh_token", res.refresh_token);
    setUser(res.user);
  }, []);

  const logout = useCallback(async () => {
    try {
      await api.post("/auth/logout");
    } catch (err) {
      // Ignore errors — we're logging out regardless
      if (!(err instanceof ApiClientError)) throw err;
    }
    clearTokens();
    setUser(null);
  }, []);

  const hasRole = useCallback(
    (role: string) => {
      if (!user) return false;
      return user.roles.some((r) => r.name === role);
    },
    [user],
  );

  const value = useMemo(
    () => ({ user, isLoading, login, register, logout, hasRole }),
    [user, isLoading, login, register, logout, hasRole],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
```

- [ ] **Step 2: Create useAuth hook**

```typescript
// frontend/src/hooks/use-auth.ts
"use client";

import { useContext } from "react";
import { AuthContext, type AuthContextValue } from "@/components/auth-provider";

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/auth-provider.tsx frontend/src/hooks/use-auth.ts
git commit -m "feat(frontend): add auth context provider and useAuth hook"
```

---

### Task 5: Design System Tokens + UI Primitives

**Files:**
- Modify: `frontend/src/app/globals.css`
- Create: `frontend/src/components/ui/button.tsx`
- Create: `frontend/src/components/ui/input.tsx`
- Create: `frontend/src/components/ui/select.tsx`
- Create: `frontend/src/components/ui/textarea.tsx`
- Create: `frontend/src/components/ui/card.tsx`
- Create: `frontend/src/components/ui/badge.tsx`
- Create: `frontend/src/components/ui/modal.tsx`
- Create: `frontend/src/components/ui/table.tsx`
- Create: `frontend/src/components/ui/toast.tsx`
- Create: `frontend/src/components/ui/index.ts`

- [ ] **Step 1: Update globals.css with design tokens**

Replace the contents of `frontend/src/app/globals.css` with:

```css
@import "tailwindcss";

:root {
  /* Brand colors */
  --color-primary: #2563eb;
  --color-primary-hover: #1d4ed8;
  --color-primary-light: #dbeafe;

  /* Semantic */
  --color-success: #16a34a;
  --color-success-light: #dcfce7;
  --color-warning: #d97706;
  --color-warning-light: #fef3c7;
  --color-danger: #dc2626;
  --color-danger-light: #fee2e2;

  /* Neutrals */
  --background: #ffffff;
  --foreground: #0f172a;
  --muted: #64748b;
  --muted-bg: #f1f5f9;
  --border: #e2e8f0;
  --ring: #2563eb;

  /* Radii */
  --radius-sm: 0.375rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
}

@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-muted: var(--muted);
  --color-muted-bg: var(--muted-bg);
  --color-border: var(--border);
  --color-ring: var(--ring);
  --color-primary: var(--color-primary);
  --color-primary-hover: var(--color-primary-hover);
  --color-primary-light: var(--color-primary-light);
  --color-success: var(--color-success);
  --color-success-light: var(--color-success-light);
  --color-warning: var(--color-warning);
  --color-warning-light: var(--color-warning-light);
  --color-danger: var(--color-danger);
  --color-danger-light: var(--color-danger-light);
  --font-sans: var(--font-geist-sans);
  --font-mono: var(--font-geist-mono);
}

@media (prefers-color-scheme: dark) {
  :root {
    --background: #0f172a;
    --foreground: #f1f5f9;
    --muted: #94a3b8;
    --muted-bg: #1e293b;
    --border: #334155;
    --ring: #3b82f6;
  }
}

body {
  background: var(--background);
  color: var(--foreground);
}
```

- [ ] **Step 2: Create Button component**

```tsx
// frontend/src/components/ui/button.tsx
import { forwardRef, type ButtonHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

type ButtonVariant = "primary" | "secondary" | "outline" | "ghost" | "danger";
type ButtonSize = "sm" | "md" | "lg";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "bg-primary text-white hover:bg-primary-hover focus-visible:ring-2 focus-visible:ring-ring",
  secondary:
    "bg-muted-bg text-foreground hover:bg-border focus-visible:ring-2 focus-visible:ring-ring",
  outline:
    "border border-border text-foreground hover:bg-muted-bg focus-visible:ring-2 focus-visible:ring-ring",
  ghost:
    "text-foreground hover:bg-muted-bg focus-visible:ring-2 focus-visible:ring-ring",
  danger:
    "bg-danger text-white hover:bg-danger/90 focus-visible:ring-2 focus-visible:ring-danger",
};

const sizeStyles: Record<ButtonSize, string> = {
  sm: "h-8 px-3 text-sm rounded-[var(--radius-sm)]",
  md: "h-10 px-4 text-sm rounded-[var(--radius-md)]",
  lg: "h-12 px-6 text-base rounded-[var(--radius-md)]",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center font-medium transition-colors focus-visible:outline-none disabled:pointer-events-none disabled:opacity-50",
          variantStyles[variant],
          sizeStyles[size],
          className,
        )}
        disabled={disabled}
        {...props}
      />
    );
  },
);
Button.displayName = "Button";
```

- [ ] **Step 3: Create Input component**

```tsx
// frontend/src/components/ui/input.tsx
import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, id, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-foreground">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn(
            "flex h-10 w-full rounded-[var(--radius-md)] border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted transition-colors focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
            error && "border-danger focus:ring-danger",
            className,
          )}
          {...props}
        />
        {error && <p className="text-sm text-danger">{error}</p>}
      </div>
    );
  },
);
Input.displayName = "Input";
```

- [ ] **Step 4: Create Select component**

```tsx
// frontend/src/components/ui/select.tsx
import { forwardRef, type SelectHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
  options: { value: string; label: string }[];
  placeholder?: string;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  ({ className, label, error, id, options, placeholder, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-foreground">
            {label}
          </label>
        )}
        <select
          ref={ref}
          id={id}
          className={cn(
            "flex h-10 w-full rounded-[var(--radius-md)] border border-border bg-background px-3 py-2 text-sm text-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
            error && "border-danger focus:ring-danger",
            className,
          )}
          {...props}
        >
          {placeholder && (
            <option value="" disabled>
              {placeholder}
            </option>
          )}
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        {error && <p className="text-sm text-danger">{error}</p>}
      </div>
    );
  },
);
Select.displayName = "Select";
```

- [ ] **Step 5: Create Textarea component**

```tsx
// frontend/src/components/ui/textarea.tsx
import { forwardRef, type TextareaHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, label, error, id, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-foreground">
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={id}
          className={cn(
            "flex min-h-[80px] w-full rounded-[var(--radius-md)] border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted transition-colors focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
            error && "border-danger focus:ring-danger",
            className,
          )}
          {...props}
        />
        {error && <p className="text-sm text-danger">{error}</p>}
      </div>
    );
  },
);
Textarea.displayName = "Textarea";
```

- [ ] **Step 6: Create Card component**

```tsx
// frontend/src/components/ui/card.tsx
import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/cn";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function Card({ className, children, ...props }: CardProps) {
  return (
    <div
      className={cn(
        "rounded-[var(--radius-lg)] border border-border bg-background p-6 shadow-sm",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function CardHeader({ className, children, ...props }: CardProps) {
  return (
    <div className={cn("mb-4", className)} {...props}>
      {children}
    </div>
  );
}

export function CardTitle({
  className,
  children,
  ...props
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3 className={cn("text-lg font-semibold text-foreground", className)} {...props}>
      {children}
    </h3>
  );
}

export function CardContent({ className, children, ...props }: CardProps) {
  return (
    <div className={cn(className)} {...props}>
      {children}
    </div>
  );
}
```

- [ ] **Step 7: Create Badge component**

```tsx
// frontend/src/components/ui/badge.tsx
import type { HTMLAttributes } from "react";
import { cn } from "@/lib/cn";

type BadgeVariant = "default" | "success" | "warning" | "danger" | "outline";

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

const variantStyles: Record<BadgeVariant, string> = {
  default: "bg-primary-light text-primary",
  success: "bg-success-light text-success",
  warning: "bg-warning-light text-warning",
  danger: "bg-danger-light text-danger",
  outline: "border border-border text-muted",
};

export function Badge({ className, variant = "default", ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
        variantStyles[variant],
        className,
      )}
      {...props}
    />
  );
}
```

- [ ] **Step 8: Create Modal component**

```tsx
// frontend/src/components/ui/modal.tsx
"use client";

import { useEffect, useRef, type ReactNode } from "react";
import { cn } from "@/lib/cn";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  className?: string;
}

export function Modal({ open, onClose, title, children, className }: ModalProps) {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (open) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [open]);

  return (
    <dialog
      ref={dialogRef}
      onClose={onClose}
      className={cn(
        "rounded-[var(--radius-lg)] border border-border bg-background p-0 shadow-lg backdrop:bg-black/50",
        "max-w-lg w-full",
        className,
      )}
    >
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">{title}</h2>
          <button
            onClick={onClose}
            className="text-muted hover:text-foreground transition-colors"
            aria-label="Close"
          >
            &#x2715;
          </button>
        </div>
        {children}
      </div>
    </dialog>
  );
}
```

- [ ] **Step 9: Create Table component**

```tsx
// frontend/src/components/ui/table.tsx
import type { HTMLAttributes, TdHTMLAttributes, ThHTMLAttributes } from "react";
import { cn } from "@/lib/cn";

export function Table({
  className,
  ...props
}: HTMLAttributes<HTMLTableElement>) {
  return (
    <div className="overflow-x-auto">
      <table
        className={cn("w-full text-sm text-left", className)}
        {...props}
      />
    </div>
  );
}

export function TableHeader({
  className,
  ...props
}: HTMLAttributes<HTMLTableSectionElement>) {
  return (
    <thead
      className={cn("border-b border-border bg-muted-bg", className)}
      {...props}
    />
  );
}

export function TableBody({
  className,
  ...props
}: HTMLAttributes<HTMLTableSectionElement>) {
  return <tbody className={cn(className)} {...props} />;
}

export function TableRow({
  className,
  ...props
}: HTMLAttributes<HTMLTableRowElement>) {
  return (
    <tr
      className={cn(
        "border-b border-border last:border-0 hover:bg-muted-bg/50 transition-colors",
        className,
      )}
      {...props}
    />
  );
}

export function TableHead({
  className,
  ...props
}: ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      className={cn(
        "px-4 py-3 text-xs font-medium text-muted uppercase tracking-wider",
        className,
      )}
      {...props}
    />
  );
}

export function TableCell({
  className,
  ...props
}: TdHTMLAttributes<HTMLTableCellElement>) {
  return (
    <td
      className={cn("px-4 py-3 text-foreground", className)}
      {...props}
    />
  );
}
```

- [ ] **Step 10: Create Toast component**

```tsx
// frontend/src/components/ui/toast.tsx
"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { cn } from "@/lib/cn";

type ToastType = "success" | "error" | "info";

interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

interface ToastContextValue {
  toast: (message: string, type?: ToastType) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}

let nextId = 0;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const toast = useCallback((message: string, type: ToastType = "info") => {
    const id = nextId++;
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  }, []);

  const value = useMemo(() => ({ toast }), [toast]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={cn(
              "px-4 py-3 rounded-[var(--radius-md)] shadow-lg text-sm text-white animate-[slideIn_0.2s_ease-out]",
              t.type === "success" && "bg-success",
              t.type === "error" && "bg-danger",
              t.type === "info" && "bg-primary",
            )}
          >
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
```

- [ ] **Step 11: Create barrel export**

```typescript
// frontend/src/components/ui/index.ts
export { Button } from "./button";
export { Input } from "./input";
export { Select } from "./select";
export { Textarea } from "./textarea";
export { Card, CardHeader, CardTitle, CardContent } from "./card";
export { Badge } from "./badge";
export { Modal } from "./modal";
export {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "./table";
export { ToastProvider, useToast } from "./toast";
```

- [ ] **Step 12: Commit**

```bash
git add frontend/src/app/globals.css frontend/src/components/ui/
git commit -m "feat(frontend): add design tokens and UI primitive components"
```

---

### Task 6: Login Page

**Files:**
- Create: `frontend/src/app/login/page.tsx`
- Modify: `frontend/src/app/layout.tsx`
- Modify: `frontend/src/app/page.tsx`

- [ ] **Step 1: Update root layout with providers**

Replace `frontend/src/app/layout.tsx` with:

```tsx
// frontend/src/app/layout.tsx
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/auth-provider";
import { ToastProvider } from "@/components/ui/toast";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Guimba-GO",
  description: "Municipal Social Protection Management System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <AuthProvider>
          <ToastProvider>{children}</ToastProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
```

- [ ] **Step 2: Update root page to redirect**

```tsx
// frontend/src/app/page.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";

export default function HomePage() {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;
    if (user) {
      router.replace("/dashboard");
    } else {
      router.replace("/login");
    }
  }, [user, isLoading, router]);

  return (
    <div className="flex h-screen items-center justify-center">
      <div className="text-muted">Loading...</div>
    </div>
  );
}
```

- [ ] **Step 3: Create login page**

```tsx
// frontend/src/app/login/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/hooks/use-auth";
import { Button, Input } from "@/components/ui";
import { ApiClientError } from "@/lib/api";

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await login({ email, password });
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground">Guimba-GO</h1>
          <p className="text-sm text-muted mt-1">
            Municipal Social Protection Management
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
              {error}
            </div>
          )}

          <Input
            id="email"
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />

          <Input
            id="password"
            label="Password"
            type="password"
            placeholder="Enter your password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="current-password"
          />

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "Signing in..." : "Sign in"}
          </Button>
        </form>

        <p className="text-center text-sm text-muted">
          Don&apos;t have an account?{" "}
          <Link href="/register" className="text-primary hover:underline">
            Register
          </Link>
        </p>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/layout.tsx frontend/src/app/page.tsx frontend/src/app/login/
git commit -m "feat(frontend): add login page with auth integration"
```

---

### Task 7: Register Page

**Files:**
- Create: `frontend/src/app/register/page.tsx`

- [ ] **Step 1: Create registration page**

```tsx
// frontend/src/app/register/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/hooks/use-auth";
import { Button, Input } from "@/components/ui";
import { ApiClientError } from "@/lib/api";

export default function RegisterPage() {
  const { register } = useAuth();
  const router = useRouter();
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    setLoading(true);

    try {
      await register({ email, full_name: fullName, password });
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground">Create Account</h1>
          <p className="text-sm text-muted mt-1">
            Register as a landlord or auditor
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
              {error}
            </div>
          )}

          <Input
            id="fullName"
            label="Full Name"
            type="text"
            placeholder="Juan Dela Cruz"
            value={fullName}
            onChange={(e) => setFullName(e.target.value)}
            required
          />

          <Input
            id="email"
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
          />

          <Input
            id="password"
            label="Password"
            type="password"
            placeholder="At least 8 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            minLength={8}
            autoComplete="new-password"
          />

          <Input
            id="confirmPassword"
            label="Confirm Password"
            type="password"
            placeholder="Re-enter your password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            autoComplete="new-password"
          />

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "Creating account..." : "Create account"}
          </Button>
        </form>

        <p className="text-center text-sm text-muted">
          Already have an account?{" "}
          <Link href="/login" className="text-primary hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/register/
git commit -m "feat(frontend): add registration page"
```

---

### Task 8: App Shell (Sidebar + Auth Guard)

**Files:**
- Create: `frontend/src/components/sidebar.tsx`
- Create: `frontend/src/components/app-shell.tsx`
- Create: `frontend/src/app/(app)/layout.tsx`

- [ ] **Step 1: Create sidebar component**

```tsx
// frontend/src/components/sidebar.tsx
"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { cn } from "@/lib/cn";

interface NavItem {
  label: string;
  href: string;
  roles?: string[];
}

const navItems: NavItem[] = [
  { label: "Dashboard", href: "/dashboard" },
  { label: "Tenants", href: "/tenants", roles: ["admin", "landlord"] },
  { label: "Properties", href: "/properties", roles: ["admin", "landlord"] },
  { label: "Debts", href: "/debts", roles: ["admin", "landlord"] },
  { label: "Transactions", href: "/transactions", roles: ["admin", "landlord"] },
  { label: "Beneficiaries", href: "/beneficiaries", roles: ["admin", "staff"] },
  { label: "Programs", href: "/programs" },
  { label: "Users", href: "/users", roles: ["admin"] },
  { label: "Audit Logs", href: "/audit", roles: ["admin", "auditor"] },
];

export function Sidebar() {
  const pathname = usePathname();
  const { user, logout, hasRole } = useAuth();

  const visibleItems = navItems.filter((item) => {
    if (!item.roles) return true;
    return item.roles.some((role) => hasRole(role));
  });

  return (
    <aside className="flex flex-col w-64 border-r border-border bg-background h-full">
      <div className="p-4 border-b border-border">
        <h1 className="text-lg font-bold text-foreground">Guimba-GO</h1>
      </div>

      <nav className="flex-1 overflow-y-auto p-2">
        {visibleItems.map((item) => {
          const isActive = pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                "flex items-center px-3 py-2 text-sm rounded-[var(--radius-md)] transition-colors mb-0.5",
                isActive
                  ? "bg-primary-light text-primary font-medium"
                  : "text-muted hover:bg-muted-bg hover:text-foreground",
              )}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-border p-4">
        <div className="text-sm text-foreground font-medium truncate">
          {user?.full_name}
        </div>
        <div className="text-xs text-muted truncate">{user?.email}</div>
        <button
          onClick={logout}
          className="mt-2 text-sm text-muted hover:text-danger transition-colors"
        >
          Sign out
        </button>
      </div>
    </aside>
  );
}
```

- [ ] **Step 2: Create app shell component**

```tsx
// frontend/src/components/app-shell.tsx
import type { ReactNode } from "react";
import { Sidebar } from "./sidebar";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-6">{children}</main>
    </div>
  );
}
```

- [ ] **Step 3: Create authenticated layout with auth guard**

```tsx
// frontend/src/app/(app)/layout.tsx
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/use-auth";
import { AppShell } from "@/components/app-shell";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace("/login");
    }
  }, [user, isLoading, router]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="text-muted">Loading...</div>
      </div>
    );
  }

  if (!user) return null;

  return <AppShell>{children}</AppShell>;
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/sidebar.tsx frontend/src/components/app-shell.tsx frontend/src/app/\(app\)/layout.tsx
git commit -m "feat(frontend): add app shell with sidebar navigation and auth guard"
```

---

### Task 9: Dashboard Page

**Files:**
- Create: `frontend/src/app/(app)/dashboard/page.tsx`

- [ ] **Step 1: Create dashboard page**

```tsx
// frontend/src/app/(app)/dashboard/page.tsx
"use client";

import { useCallback, useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, Badge } from "@/components/ui";
import { api } from "@/lib/api";
import type { DashboardStatsResponse, RecentActivitiesResponse } from "@/types/api";

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStatsResponse | null>(null);
  const [activities, setActivities] = useState<RecentActivitiesResponse | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchData = useCallback(async () => {
    try {
      const [statsRes, activitiesRes] = await Promise.all([
        api.get<DashboardStatsResponse>("/dashboard/stats"),
        api.get<RecentActivitiesResponse>("/dashboard/recent-activities"),
      ]);
      setStats(statsRes);
      setActivities(activitiesRes);
    } catch {
      // Errors handled by API client (401 → redirect)
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return <div className="text-muted">Loading dashboard...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total Tenants" value={stats?.total_tenants ?? 0} />
        <StatCard label="Total Properties" value={stats?.total_properties ?? 0} />
        <StatCard
          label="Active Debts"
          value={stats?.active_debts ?? 0}
          variant="warning"
        />
        <StatCard
          label="Overdue Debts"
          value={stats?.overdue_debts ?? 0}
          variant="danger"
        />
      </div>

      {/* Recent Activities */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Activities</CardTitle>
        </CardHeader>
        <CardContent>
          {!activities?.data?.length ? (
            <p className="text-sm text-muted">No recent activities</p>
          ) : (
            <div className="space-y-3">
              {activities.data.map((activity, i) => (
                <div
                  key={i}
                  className="flex items-start justify-between gap-4 py-2 border-b border-border last:border-0"
                >
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-foreground">{activity.description}</p>
                    <p className="text-xs text-muted mt-0.5">
                      {new Date(activity.timestamp).toLocaleString()}
                    </p>
                  </div>
                  <Badge variant="outline">{activity.action}</Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({
  label,
  value,
  variant,
}: {
  label: string;
  value: number;
  variant?: "warning" | "danger";
}) {
  return (
    <Card>
      <CardContent>
        <p className="text-sm text-muted">{label}</p>
        <p
          className={`text-3xl font-bold mt-1 ${
            variant === "danger"
              ? "text-danger"
              : variant === "warning"
                ? "text-warning"
                : "text-foreground"
          }`}
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/dashboard/
git commit -m "feat(frontend): add dashboard page with stats and recent activities"
```

---

### Task 10: Verify Build

- [ ] **Step 1: Run lint check**

```bash
cd frontend && npx next lint
```

Expected: No errors.

- [ ] **Step 2: Run TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: No errors.

- [ ] **Step 3: Run build**

```bash
cd frontend && npm run build
```

Expected: Build succeeds.

- [ ] **Step 4: Final commit if any fixes needed**

If lint or build issues required fixes, commit them:

```bash
git add -A frontend/
git commit -m "fix(frontend): resolve lint and build issues"
```
