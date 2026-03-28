# Phase 5: Feature Pages Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the CRUD feature pages (tenants, properties, debts, transactions, audit) so landlords can manage their portfolio through the web UI.

**Architecture:** Each module gets a list page with a data table + create modal. All pages live under the `(app)` route group (already has sidebar + auth guard). A shared `useFetch` hook eliminates repeated loading/error boilerplate. List pages follow a consistent pattern: page header with "Add" button, data table, empty state. Create forms use the Modal component with controlled form state.

**Tech Stack:** Next.js 16 App Router, React 19, TypeScript, Tailwind CSS v4, existing UI primitives from `@/components/ui`.

**Existing infrastructure (do NOT recreate):**
- `@/lib/api` — typed API client with `api.get/post/put/delete`
- `@/lib/cn` — className merge utility
- `@/hooks/use-auth` — `useAuth()` with `user`, `hasRole`
- `@/types/api` — all TypeScript types matching backend DTOs
- `@/components/ui` — Button, Input, Select, Textarea, Card, Badge, Modal, Table*, Toast
- `@/app/(app)/layout.tsx` — auth guard + AppShell with sidebar

---

## File Structure

```
frontend/src/
├── hooks/
│   └── use-fetch.ts                          # Shared data-fetching hook
├── lib/
│   └── format.ts                             # Money/date formatting helpers
├── app/(app)/
│   ├── tenants/
│   │   └── page.tsx                          # Tenant list + create modal
│   ├── properties/
│   │   └── page.tsx                          # Property list + create modal
│   ├── debts/
│   │   └── page.tsx                          # Debt list + create/pay/cancel modals
│   ├── transactions/
│   │   └── page.tsx                          # Transaction list (read-only)
│   └── audit/
│       └── page.tsx                          # Audit log list (read-only)
```

---

### Task 1: Shared Hooks and Helpers

**Files:**
- Create: `frontend/src/hooks/use-fetch.ts`
- Create: `frontend/src/lib/format.ts`

- [ ] **Step 1: Create useFetch hook**

A generic hook that fetches paginated data from the API, with loading state, error handling, and a `refetch` function.

```typescript
// frontend/src/hooks/use-fetch.ts
"use client";

import { useCallback, useEffect, useState } from "react";
import { api } from "@/lib/api";

interface UseFetchResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useFetch<T>(path: string | null): UseFetchResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!path) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const result = await api.get<T>(path);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [path]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}
```

- [ ] **Step 2: Create format helpers**

```typescript
// frontend/src/lib/format.ts
import type { MoneyDTO } from "@/types/api";

export function formatMoney(money: MoneyDTO): string {
  const amount = parseFloat(money.amount);
  return `${money.currency} ${amount.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString();
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString();
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/hooks/use-fetch.ts frontend/src/lib/format.ts
git commit -m "feat(frontend): add useFetch hook and format helpers"
```

---

### Task 2: Tenants Page

**Files:**
- Create: `frontend/src/app/(app)/tenants/page.tsx`

- [ ] **Step 1: Create tenants list page with create modal**

```tsx
// frontend/src/app/(app)/tenants/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import {
  Button,
  Input,
  Textarea,
  Badge,
  Modal,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  useToast,
} from "@/components/ui";
import { api, ApiClientError } from "@/lib/api";
import { useFetch } from "@/hooks/use-fetch";
import { formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  TenantResponse,
  CreateTenantRequest,
} from "@/types/api";

