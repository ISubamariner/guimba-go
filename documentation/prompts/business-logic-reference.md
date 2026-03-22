# Guimba Debt Tracker — Complete Business Logic Reference

> **Purpose**: This document is the single source of truth for all business processes, domain rules, entity behaviors, and system workflows extracted from the original `guimba-debt-tracker` Python/FastAPI project. Any session working on Guimba-GO must reference this file to ensure behavioral parity with the original system.
>
> **Source**: Extracted from `guimba-debt-tracker` v3.1.0 (Python/FastAPI + Vue 3)  
> **Target**: `Guimba-GO` (Go/Chi + Next.js)  
> **Last Updated**: 2026-03-22

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Domain Entities](#2-domain-entities)
3. [Value Objects](#3-value-objects)
4. [Business Rules by Domain](#4-business-rules-by-domain)
5. [Service Layer Workflows](#5-service-layer-workflows)
6. [Authentication & Authorization](#6-authentication--authorization)
7. [Background Jobs & Notifications](#7-background-jobs--notifications)
8. [Dashboard & Reporting](#8-dashboard--reporting)
9. [OCR Receipt Scanning](#9-ocr-receipt-scanning)
10. [Audit System](#10-audit-system)
11. [Error Taxonomy](#11-error-taxonomy)
12. [API Endpoint Map](#12-api-endpoint-map)
13. [Data Export](#13-data-export)

---

## 1. System Overview

**Guimba Debt Tracker** is a multi-tenant web application for tracking debts and managing land parcels in the Municipality of Guimba. It operates in **Philippine Peso (PHP)** as the base currency.

### Core Domains
| Domain | Description |
|:---|:---|
| **User Management** | Registration, authentication, role-based access, profile management |
| **Tenant Management** | People who owe money; belong to a specific landlord |
| **Property Management** | Land parcels and buildings with GeoJSON support |
| **Debt Tracking** | Full ledger system: creation, payments, refunds, overdue detection |
| **Transaction Management** | Immutable payment/refund records with receipt tracking |
| **Audit System** | Immutable audit logs of all state changes (stored in MongoDB) |
| **Role & Permission Management** | Dynamic RBAC with granular permission control |
| **OCR Receipt Scanning** | Extract financial data from receipt images via AI vision model |
| **Notifications** | Email notifications for overdue debts via background workers |

### Multi-Tenant Data Isolation
- Each **Landlord** sees only their own tenants, properties, and debts
- **Super Admins** have full system access
- **Auditors** have read-only access to financials and audit logs
- **Tenants** have read-only access to their own debt information

---

## 2. Domain Entities

### 2.1 User
```
Fields:
  - id: UUID (PK)
  - email: string (unique, required)
  - full_name: string (required)
  - hashed_password: string (required)
  - role: UserRole enum
  - is_active: boolean (default: true)
  - is_email_verified: boolean (default: false)
  - last_login_at: datetime (nullable)
  - created_at: datetime
  - updated_at: datetime

Methods:
  - activate() → sets is_active = true
  - deactivate() → sets is_active = false
  - verify_email() → sets is_email_verified = true
  - update_password(new_hashed_password) → updates password
  - update_profile(full_name?, email?) → partial update
  - change_role(new_role) → changes role enum
  - record_login() → updates last_login_at to now
  - can_manage_users() → true if SUPER_ADMIN
  - can_manage_properties() → true if SUPER_ADMIN or LANDLORD
  - can_manage_debts() → true if SUPER_ADMIN or LANDLORD
  - can_view_audit_logs() → true if SUPER_ADMIN or AUDITOR
```

### 2.2 Tenant
```
Fields:
  - id: UUID (PK)
  - full_name: string (required)
  - email: string (nullable)
  - phone_number: string (nullable)
  - national_id: string (nullable)
  - address: Address value object
  - landlord_id: UUID (FK → User, required)
  - user_id: UUID (FK → User, nullable — linked login account)
  - is_active: boolean (default: true)
  - notes: string (nullable)
  - created_at: datetime
  - updated_at: datetime

Validation Rules:
  - Must have at least one contact method (email OR phone_number)
  - full_name is required and non-empty

Methods:
  - update_contact_info(email?, phone?, address?) → partial update
  - update_notes(notes) → replaces notes
  - deactivate() → sets is_active = false
  - reactivate() → sets is_active = true
  - has_contact_email() → boolean
  - has_contact_phone() → boolean
```

### 2.3 Property
```
Fields:
  - id: UUID (PK)
  - name: string (required)
  - property_code: string (unique, required)
  - address: Address value object
  - geojson_coordinates: string/JSON (nullable)
  - property_type: string (default: "LAND")
  - size_in_acres: decimal (nullable)
  - size_in_sqm: decimal (required, must be > 0)
  - owner_id: UUID (FK → User, required)
  - is_available_for_rent: boolean (default: true)
  - is_active: boolean (default: true)
  - monthly_rent_amount: decimal (nullable)
  - description: string (nullable)
  - created_at: datetime
  - updated_at: datetime

Methods:
  - update_details(name?, property_code?, ...) → partial update
  - update_address(address) → replaces address
  - update_geolocation(geojson) → updates coordinates
  - set_rent_amount(amount) → sets monthly rent
  - mark_as_rented() → is_available_for_rent = false
  - mark_as_available() → is_available_for_rent = true
  - deactivate() → is_active = false
  - reactivate() → is_active = true
  - has_geolocation() → boolean
```

### 2.4 Debt
```
Fields:
  - id: UUID (PK)
  - tenant_id: UUID (FK → Tenant, required)
  - landlord_id: UUID (FK → User, required)
  - property_id: UUID (FK → Property, nullable)
  - debt_type: DebtType enum (required)
  - description: string (required, non-empty)
  - original_amount: Money value object (required, must be > 0)
  - amount_paid: Money value object (starts at 0)
  - due_date: date (required)
  - status: DebtStatus enum (default: PENDING)
  - notes: string (nullable)
  - created_at: datetime
  - updated_at: datetime

Enums:
  DebtStatus: PENDING | PARTIAL | PAID | OVERDUE | CANCELLED
  DebtType: RENT | UTILITIES | MAINTENANCE | PENALTY | OTHER

Methods:
  - record_payment(amount: Money) → increases amount_paid, auto-transitions status:
      • If amount_paid == original_amount → status = PAID
      • If amount_paid < original_amount → status = PARTIAL
  - reverse_payment(amount: Money) → decreases amount_paid, recalculates status
  - mark_as_overdue() → status = OVERDUE
  - cancel() → status = CANCELLED
  - get_balance() → Money (original_amount - amount_paid)
  - is_fully_paid() → boolean (amount_paid >= original_amount)
  - is_overdue() → boolean (due_date < today AND status not in [PAID, CANCELLED])

Status Transition Rules:
  PENDING → PARTIAL (partial payment received)
  PENDING → PAID (full payment received)
  PENDING → OVERDUE (past due date, lazy detection)
  PENDING → CANCELLED (manual cancellation)
  PARTIAL → PAID (remaining balance paid)
  PARTIAL → OVERDUE (past due date, lazy detection)
  PARTIAL → CANCELLED (manual cancellation)
  OVERDUE → PARTIAL (partial payment on overdue debt)
  OVERDUE → PAID (full payment on overdue debt)
  OVERDUE → CANCELLED (manual cancellation)
  PAID → [no transitions allowed — terminal state]
  CANCELLED → [no transitions allowed — terminal state]

Invalid Transitions:
  PAID → CANCELLED ✗ (raises InvalidStateTransitionError)
  CANCELLED → any ✗
```

### 2.5 Transaction
```
Fields:
  - id: UUID (PK)
  - debt_id: UUID (FK → Debt, required)
  - tenant_id: UUID (FK → Tenant, required)
  - landlord_id: UUID (FK → User, required)
  - recorded_by_user_id: UUID (FK → User, nullable)
  - transaction_type: TransactionType enum (required)
  - amount: Money value object (required, must be > 0)
  - payment_method: PaymentMethod enum (required)
  - transaction_date: date (required)
  - description: string
  - receipt_number: string (nullable)
  - reference_number: string (nullable)
  - ocr_extracted_data: string (nullable — JSON from OCR scan)
  - is_verified: boolean (default: false)
  - verified_by_user_id: UUID (nullable)
  - verified_at: datetime (nullable)
  - created_at: datetime
  - updated_at: datetime

Enums:
  TransactionType: PAYMENT | REFUND | PENALTY | ADJUSTMENT
  PaymentMethod: CASH | BANK_TRANSFER | MOBILE_MONEY | CHECK | CREDIT_CARD | OTHER

Immutability Rule:
  - Transactions are IMMUTABLE once created
  - Only verification fields (is_verified, verified_by, verified_at) can be updated
  - No amount, type, or date changes after creation

Methods:
  - verify() → marks as verified
  - has_receipt_image() → boolean
  - has_ocr_data() → boolean
  - is_payment() → boolean (type == PAYMENT)
  - is_refund() → boolean (type == REFUND)
```

### 2.6 Role
```
Fields:
  - id: UUID (PK)
  - name: string (unique, normalized to lowercase_underscore)
  - display_name: string (required)
  - description: string
  - permissions: Set<string> (permission names)
  - is_system_role: boolean (default: false)
  - is_active: boolean (default: true)
  - created_at: datetime
  - updated_at: datetime

Methods:
  - add_permission(permission_name) → adds to set
  - remove_permission(permission_name) → removes from set
  - has_permission(name) → boolean
  - has_any_permission(names) → boolean
  - has_all_permissions(names) → boolean
  - update_details(display_name?, description?) → partial update
  - activate() / deactivate()
  - can_be_deleted() → false if is_system_role

System Roles (cannot be deleted):
  - super_admin, landlord, auditor, tenant
```

### 2.7 Permission
```
Fields:
  - id: UUID (PK)
  - name: string (unique, auto-normalized to lowercase_underscore)
  - display_name: string
  - description: string
  - category: string (grouping)
  - is_system_permission: boolean (default: false)
  - created_at: datetime
  - updated_at: datetime

Methods:
  - update_details(display_name?, description?, category?)
  - can_be_deleted() → false if is_system_permission
```

### 2.8 OcrResult
```
Fields:
  - id: UUID (PK)
  - scanned_at: datetime
  - raw_response: string
  - amount: decimal (nullable, must be >= 0)
  - currency: string (default: "PHP")
  - description: string (nullable)
  - document_date: date (nullable)
  - vendor_name: string (nullable)
  - reference_number: string (nullable)
  - confidence: float (0.0–1.0)

Properties:
  - is_high_confidence → true if confidence >= 0.7
```

---

## 3. Value Objects

### 3.1 Money (Immutable)
```
Fields:
  - amount: Decimal (non-negative, rounded to 2 decimal places)
  - currency: Currency enum

Operations:
  - add(other: Money) → Money (same currency required)
  - subtract(other: Money) → Money (same currency, result must be >= 0)
  - multiply(factor: Decimal) → Money
  - is_zero() → boolean
  - is_greater_than(other: Money) → boolean (same currency required)
  - zero(currency) → class method, creates Money(0, currency)

Currency Enum:
  PHP (base currency), USD, EUR, GBP, KES, TZS, UGX

Critical Rules:
  - All arithmetic operations MUST verify same currency
  - Cross-currency operations raise ValueError
  - Uses Decimal (not float) for precision
  - Amount auto-rounded to 2 decimal places on creation
  - Negative amounts raise ValueError
```

### 3.2 Address (Immutable)
```
Fields:
  - street: string (required)
  - city: string (required)
  - state_or_region: string (required)
  - postal_code: string (nullable)
  - country: string (default: "Philippines")

Methods:
  - full_address() → formatted string: "street, city, state, postal_code, country"
```

### 3.3 UserRole (Enum)
```
Values:
  - SUPER_ADMIN: Full system access, can manage other admins
  - LANDLORD: Manages own properties, tenants, debts
  - AUDITOR: Read-only access to financials and audit logs
  - TENANT: Read-only access to own debt information

Methods:
  - has_admin_privileges() → true if SUPER_ADMIN
  - can_manage_properties() → true if SUPER_ADMIN or LANDLORD
  - can_manage_debts() → true if SUPER_ADMIN or LANDLORD
  - can_view_audit_logs() → true if SUPER_ADMIN or AUDITOR
```

---

## 4. Business Rules by Domain

### 4.1 Debt Creation Rules
1. **Tenant must exist** and be active
2. **Landlord must exist** and be active
3. **Tenant must belong to the landlord** (`tenant.landlord_id == landlord_id`)
4. **Property must belong to the landlord** (if property_id provided: `property.owner_id == landlord_id`)
5. **Amount must be positive** (> 0)
6. **Description is required** and non-empty
7. **Initial status is always PENDING**
8. **Initial amount_paid is always 0** in the same currency as original_amount

### 4.2 Payment Rules
1. **Payment amount must be positive** (> 0)
2. **Currency must match** the debt's currency
3. **Debt must be in a payable state** — not PAID, not CANCELLED
4. **No overpayment** — payment amount must not exceed remaining balance (`get_balance()`)
5. **Duplicate payment detection** — if `reference_number` is provided, reject if same reference already exists on this debt for a PAYMENT transaction
6. **Tenant must match** the debt's tenant (`debt.tenant_id == tenant_id`)
7. **Landlord must match** the debt's landlord (`debt.landlord_id == landlord_id`)
8. After payment: debt's `amount_paid` increases, status auto-transitions (PARTIAL or PAID)

### 4.3 Refund Rules
1. **Refund amount must be positive** (> 0)
2. **Currency must match** the debt's currency
3. **Refund cannot exceed amount already paid** (`amount > debt.amount_paid` raises PaymentError)
4. **Tenant and landlord must match** the debt's records
5. After refund: debt's `amount_paid` decreases via `reverse_payment()`, status recalculates

### 4.4 Mark as Paid Rules
1. If there's an outstanding balance, **automatically records a payment** for the remaining amount (useful for cash payments where exact tracking isn't needed)
2. Forces status to PAID regardless of current amount_paid

### 4.5 Debt Cancellation Rules
1. **Cannot cancel a PAID debt** — raises `InvalidStateTransitionError`
2. Cancellation reason is appended to debt notes
3. Status forced to CANCELLED

### 4.6 Overdue Detection (Lazy)
The system uses **lazy overdue detection** — debts are NOT proactively scanned. Instead:
1. When a debt is **retrieved by ID**, if `is_overdue()` is true and status != OVERDUE → auto-mark as OVERDUE and persist
2. When debts are **listed** (by tenant or landlord), iterate all results and auto-mark overdue ones
3. The `is_overdue()` check: `due_date < today AND status NOT IN [PAID, CANCELLED]`

### 4.7 Property Rules
1. **Owner must exist** and have role LANDLORD or SUPER_ADMIN
2. **Property code must be unique** across all properties
3. **Size must be positive** (> 0)
4. **Cannot deactivate a property with outstanding debts** — must resolve or cancel all non-PAID/non-CANCELLED debts first
5. Deactivation is a soft delete (sets `is_active = false`)

### 4.8 Tenant Rules
1. **Landlord must exist** and have role LANDLORD or SUPER_ADMIN
2. **Email must be unique** across all tenants
3. **At least one contact method** required (email or phone)
4. **User account creation** — when creating a tenant, optionally creates a User with TENANT role:
   - If email matches existing user and that user has no tenant link → link them
   - If email matches existing user who already has a tenant → raise ValidationError
   - If no user exists → create new User with auto-generated secure password (12 chars: letters, digits, symbols)
5. Deactivation is a soft delete

### 4.9 User Management Rules
1. **Only SUPER_ADMIN can change user roles**
2. **Only SUPER_ADMIN can deactivate/delete users**
3. **Cannot deactivate/delete your own account** (self-protection)
4. **Email uniqueness** enforced at creation
5. **Hard delete** available (use with caution — prefer deactivation)

### 4.10 Role Management Rules
1. **Role names are normalized** to `lowercase_underscore` format
2. **Role names must be unique**
3. **System roles cannot be deleted** (super_admin, landlord, auditor, tenant)
4. **Permissions must exist** before being assigned to a role
5. Roles can be activated/deactivated

---

## 5. Service Layer Workflows

### 5.1 Debt Service — `create_debt`
```
Input: tenant_id, landlord_id, property_id?, amount, description, due_date, debt_type
Steps:
  1. Validate tenant exists → NotFoundError
  2. Validate landlord exists → NotFoundError
  3. If property_id: validate property exists → NotFoundError
  4. If property_id: validate property.owner_id == landlord_id → ValidationError
  5. Validate tenant.landlord_id == landlord_id → ValidationError
  6. Create Debt entity (status=PENDING, amount_paid=0)
  7. Persist to database
  8. Log audit: CREATE_DEBT (includes tenant name, property name, amount, currency, type)
  9. Return created debt
```

### 5.2 Transaction Service — `record_payment`
```
Input: debt_id, tenant_id, landlord_id, amount, payment_method, payment_date, reference_number?, notes?, ocr_data?
Steps:
  1. Validate debt exists → NotFoundError
  2. Validate tenant exists → NotFoundError
  3. Validate landlord exists → NotFoundError
  4. Validate debt.tenant_id == tenant_id → ValidationError
  5. Validate debt.landlord_id == landlord_id → ValidationError
  6. Validate amount.currency == debt.original_amount.currency → ValidationError
  7. Validate amount > 0 → ValidationError
  8. Validate debt status not in [PAID, CANCELLED] → ValidationError
  9. Validate amount <= debt.get_balance() (no overpayment) → ValidationError
  10. If reference_number: check for duplicate on same debt → ValidationError
  11. Create Transaction entity (type=PAYMENT)
  12. Persist transaction
  13. Call debt.record_payment(amount) → updates amount_paid and status
  14. Persist updated debt
  15. Return created transaction
```

### 5.3 Transaction Service — `record_refund`
```
Input: debt_id, tenant_id, landlord_id, amount, payment_method, refund_date, reference_number?, notes?
Steps:
  1. Validate debt exists → NotFoundError
  2. Validate tenant exists → NotFoundError
  3. Validate landlord exists → NotFoundError
  4. Validate amount > 0 → ValidationError
  5. Validate tenant/landlord match debt → ValidationError
  6. Validate currency match → ValidationError
  7. Validate amount <= debt.amount_paid (cannot refund more than paid) → PaymentError
  8. Create Transaction entity (type=REFUND)
  9. Persist transaction
  10. Call debt.reverse_payment(amount) → decreases amount_paid, recalculates status
  11. Persist updated debt
  12. Return created transaction
```

### 5.4 Debt Service — `apply_payment`
```
Input: debt_id, payment_amount
Steps:
  1. Get debt (triggers lazy overdue detection)
  2. Validate currency match → ValidationError
  3. Capture balance before payment
  4. Call debt.record_payment(payment_amount)
  5. Persist updated debt
  6. Log audit: APPLY_PAYMENT (includes balance_before, balance_after, tenant name, property name)
  7. Return updated debt
```

### 5.5 Debt Service — `mark_as_paid`
```
Input: debt_id
Steps:
  1. Get debt
  2. Calculate outstanding balance
  3. If outstanding > 0: auto-apply payment for remaining amount
  4. Force status = PAID
  5. Persist
  6. Log audit: MARK_DEBT_PAID (includes original_amount, status_before, status_after)
  7. Return updated debt
```

### 5.6 Debt Service — `cancel_debt`
```
Input: debt_id, reason?
Steps:
  1. Get debt
  2. If status == PAID → raise InvalidStateTransitionError
  3. Set status = CANCELLED
  4. Append cancellation reason to notes
  5. Persist
  6. Log audit: CANCEL_DEBT (includes reason, status_before, status_after)
  7. Return updated debt
```

### 5.7 Tenant Service — `create_tenant`
```
Input: landlord_id, name, email, phone, address, national_id?, notes?, password?, create_user_account?
Steps:
  1. Validate landlord exists and has LANDLORD/SUPER_ADMIN role → NotFoundError/ValidationError
  2. Check email uniqueness among tenants → DuplicateError
  3. If create_user_account == true AND email provided:
     a. Check if User with this email exists
     b. If exists AND has tenant link → ValidationError
     c. If exists AND no tenant link → reuse user_id
     d. If not exists → create User(role=TENANT, password=auto-generated or provided)
  4. Create Tenant entity with optional user_id link
  5. Persist
  6. Log audit: CREATE_TENANT
  7. Return created tenant
```

### 5.8 Property Service — `deactivate_property`
```
Input: property_id
Steps:
  1. Query all debts for this property
  2. Filter to outstanding debts (status not in [PAID, CANCELLED])
  3. If outstanding debts exist → ValidationError with count
  4. Set property.is_active = false
  5. Persist
  6. Return deactivated property
```

### 5.9 User Management — `change_user_role`
```
Input: user_id, new_role_name, admin_user_id
Steps:
  1. Get target user → NotFoundError
  2. Get admin user → NotFoundError
  3. Validate admin is SUPER_ADMIN → ValidationError
  4. Validate new role exists in role repository → NotFoundError
  5. Map role name to UserRole enum → ValidationError if invalid
  6. Call user.change_role(new_role)
  7. Persist
  8. Log audit: CHANGE_USER_ROLE (includes role_before, role_after)
  9. Return updated user
```

### 5.10 User Management — `deactivate_user`
```
Input: user_id, admin_user_id
Steps:
  1. Get target user → NotFoundError
  2. Get admin user → NotFoundError
  3. Validate admin is SUPER_ADMIN → ValidationError
  4. Validate user_id != admin_user_id (no self-deactivation) → ValidationError
  5. Call user.deactivate()
  6. Persist
  7. Log audit: DEACTIVATE_USER
  8. Return updated user
```

---

## 6. Authentication & Authorization

### 6.1 Registration
- Endpoint: `POST /api/v1/auth/register`
- Rate limited: **5 requests/minute**
- Allowed self-registration roles: **LANDLORD**, **AUDITOR**
- Blocked self-registration: **TENANT** (must be created by landlord), **SUPER_ADMIN** (must be created by another super admin)
- Email uniqueness enforced
- Password is bcrypt-hashed before storage

### 6.2 Login
- Endpoint: `POST /api/v1/auth/login`
- Rate limited: **10 requests/minute**
- Uses OAuth2 password flow (username = email)
- Returns both **access token** and **refresh token**
- Access token contains: `sub` (user_id), `role`, `jti` (unique ID), `exp`
- Refresh token contains: `sub` (user_id), `type: "refresh"`, `exp`
- Access token expiry: configurable (default 30 minutes)
- Refresh token expiry: configurable (default 7 days)
- Inactive users (is_active=false) are blocked from login (HTTP 403)

### 6.3 Token Refresh
- Endpoint: `POST /api/v1/auth/refresh`
- Validates refresh token type
- Verifies user still exists and is active
- Issues new access + refresh token pair (token rotation)

### 6.4 Logout
- Endpoint: `POST /api/v1/auth/logout`
- Adds the access token's `jti` to a **Redis blocklist** with TTL matching remaining token lifetime
- Blocklist entries auto-expire when the original token would have expired
- Graceful degradation: if Redis is unavailable, logout still returns success

### 6.5 Forgot Password
- Endpoint: `POST /api/v1/auth/forgot-password`
- **Always returns success** (prevents email enumeration)
- Generates a **password_reset JWT** with 15-minute expiry
- If SMTP configured: sends email with reset link
- If SMTP not configured (dev mode): returns reset token in response body
- Token contains: `sub` (user_id), `type: "password_reset"`, `exp`, `jti`

### 6.6 Reset Password
- Endpoint: `POST /api/v1/auth/reset-password`
- Validates reset token is a `password_reset` type
- Validates user exists and is active
- Hashes and updates password
- **Invalidates the reset token** by adding `jti` to Redis blocklist (prevents reuse)

### 6.7 Change Password
- Endpoint: `POST /api/v1/auth/change-password`
- Requires authentication
- Verifies current password before accepting change
- Rejects if new password == current password

---

## 7. Background Jobs & Notifications

### 7.1 Celery Worker Configuration
- Broker: Redis
- Serialization: JSON
- Timezone: UTC
- Task acknowledgment: late (ack after execution)
- Prefetch multiplier: 1 (one task at a time per worker)

### 7.2 Daily Overdue Notifications
- **Schedule**: Every 24 hours (86400 seconds) — runs at 08:00 UTC
- **Task expiry**: If not picked up within 1 hour, skip
- **Max retries**: 3 (with exponential backoff: 60s, 120s, 180s)
- **Workflow**:
  1. Get all active users with LANDLORD role
  2. For each landlord, calculate overdue debts using `CalculateOverdueDebtsUseCase`
  3. If landlord has overdue debts, send HTML email digest containing:
     - Overdue count
     - Total outstanding amount
     - Table of individual overdue debts (tenant, debt ID, due date, outstanding amount)
     - Report date
  4. Track: landlords notified, total overdue debts, emails successfully sent

### 7.3 Health Check Task
- Simple `health_check` task returns `{"status": "ok"}`
- Used to verify worker is alive

---

## 8. Dashboard & Reporting

### 8.1 Dashboard Statistics (`GET /api/v1/dashboard/stats`)
Returns for the current authenticated user (scoped to their portfolio):
- **total_tenants**: Count of active tenants belonging to this landlord
- **total_properties**: Count of active properties owned by this landlord
- **active_debts**: Count of debts with status PENDING or PARTIAL
- **overdue_debts**: Count of overdue debts

### 8.2 Recent Activities (`GET /api/v1/dashboard/recent-activities`)
- Returns recent audit log entries for the current user's portfolio
- Limit parameter: 1–50 (default: 10)
- Only shows **successful** actions
- Converts raw audit actions into **human-readable descriptions**:
  - `CREATE_TENANT` → "Added new tenant: {name}"
  - `UPDATE_TENANT` → "Updated tenant details: {changed_fields}"
  - `CREATE_DEBT` → "Created debt record: PHP 5,000 (rent) for {tenant} at {property}"
  - `APPLY_PAYMENT` → "Recorded payment: PHP 2,500 from {tenant} for {property} (Fully paid)"
  - `MARK_DEBT_PAID` → "Marked debt as paid: PHP 5,000 from {tenant}"
  - `CANCEL_DEBT` → "Cancelled debt for {tenant} at {property}: {reason}"
  - etc.

---

## 9. OCR Receipt Scanning

### 9.1 Scan Receipt (`POST /api/v1/ocr/scan-receipt`)
- **Requires authentication**
- **Requires GEMINI_API_KEY** environment variable
- Accepted file types: JPEG, PNG, WebP, HEIC, HEIF
- Maximum file size: **10 MB**
- Empty files rejected

### 9.2 Extracted Fields
- `amount`: decimal (nullable)
- `currency`: string (default: "PHP")
- `description`: string (nullable)
- `document_date`: ISO date string (nullable)
- `vendor_name`: string (nullable)
- `reference_number`: string (nullable)
- `confidence`: float 0.0–1.0

### 9.3 Confidence Threshold
- `is_high_confidence`: true if confidence >= 0.7
- Frontend uses this to auto-fill debt form fields or prompt for manual review

---

## 10. Audit System

### 10.1 Architecture
- **Storage**: MongoDB (append-only, immutable)
- **Triggered by**: All service layer mutations (create, update, delete, status changes)
- **NOT triggered by**: Read operations

### 10.2 Audit Log Fields
```
- user_id: UUID (who performed the action)
- user_email: string
- user_role: string
- action: string (e.g., CREATE_TENANT, APPLY_PAYMENT, CHANGE_USER_ROLE)
- resource_type: string (e.g., Tenant, Debt, User, Property)
- resource_id: UUID (what was affected)
- ip_address: string
- user_agent: string
- endpoint: string (API endpoint called)
- method: string (HTTP method)
- status_code: int (HTTP status)
- success: boolean
- error_message: string (if failed)
- metadata: dict (rich context — see below)
- timestamp: datetime
```

### 10.3 Audit Metadata Examples
```
CREATE_DEBT metadata:
  landlord_id, tenant_id, tenant_name, amount, currency, debt_type, description, property_id?, property_name?

APPLY_PAYMENT metadata:
  landlord_id, tenant_id, payment_amount, currency, balance_before, balance_after, debt_type, description?, tenant_name?, property_id?, property_name?

CHANGE_USER_ROLE metadata:
  role_before, role_after, target_user_email

UPDATE_TENANT metadata:
  landlord_id, changes: { field_name: { before: old_value, after: new_value } }
```

### 10.4 Audit Queries
- Filter by: user_id, action, resource_type, success, date range
- Pagination: skip/limit
- Landlord-scoped queries: returns actions by the landlord AND actions on their portfolio

---

## 11. Error Taxonomy

| Error Type | HTTP Status | When Raised |
|:---|:---|:---|
| `ValidationError` | 422 | Invalid input, business rule violation |
| `NotFoundError` | 404 | Entity not found by ID/email |
| `DuplicateError` | 409 | Unique constraint violation (email, property_code, role name) |
| `BusinessRuleViolation` | 422 | Domain invariant violated |
| `InvalidStateTransitionError` | 422 | Illegal status change (e.g., PAID → CANCELLED) |
| `PaymentError` | 422 | Payment processing failure (refund > paid, etc.) |
| `InsufficientPaymentError` | 422 | Payment amount too low |
| `InsufficientPermissionsError` | 403 | User lacks required role/permission |
| `ServiceError` | 500 | Generic service-layer failure |

### Error Response Shape
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "details": { "field": "value" }
  }
}
```

---

## 12. API Endpoint Map

### Authentication (6 endpoints)
| Method | Path | Description | Auth | Rate Limit |
|:---|:---|:---|:---|:---|
| POST | `/auth/register` | Register new user | No | 5/min |
| POST | `/auth/login` | Login, get tokens | No | 10/min |
| POST | `/auth/refresh` | Refresh access token | No | — |
| GET | `/auth/me` | Get current user | Yes | — |
| POST | `/auth/forgot-password` | Request password reset | No | — |
| POST | `/auth/reset-password` | Reset password with token | No | — |
| POST | `/auth/change-password` | Change own password | Yes | — |
| POST | `/auth/logout` | Logout (blocklist token) | Yes | — |

### Users (6 endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| GET | `/users` | List all users | Yes | SUPER_ADMIN |
| GET | `/users/{id}` | Get user by ID | Yes | SUPER_ADMIN |
| PUT | `/users/{id}` | Update user profile | Yes | SUPER_ADMIN / self |
| PUT | `/users/{id}/role` | Change user role | Yes | SUPER_ADMIN |
| PUT | `/users/{id}/activate` | Activate user | Yes | SUPER_ADMIN |
| PUT | `/users/{id}/deactivate` | Deactivate user | Yes | SUPER_ADMIN |
| DELETE | `/users/{id}` | Delete user | Yes | SUPER_ADMIN |

### Tenants (6 endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| POST | `/tenants` | Create tenant | Yes | LANDLORD+ |
| GET | `/tenants` | List tenants (filtered by landlord) | Yes | LANDLORD+ |
| GET | `/tenants/{id}` | Get tenant | Yes | LANDLORD+ |
| PUT | `/tenants/{id}` | Update tenant | Yes | LANDLORD+ |
| PUT | `/tenants/{id}/deactivate` | Deactivate tenant | Yes | LANDLORD+ |
| DELETE | `/tenants/{id}` | Delete tenant | Yes | LANDLORD+ |

### Properties (6 endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| POST | `/properties` | Create property | Yes | LANDLORD+ |
| GET | `/properties` | List properties (filtered by owner) | Yes | LANDLORD+ |
| GET | `/properties/{id}` | Get property | Yes | LANDLORD+ |
| PUT | `/properties/{id}` | Update property | Yes | LANDLORD+ |
| PUT | `/properties/{id}/deactivate` | Deactivate property | Yes | LANDLORD+ |
| DELETE | `/properties/{id}` | Delete property | Yes | LANDLORD+ |

### Debts (8+ endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| POST | `/debts` | Create debt | Yes | LANDLORD+ |
| GET | `/debts` | List debts (filtered) | Yes | LANDLORD+ |
| GET | `/debts/{id}` | Get debt (triggers lazy overdue) | Yes | LANDLORD+ |
| PUT | `/debts/{id}` | Update debt details | Yes | LANDLORD+ |
| PUT | `/debts/{id}/pay` | Mark as paid | Yes | LANDLORD+ |
| PUT | `/debts/{id}/cancel` | Cancel debt | Yes | LANDLORD+ |
| GET | `/debts/overdue` | List overdue debts | Yes | LANDLORD+ |

### Transactions (6+ endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| POST | `/transactions/payment` | Record payment | Yes | LANDLORD+ |
| POST | `/transactions/refund` | Record refund | Yes | LANDLORD+ |
| GET | `/transactions/{id}` | Get transaction | Yes | LANDLORD+ |
| PUT | `/transactions/{id}/verify` | Verify transaction | Yes | LANDLORD+ |
| GET | `/transactions/debt/{debt_id}` | Transactions by debt | Yes | LANDLORD+ |
| GET | `/transactions/tenant/{tenant_id}` | Transactions by tenant | Yes | LANDLORD+ |

### Roles (CRUD endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| POST | `/roles` | Create role | Yes | SUPER_ADMIN |
| GET | `/roles` | List roles | Yes | SUPER_ADMIN |
| GET | `/roles/{id}` | Get role | Yes | SUPER_ADMIN |
| PUT | `/roles/{id}` | Update role | Yes | SUPER_ADMIN |
| POST | `/roles/{id}/permissions` | Add permissions | Yes | SUPER_ADMIN |
| DELETE | `/roles/{id}/permissions` | Remove permissions | Yes | SUPER_ADMIN |
| DELETE | `/roles/{id}` | Delete role | Yes | SUPER_ADMIN |

### Dashboard (2 endpoints)
| Method | Path | Description | Auth |
|:---|:---|:---|:---|
| GET | `/dashboard/stats` | Dashboard statistics | Yes |
| GET | `/dashboard/recent-activities` | Recent activity feed | Yes |

### OCR (1 endpoint)
| Method | Path | Description | Auth |
|:---|:---|:---|:---|
| POST | `/ocr/scan-receipt` | Scan receipt image | Yes |

### Audit (2+ endpoints)
| Method | Path | Description | Auth | Role |
|:---|:---|:---|:---|:---|
| GET | `/audit` | Query audit logs | Yes | SUPER_ADMIN / AUDITOR |
| GET | `/audit/landlord` | Landlord-scoped audit logs | Yes | LANDLORD+ |

### Health (1 endpoint)
| Method | Path | Description | Auth |
|:---|:---|:---|:---|
| GET | `/health` | Health check | No |

---

## 13. Data Export

The original system supports CSV and PDF export for:
- **Tenants** — full list with contact info
- **Properties** — full list with addresses and sizes
- **Debts** — filtered by status, tenant, landlord
- **Audit Logs** — filtered by date range, action type

Export endpoints are typically scoped to the requesting landlord's data.

---

## Usage as a Reference Prompt

When implementing any feature in Guimba-GO, reference this document to ensure:

1. **Entity structures** match the field definitions in Section 2
2. **Business rules** in Section 4 are enforced in the Go use case/service layer
3. **Validation sequences** in Section 5 are followed step-by-step
4. **Auth flows** in Section 6 are replicated (rate limiting, token rotation, blocklist)
5. **Error types** in Section 11 are mapped to appropriate Go error types in `pkg/apperror/`
6. **Audit logging** in Section 10 captures the same metadata fields
7. **API endpoints** in Section 12 are implemented with the same HTTP methods and paths