export default function TenantsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const { data, loading, refetch } = useFetch<PaginatedResponse<TenantResponse>>("/tenants");
  const { toast } = useToast();

  const tenants = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Tenants</h1>
        <Button onClick={() => setShowCreate(true)}>Add Tenant</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading tenants...</p>
      ) : tenants.length === 0 ? (
        <p className="text-muted">No tenants yet. Add your first tenant to get started.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Phone</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {tenants.map((tenant) => (
              <TableRow key={tenant.id}>
                <TableCell className="font-medium">{tenant.full_name}</TableCell>
                <TableCell>{tenant.email ?? "—"}</TableCell>
                <TableCell>{tenant.phone_number ?? "—"}</TableCell>
                <TableCell>
                  <Badge variant={tenant.is_active ? "success" : "danger"}>
                    {tenant.is_active ? "Active" : "Inactive"}
                  </Badge>
                </TableCell>
                <TableCell>{formatDate(tenant.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreateTenantModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Tenant created successfully", "success");
        }}
      />
    </div>
  );
}

function CreateTenantModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [phoneNumber, setPhoneNumber] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreateTenantRequest = {
      full_name: fullName,
      ...(email && { email }),
      ...(phoneNumber && { phone_number: phoneNumber }),
      ...(notes && { notes }),
    };

    try {
      await api.post("/tenants", body);
      setFullName("");
      setEmail("");
      setPhoneNumber("");
      setNotes("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create tenant");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Tenant">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="tenant-name"
          label="Full Name"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          required
        />
        <Input
          id="tenant-email"
          label="Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
        />
        <Input
          id="tenant-phone"
          label="Phone Number"
          value={phoneNumber}
          onChange={(e) => setPhoneNumber(e.target.value)}
        />
        <Textarea
          id="tenant-notes"
          label="Notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Tenant"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/tenants/
git commit -m "feat(frontend): add tenants list page with create modal"
```

---

### Task 3: Properties Page

**Files:**
- Create: `frontend/src/app/(app)/properties/page.tsx`

- [ ] **Step 1: Create properties list page with create modal**

```tsx
// frontend/src/app/(app)/properties/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import {
  Button,
  Input,
  Select,
  Textarea,
  Badge,
  Modal,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  useToast,
} from "@/components/ui";
import { api, ApiClientError } from "@/lib/api";
import { useFetch } from "@/hooks/use-fetch";
import { formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  PropertyResponse,
  CreatePropertyRequest,
} from "@/types/api";

const propertyTypes = [
  { value: "LAND", label: "Land" },
  { value: "BUILDING", label: "Building" },
  { value: "RESIDENTIAL", label: "Residential" },
  { value: "COMMERCIAL", label: "Commercial" },
  { value: "OTHER", label: "Other" },
];

export default function PropertiesPage() {
  const [showCreate, setShowCreate] = useState(false);
  const { data, loading, refetch } = useFetch<PaginatedResponse<PropertyResponse>>("/properties");
  const { toast } = useToast();

  const properties = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Properties</h1>
        <Button onClick={() => setShowCreate(true)}>Add Property</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading properties...</p>
      ) : properties.length === 0 ? (
        <p className="text-muted">No properties yet. Add your first property to get started.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Code</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Size (sqm)</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {properties.map((property) => (
              <TableRow key={property.id}>
                <TableCell className="font-medium">{property.name}</TableCell>
                <TableCell>{property.property_code}</TableCell>
                <TableCell>{property.property_type}</TableCell>
                <TableCell>{property.size_in_sqm.toLocaleString()}</TableCell>
                <TableCell>
                  <Badge variant={property.is_active ? "success" : "danger"}>
                    {property.is_active ? "Active" : "Inactive"}
                  </Badge>
                </TableCell>
                <TableCell>{formatDate(property.created_at)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreatePropertyModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Property created successfully", "success");
        }}
      />
    </div>
  );
}

function CreatePropertyModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const [name, setName] = useState("");
  const [propertyCode, setPropertyCode] = useState("");
  const [propertyType, setPropertyType] = useState("LAND");
  const [sizeInSqm, setSizeInSqm] = useState("");
  const [monthlyRent, setMonthlyRent] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreatePropertyRequest = {
      name,
      property_code: propertyCode,
      property_type: propertyType,
      size_in_sqm: parseFloat(sizeInSqm),
      ...(monthlyRent && { monthly_rent_amount: parseFloat(monthlyRent) }),
      ...(description && { description }),
    };

    try {
      await api.post("/properties", body);
      setName("");
      setPropertyCode("");
      setPropertyType("LAND");
      setSizeInSqm("");
      setMonthlyRent("");
      setDescription("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create property");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Property">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="prop-name"
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
        <Input
          id="prop-code"
          label="Property Code"
          value={propertyCode}
          onChange={(e) => setPropertyCode(e.target.value)}
          required
        />
        <Select
          id="prop-type"
          label="Property Type"
          value={propertyType}
          onChange={(e) => setPropertyType(e.target.value)}
          options={propertyTypes}
        />
        <Input
          id="prop-size"
          label="Size (sqm)"
          type="number"
          step="0.01"
          min="0.01"
          value={sizeInSqm}
          onChange={(e) => setSizeInSqm(e.target.value)}
          required
        />
        <Input
          id="prop-rent"
          label="Monthly Rent Amount"
          type="number"
          step="0.01"
          min="0"
          value={monthlyRent}
          onChange={(e) => setMonthlyRent(e.target.value)}
        />
        <Textarea
          id="prop-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Property"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/properties/
git commit -m "feat(frontend): add properties list page with create modal"
```

---

### Task 4: Debts Page

**Files:**
- Create: `frontend/src/app/(app)/debts/page.tsx`

This is the most complex page — it includes the debt list, create debt modal, record payment modal, and cancel debt modal.

- [ ] **Step 1: Create debts list page with create, pay, and cancel modals**

```tsx
// frontend/src/app/(app)/debts/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import {
  Button,
  Input,
  Select,
  Textarea,
  Badge,
  Modal,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  useToast,
} from "@/components/ui";
import { api, ApiClientError } from "@/lib/api";
import { useFetch } from "@/hooks/use-fetch";
import { formatMoney, formatDate } from "@/lib/format";
import type {
  PaginatedResponse,
  DebtResponse,
  DebtStatus,
  CreateDebtRequest,
  RecordPaymentRequest,
  CancelDebtRequest,
  TenantResponse,
} from "@/types/api";

const debtTypes = [
  { value: "RENT", label: "Rent" },
  { value: "UTILITIES", label: "Utilities" },
  { value: "MAINTENANCE", label: "Maintenance" },
  { value: "PENALTY", label: "Penalty" },
  { value: "OTHER", label: "Other" },
];

const paymentMethods = [
  { value: "CASH", label: "Cash" },
  { value: "BANK_TRANSFER", label: "Bank Transfer" },
  { value: "MOBILE_MONEY", label: "Mobile Money" },
  { value: "CHECK", label: "Check" },
  { value: "CREDIT_CARD", label: "Credit Card" },
  { value: "OTHER", label: "Other" },
];

const statusVariant: Record<DebtStatus, "default" | "success" | "warning" | "danger" | "outline"> = {
  PENDING: "default",
  PARTIAL: "warning",
  PAID: "success",
  OVERDUE: "danger",
  CANCELLED: "outline",
};

export default function DebtsPage() {
  const [showCreate, setShowCreate] = useState(false);
  const [payDebt, setPayDebt] = useState<DebtResponse | null>(null);
  const [cancelDebt, setCancelDebt] = useState<DebtResponse | null>(null);
  const { data, loading, refetch } = useFetch<PaginatedResponse<DebtResponse>>("/debts");
  const { toast } = useToast();

  const debts = data?.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">Debts</h1>
        <Button onClick={() => setShowCreate(true)}>Add Debt</Button>
      </div>

      {loading ? (
        <p className="text-muted">Loading debts...</p>
      ) : debts.length === 0 ? (
        <p className="text-muted">No debts yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Type</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Original</TableHead>
              <TableHead>Balance</TableHead>
              <TableHead>Due Date</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {debts.map((debt) => (
              <TableRow key={debt.id}>
                <TableCell>{debt.debt_type}</TableCell>
                <TableCell className="font-medium max-w-[200px] truncate">
                  {debt.description}
                </TableCell>
                <TableCell>{formatMoney(debt.original_amount)}</TableCell>
                <TableCell>{formatMoney(debt.balance)}</TableCell>
                <TableCell>{formatDate(debt.due_date)}</TableCell>
                <TableCell>
                  <Badge variant={statusVariant[debt.status]}>{debt.status}</Badge>
                </TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    {(debt.status === "PENDING" || debt.status === "PARTIAL" || debt.status === "OVERDUE") && (
                      <Button size="sm" variant="outline" onClick={() => setPayDebt(debt)}>
                        Pay
                      </Button>
                    )}
                    {debt.status !== "PAID" && debt.status !== "CANCELLED" && (
                      <Button size="sm" variant="ghost" onClick={() => setCancelDebt(debt)}>
                        Cancel
                      </Button>
                    )}
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}

      <CreateDebtModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={() => {
          setShowCreate(false);
          refetch();
          toast("Debt created successfully", "success");
        }}
      />

      {payDebt && (
        <PayDebtModal
          debt={payDebt}
          open={!!payDebt}
          onClose={() => setPayDebt(null)}
          onPaid={() => {
            setPayDebt(null);
            refetch();
            toast("Payment recorded successfully", "success");
          }}
        />
      )}

      {cancelDebt && (
        <CancelDebtModal
          debt={cancelDebt}
          open={!!cancelDebt}
          onClose={() => setCancelDebt(null)}
          onCancelled={() => {
            setCancelDebt(null);
            refetch();
            toast("Debt cancelled", "success");
          }}
        />
      )}
    </div>
  );
}

function CreateDebtModal({
  open,
  onClose,
  onCreated,
}: {
  open: boolean;
  onClose: () => void;
  onCreated: () => void;
}) {
  const { data: tenantsData } = useFetch<PaginatedResponse<TenantResponse>>(open ? "/tenants" : null);
  const [tenantId, setTenantId] = useState("");
  const [debtType, setDebtType] = useState("RENT");
  const [description, setDescription] = useState("");
  const [amount, setAmount] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const tenantOptions = (tenantsData?.data ?? []).map((t) => ({
    value: t.id,
    label: t.full_name,
  }));

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CreateDebtRequest = {
      tenant_id: tenantId,
      debt_type: debtType as CreateDebtRequest["debt_type"],
      description,
      original_amount: { amount, currency: "PHP" },
      due_date: dueDate,
      ...(notes && { notes }),
    };

    try {
      await api.post("/debts", body);
      setTenantId("");
      setDebtType("RENT");
      setDescription("");
      setAmount("");
      setDueDate("");
      setNotes("");
      onCreated();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to create debt");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Debt">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Select
          id="debt-tenant"
          label="Tenant"
          value={tenantId}
          onChange={(e) => setTenantId(e.target.value)}
          options={tenantOptions}
          placeholder="Select a tenant"
          required
        />
        <Select
          id="debt-type"
          label="Debt Type"
          value={debtType}
          onChange={(e) => setDebtType(e.target.value)}
          options={debtTypes}
        />
        <Input
          id="debt-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          required
        />
        <Input
          id="debt-amount"
          label="Amount (PHP)"
          type="number"
          step="0.01"
          min="0.01"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          required
        />
        <Input
          id="debt-due"
          label="Due Date"
          type="date"
          value={dueDate}
          onChange={(e) => setDueDate(e.target.value)}
          required
        />
        <Textarea
          id="debt-notes"
          label="Notes"
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Creating..." : "Create Debt"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function PayDebtModal({
  debt,
  open,
  onClose,
  onPaid,
}: {
  debt: DebtResponse;
  open: boolean;
  onClose: () => void;
  onPaid: () => void;
}) {
  const [amount, setAmount] = useState(debt.balance.amount);
  const [paymentMethod, setPaymentMethod] = useState("CASH");
  const [transactionDate, setTransactionDate] = useState(
    new Date().toISOString().split("T")[0],
  );
  const [description, setDescription] = useState("");
  const [referenceNumber, setReferenceNumber] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: RecordPaymentRequest = {
      debt_id: debt.id,
      tenant_id: debt.tenant_id,
      amount: { amount, currency: debt.balance.currency },
      payment_method: paymentMethod,
      transaction_date: transactionDate,
      description: description || `Payment for ${debt.description}`,
      ...(referenceNumber && { reference_number: referenceNumber }),
    };

    try {
      await api.post("/transactions/payment", body);
      onPaid();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to record payment");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Record Payment">
      <p className="text-sm text-muted mb-4">
        Balance: <span className="font-semibold text-foreground">{formatMoney(debt.balance)}</span>
      </p>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Input
          id="pay-amount"
          label="Amount"
          type="number"
          step="0.01"
          min="0.01"
          max={debt.balance.amount}
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          required
        />
        <Select
          id="pay-method"
          label="Payment Method"
          value={paymentMethod}
          onChange={(e) => setPaymentMethod(e.target.value)}
          options={paymentMethods}
        />
        <Input
          id="pay-date"
          label="Transaction Date"
          type="date"
          value={transactionDate}
          onChange={(e) => setTransactionDate(e.target.value)}
          required
        />
        <Input
          id="pay-desc"
          label="Description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
        <Input
          id="pay-ref"
          label="Reference Number"
          value={referenceNumber}
          onChange={(e) => setReferenceNumber(e.target.value)}
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Cancel</Button>
          <Button type="submit" disabled={submitting}>
            {submitting ? "Recording..." : "Record Payment"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function CancelDebtModal({
  debt,
  open,
  onClose,
  onCancelled,
}: {
  debt: DebtResponse;
  open: boolean;
  onClose: () => void;
  onCancelled: () => void;
}) {
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setSubmitting(true);

    const body: CancelDebtRequest = { ...(reason && { reason }) };

    try {
      await api.put(`/debts/${debt.id}/cancel`, body);
      onCancelled();
    } catch (err) {
      setError(err instanceof ApiClientError ? err.message : "Failed to cancel debt");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Cancel Debt">
      <p className="text-sm text-muted mb-4">
        Cancel debt: <span className="font-semibold text-foreground">{debt.description}</span> ({formatMoney(debt.original_amount)})
      </p>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-[var(--radius-md)] bg-danger-light px-4 py-3 text-sm text-danger">
            {error}
          </div>
        )}
        <Textarea
          id="cancel-reason"
          label="Reason (optional)"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Why is this debt being cancelled?"
        />
        <div className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose}>Keep Debt</Button>
          <Button type="submit" variant="danger" disabled={submitting}>
            {submitting ? "Cancelling..." : "Cancel Debt"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/debts/
git commit -m "feat(frontend): add debts page with create, pay, and cancel modals"
```

---

### Task 5: Transactions Page

**Files:**
- Create: `frontend/src/app/(app)/transactions/page.tsx`

- [ ] **Step 1: Create transactions list page (read-only)**

```tsx
// frontend/src/app/(app)/transactions/page.tsx
"use client";

import {
  Badge,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui";
import { useFetch } from "@/hooks/use-fetch";
import { formatMoney, formatDate } from "@/lib/format";
import type { PaginatedResponse, TransactionResponse } from "@/types/api";

export default function TransactionsPage() {
  const { data, loading } = useFetch<PaginatedResponse<TransactionResponse>>("/transactions");

  const transactions = data?.data ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Transactions</h1>

      {loading ? (
        <p className="text-muted">Loading transactions...</p>
      ) : transactions.length === 0 ? (
        <p className="text-muted">No transactions yet.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Amount</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Reference</TableHead>
              <TableHead>Verified</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {transactions.map((tx) => (
              <TableRow key={tx.id}>
                <TableCell>{formatDate(tx.transaction_date)}</TableCell>
                <TableCell>
                  <Badge variant={tx.transaction_type === "PAYMENT" ? "success" : "warning"}>
                    {tx.transaction_type}
                  </Badge>
                </TableCell>
                <TableCell className="font-medium">{formatMoney(tx.amount)}</TableCell>
                <TableCell>{tx.payment_method}</TableCell>
                <TableCell className="max-w-[200px] truncate">{tx.description}</TableCell>
                <TableCell>{tx.reference_number ?? "—"}</TableCell>
                <TableCell>
                  <Badge variant={tx.is_verified ? "success" : "outline"}>
                    {tx.is_verified ? "Verified" : "Pending"}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/transactions/
git commit -m "feat(frontend): add transactions list page"
```

---

### Task 6: Audit Logs Page

**Files:**
- Create: `frontend/src/app/(app)/audit/page.tsx`

- [ ] **Step 1: Create audit logs page (read-only)**

```tsx
// frontend/src/app/(app)/audit/page.tsx
"use client";

import {
  Badge,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from "@/components/ui";
import { useFetch } from "@/hooks/use-fetch";
import { formatDateTime } from "@/lib/format";
import { useAuth } from "@/hooks/use-auth";
import type { PaginatedResponse, AuditEntryResponse } from "@/types/api";

export default function AuditPage() {
  const { hasRole } = useAuth();

  const endpoint = hasRole("admin") || hasRole("auditor") ? "/audit" : "/audit/landlord";
  const { data, loading } = useFetch<PaginatedResponse<AuditEntryResponse>>(endpoint);

  const entries = data?.data ?? [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">Audit Logs</h1>

      {loading ? (
        <p className="text-muted">Loading audit logs...</p>
      ) : entries.length === 0 ? (
        <p className="text-muted">No audit entries found.</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Timestamp</TableHead>
              <TableHead>User</TableHead>
              <TableHead>Action</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead>Method</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => (
              <TableRow key={entry.id}>
                <TableCell className="whitespace-nowrap">
                  {formatDateTime(entry.timestamp)}
                </TableCell>
                <TableCell>{entry.user_email}</TableCell>
                <TableCell>{entry.action}</TableCell>
                <TableCell>
                  {entry.resource_type}
                  <span className="text-muted text-xs ml-1">#{entry.resource_id.slice(0, 8)}</span>
                </TableCell>
                <TableCell>
                  <Badge variant="outline">{entry.method}</Badge>
                </TableCell>
                <TableCell>
                  <Badge variant={entry.success ? "success" : "danger"}>
                    {entry.status_code}
                  </Badge>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/app/\(app\)/audit/
git commit -m "feat(frontend): add audit logs page"
```

---

### Task 7: Verify Build

- [ ] **Step 1: Run TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: No errors.

- [ ] **Step 2: Run ESLint**

```bash
cd frontend && npx eslint src/
```

Expected: No errors.

- [ ] **Step 3: Run production build**

```bash
cd frontend && npm run build
```

Expected: Build succeeds, routes include `/tenants`, `/properties`, `/debts`, `/transactions`, `/audit`.

- [ ] **Step 4: Commit fixes if needed**

If lint or build issues required fixes, commit them:

```bash
git add -A frontend/
git commit -m "fix(frontend): resolve lint and build issues in feature pages"
```
