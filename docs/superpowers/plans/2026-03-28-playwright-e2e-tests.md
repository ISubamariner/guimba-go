# Playwright E2E Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Playwright E2E tests covering all critical user flows across the Guimba-GO frontend.

**Architecture:** Playwright with Chromium, page object model pattern, hybrid data strategy (global-setup admin user + per-test API data), tests in `tests/playwright/`.

**Tech Stack:** Playwright, TypeScript, Node.js

**Prerequisites:** Backend running at `http://localhost:8080`, frontend at `http://localhost:3000`, clean database.

---

## File Structure

| File | Purpose |
|---|---|
| `tests/playwright/package.json` | Playwright dependencies |
| `tests/playwright/tsconfig.json` | TypeScript config for test files |
| `tests/playwright/playwright.config.ts` | Playwright config (base URL, timeouts, global setup) |
| `tests/playwright/helpers/api-client.ts` | Typed API client for test data setup |
| `tests/playwright/helpers/global-setup.ts` | Register admin user + save auth state |
| `tests/playwright/fixtures/auth.fixture.ts` | Extended `test` with pre-authenticated page |
| `tests/playwright/pages/login.page.ts` | Login page object |
| `tests/playwright/pages/dashboard.page.ts` | Dashboard page object |
| `tests/playwright/pages/tenants.page.ts` | Tenants page object |
| `tests/playwright/pages/properties.page.ts` | Properties page object |
| `tests/playwright/pages/debts.page.ts` | Debts page object |
| `tests/playwright/pages/transactions.page.ts` | Transactions page object |
| `tests/playwright/pages/audit.page.ts` | Audit page object |
| `tests/playwright/specs/auth/login.spec.ts` | Login flow tests |
| `tests/playwright/specs/auth/register.spec.ts` | Register flow tests |
| `tests/playwright/specs/auth/auth-guard.spec.ts` | Auth guard redirect tests |
| `tests/playwright/specs/dashboard/dashboard.spec.ts` | Dashboard stats + activities tests |
| `tests/playwright/specs/tenants/tenants-crud.spec.ts` | Tenant CRUD tests |
| `tests/playwright/specs/properties/properties-crud.spec.ts` | Property CRUD tests |
| `tests/playwright/specs/debts/debts-crud.spec.ts` | Debt create + pay + cancel tests |
| `tests/playwright/specs/transactions/transactions.spec.ts` | Transaction list tests |
| `tests/playwright/specs/audit/audit.spec.ts` | Audit log tests |
| `tests/playwright/specs/navigation/sidebar.spec.ts` | Role-based sidebar tests |

---

### Task 1: Initialize Playwright Project

**Files:**
- Create: `tests/playwright/package.json`
- Create: `tests/playwright/tsconfig.json`
- Create: `tests/playwright/playwright.config.ts`

- [ ] **Step 1: Create package.json**

```json
{
  "name": "guimba-go-e2e",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "test": "playwright test",
    "test:ui": "playwright test --ui",
    "test:headed": "playwright test --headed"
  },
  "devDependencies": {
    "@playwright/test": "^1.52.0"
  }
}
```

- [ ] **Step 2: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist",
    "rootDir": ".",
    "baseUrl": ".",
    "paths": {
      "@pages/*": ["./pages/*"],
      "@helpers/*": ["./helpers/*"],
      "@fixtures/*": ["./fixtures/*"]
    }
  },
  "include": ["**/*.ts"]
}
```

- [ ] **Step 3: Create playwright.config.ts**

```typescript
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./specs",
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",
  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  globalSetup: "./helpers/global-setup.ts",
  projects: [
    {
      name: "chromium",
      use: { browserName: "chromium" },
    },
  ],
});
```

- [ ] **Step 4: Install dependencies**

Run: `cd tests/playwright && npm install`
Expected: `node_modules/` created, `package-lock.json` generated

- [ ] **Step 5: Install Playwright browsers**

Run: `cd tests/playwright && npx playwright install chromium`
Expected: Chromium browser downloaded

- [ ] **Step 6: Commit**

```bash
git add tests/playwright/package.json tests/playwright/package-lock.json tests/playwright/tsconfig.json tests/playwright/playwright.config.ts
git commit -m "test(e2e): initialize Playwright project with config"
```

---

### Task 2: API Client Helper + Global Setup

**Files:**
- Create: `tests/playwright/helpers/api-client.ts`
- Create: `tests/playwright/helpers/global-setup.ts`

- [ ] **Step 1: Create the API client helper**

This provides typed methods for creating test data via the backend API. Tests use this to set up state before browser interactions.

```typescript
const API_BASE = process.env.API_URL || "http://localhost:8080/api/v1";

interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

interface AuthResponse {
  user: { id: string; email: string; full_name: string; roles: { id: string; name: string }[] };
  access_token: string;
  refresh_token: string;
}

async function apiRequest<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${options.method || "GET"} ${path} failed (${res.status}): ${body}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export class TestApiClient {
  private tokens: AuthTokens | null = null;

  private authHeaders(): Record<string, string> {
    if (!this.tokens) return {};
    return { Authorization: `Bearer ${this.tokens.access_token}` };
  }

  async register(email: string, fullName: string, password: string): Promise<AuthResponse> {
    const res = await apiRequest<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, full_name: fullName, password }),
    });
    this.tokens = { access_token: res.access_token, refresh_token: res.refresh_token };
    return res;
  }

  async login(email: string, password: string): Promise<AuthResponse> {
    const res = await apiRequest<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    this.tokens = { access_token: res.access_token, refresh_token: res.refresh_token };
    return res;
  }

  async createTenant(data: { full_name: string; email?: string; phone_number?: string }): Promise<{ id: string }> {
    return apiRequest("/tenants", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async createProperty(data: {
    name: string;
    property_code: string;
    property_type: string;
    size_in_sqm: number;
  }): Promise<{ id: string }> {
    return apiRequest("/properties", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async createDebt(data: {
    tenant_id: string;
    debt_type: string;
    description: string;
    original_amount: { amount: string; currency: string };
    due_date: string;
  }): Promise<{ id: string }> {
    return apiRequest("/debts", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  async payDebt(data: {
    debt_id: string;
    tenant_id: string;
    amount: { amount: string; currency: string };
    payment_method: string;
    transaction_date: string;
    description: string;
  }): Promise<{ id: string }> {
    return apiRequest("/transactions/payment", {
      method: "POST",
      body: JSON.stringify(data),
      headers: this.authHeaders(),
    });
  }

  getTokens(): AuthTokens | null {
    return this.tokens;
  }
}
```

- [ ] **Step 2: Create global setup**

This registers an admin test user before all tests run, and saves auth tokens to a file so the auth fixture can reuse them.

```typescript
import { TestApiClient } from "./api-client";
import * as fs from "fs";
import * as path from "path";

const ADMIN_EMAIL = "e2e-admin@guimba.test";
const ADMIN_NAME = "E2E Admin";
const ADMIN_PASSWORD = "TestPassword123!";
const AUTH_STATE_PATH = path.join(__dirname, "../.auth-state.json");

async function globalSetup() {
  const api = new TestApiClient();

  try {
    await api.register(ADMIN_EMAIL, ADMIN_NAME, ADMIN_PASSWORD);
  } catch {
    // User may already exist from a previous run — try login instead
    await api.login(ADMIN_EMAIL, ADMIN_PASSWORD);
  }

  const tokens = api.getTokens();
  fs.writeFileSync(AUTH_STATE_PATH, JSON.stringify({
    email: ADMIN_EMAIL,
    password: ADMIN_PASSWORD,
    fullName: ADMIN_NAME,
    tokens,
  }));
}

export default globalSetup;
```

- [ ] **Step 3: Add .auth-state.json to .gitignore**

Append to `tests/playwright/.gitignore` (create if needed):

```
node_modules/
test-results/
playwright-report/
.auth-state.json
```

- [ ] **Step 4: Verify global setup runs**

Run: `cd tests/playwright && npx playwright test --list`
Expected: Global setup runs without error (registers or logs in admin user). No tests found yet (specs dir is empty of test files).

- [ ] **Step 5: Commit**

```bash
git add tests/playwright/helpers/api-client.ts tests/playwright/helpers/global-setup.ts tests/playwright/.gitignore
git commit -m "test(e2e): add API client helper and global setup for admin user"
```

---

### Task 3: Auth Fixture + Login Page Object

**Files:**
- Create: `tests/playwright/fixtures/auth.fixture.ts`
- Create: `tests/playwright/pages/login.page.ts`

- [ ] **Step 1: Create the auth fixture**

This extends Playwright's `test` to provide a pre-authenticated page. It reads the auth state saved by global setup and injects tokens into localStorage before navigating.

```typescript
import { test as base, type Page } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";

interface AuthState {
  email: string;
  password: string;
  fullName: string;
  tokens: {
    access_token: string;
    refresh_token: string;
  };
}

function loadAuthState(): AuthState {
  const filePath = path.join(__dirname, "../.auth-state.json");
  return JSON.parse(fs.readFileSync(filePath, "utf-8"));
}

export const test = base.extend<{ authedPage: Page }>({
  authedPage: async ({ page }, use) => {
    const authState = loadAuthState();

    // Navigate to app first so localStorage is on the correct origin
    await page.goto("/login");

    // Inject tokens into localStorage
    await page.evaluate((tokens) => {
      localStorage.setItem("access_token", tokens.access_token);
      localStorage.setItem("refresh_token", tokens.refresh_token);
    }, authState.tokens);

    // Navigate to dashboard — auth provider will pick up tokens
    await page.goto("/dashboard");
    await page.waitForSelector("h1:has-text('Dashboard')");

    await use(page);
  },
});

export { expect } from "@playwright/test";
export { loadAuthState };
```

- [ ] **Step 2: Create the login page object**

Page objects encapsulate selectors so tests don't hardcode CSS/text queries.

```typescript
import type { Page } from "@playwright/test";

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/login");
  }

  async login(email: string, password: string) {
    await this.page.fill("#email", email);
    await this.page.fill("#password", password);
    await this.page.click('button[type="submit"]');
  }

  async getErrorMessage(): Promise<string | null> {
    const error = this.page.locator(".bg-danger-light");
    if (await error.isVisible()) {
      return error.textContent();
    }
    return null;
  }

  async getHeading(): Promise<string | null> {
    return this.page.locator("h1").textContent();
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add tests/playwright/fixtures/auth.fixture.ts tests/playwright/pages/login.page.ts
git commit -m "test(e2e): add auth fixture and login page object"
```

---

### Task 4: Auth Specs (Login, Register, Auth Guard)

**Files:**
- Create: `tests/playwright/specs/auth/login.spec.ts`
- Create: `tests/playwright/specs/auth/register.spec.ts`
- Create: `tests/playwright/specs/auth/auth-guard.spec.ts`

- [ ] **Step 1: Create login spec**

```typescript
import { test, expect } from "@playwright/test";
import { LoginPage } from "../../pages/login.page";
import { loadAuthState } from "../../fixtures/auth.fixture";

test.describe("Login", () => {
  test("successful login redirects to dashboard", async ({ page }) => {
    const loginPage = new LoginPage(page);
    const authState = loadAuthState();

    await loginPage.goto();
    await loginPage.login(authState.email, authState.password);

    await page.waitForURL("**/dashboard");
    await expect(page.locator("h1")).toHaveText("Dashboard");
  });

  test("invalid credentials show error message", async ({ page }) => {
    const loginPage = new LoginPage(page);

    await loginPage.goto();
    await loginPage.login("wrong@example.com", "wrongpassword");

    await expect(page.locator(".bg-danger-light")).toBeVisible();
  });

  test("login page has link to register", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    const registerLink = page.locator('a[href="/register"]');
    await expect(registerLink).toBeVisible();
    await expect(registerLink).toHaveText("Register");
  });
});
```

- [ ] **Step 2: Create register spec**

```typescript
import { test, expect } from "@playwright/test";

test.describe("Register", () => {
  test("password mismatch shows client-side error", async ({ page }) => {
    await page.goto("/register");

    await page.fill("#fullName", "Test User");
    await page.fill("#email", "mismatch@guimba.test");
    await page.fill("#password", "Password123!");
    await page.fill("#confirmPassword", "DifferentPass123!");
    await page.click('button[type="submit"]');

    await expect(page.locator(".bg-danger-light")).toHaveText("Passwords do not match");
  });

  test("short password shows client-side error", async ({ page }) => {
    await page.goto("/register");

    await page.fill("#fullName", "Test User");
    await page.fill("#email", "short@guimba.test");
    await page.fill("#password", "short");
    await page.fill("#confirmPassword", "short");
    await page.click('button[type="submit"]');

    await expect(page.locator(".bg-danger-light")).toHaveText("Password must be at least 8 characters");
  });

  test("register page has link to login", async ({ page }) => {
    await page.goto("/register");

    const loginLink = page.locator('a[href="/login"]');
    await expect(loginLink).toBeVisible();
    await expect(loginLink).toHaveText("Sign in");
  });
});
```

- [ ] **Step 3: Create auth guard spec**

```typescript
import { test, expect } from "@playwright/test";

test.describe("Auth Guard", () => {
  test("unauthenticated user is redirected to login from dashboard", async ({ page }) => {
    // Clear any existing tokens
    await page.goto("/login");
    await page.evaluate(() => {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
    });

    await page.goto("/dashboard");
    await page.waitForURL("**/login");
    await expect(page).toHaveURL(/\/login/);
  });

  test("unauthenticated user is redirected to login from tenants", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
    });

    await page.goto("/tenants");
    await page.waitForURL("**/login");
    await expect(page).toHaveURL(/\/login/);
  });
});
```

- [ ] **Step 4: Run auth specs**

Run: `cd tests/playwright && npx playwright test specs/auth/`
Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
git add tests/playwright/specs/auth/
git commit -m "test(e2e): add auth specs (login, register, auth guard)"
```

---

### Task 5: Dashboard Page Object + Spec

**Files:**
- Create: `tests/playwright/pages/dashboard.page.ts`
- Create: `tests/playwright/specs/dashboard/dashboard.spec.ts`

- [ ] **Step 1: Create dashboard page object**

```typescript
import type { Page } from "@playwright/test";

export class DashboardPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/dashboard");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getStatCard(label: string) {
    return this.page.locator(`p.text-muted:has-text("${label}")`).locator("..");
  }

  getStatValue(label: string) {
    return this.getStatCard(label).locator("p.text-3xl");
  }

  getActivitiesSection() {
    return this.page.locator('text="Recent Activities"').locator("..");
  }

  getActivityItems() {
    return this.page.locator(".border-b.border-border");
  }
}
```

- [ ] **Step 2: Create dashboard spec**

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { DashboardPage } from "../../pages/dashboard.page";

test.describe("Dashboard", () => {
  test("displays stats cards with numeric values", async ({ authedPage }) => {
    const dashboard = new DashboardPage(authedPage);

    await expect(dashboard.getHeading()).toHaveText("Dashboard");
    await expect(dashboard.getStatValue("Total Tenants")).toBeVisible();
    await expect(dashboard.getStatValue("Total Properties")).toBeVisible();
    await expect(dashboard.getStatValue("Active Debts")).toBeVisible();
    await expect(dashboard.getStatValue("Overdue Debts")).toBeVisible();
  });

  test("displays recent activities section", async ({ authedPage }) => {
    const dashboard = new DashboardPage(authedPage);

    await expect(dashboard.getActivitiesSection()).toBeVisible();
  });
});
```

- [ ] **Step 3: Run dashboard spec**

Run: `cd tests/playwright && npx playwright test specs/dashboard/`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add tests/playwright/pages/dashboard.page.ts tests/playwright/specs/dashboard/
git commit -m "test(e2e): add dashboard page object and specs"
```

---

### Task 6: Tenants Page Object + Spec

**Files:**
- Create: `tests/playwright/pages/tenants.page.ts`
- Create: `tests/playwright/specs/tenants/tenants-crud.spec.ts`

- [ ] **Step 1: Create tenants page object**

```typescript
import type { Page } from "@playwright/test";

export class TenantsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/tenants");
    await this.page.waitForSelector("h1:has-text('Tenants')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Tenant')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No tenants yet");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: { fullName: string; email?: string; phone?: string; notes?: string }) {
    await this.page.fill("#tenant-name", data.fullName);
    if (data.email) await this.page.fill("#tenant-email", data.email);
    if (data.phone) await this.page.fill("#tenant-phone", data.phone);
    if (data.notes) await this.page.fill("#tenant-notes", data.notes);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Tenant")');
  }

  getRowByName(name: string) {
    return this.page.locator(`tbody tr:has-text("${name}")`);
  }
}
```

- [ ] **Step 2: Create tenants CRUD spec**

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { TenantsPage } from "../../pages/tenants.page";

test.describe("Tenants", () => {
  test("displays tenants page with heading and add button", async ({ authedPage }) => {
    const tenantsPage = new TenantsPage(authedPage);
    await tenantsPage.goto();

    await expect(tenantsPage.getHeading()).toHaveText("Tenants");
    await expect(tenantsPage.getAddButton()).toBeVisible();
  });

  test("create tenant via modal and verify it appears in table", async ({ authedPage }) => {
    const tenantsPage = new TenantsPage(authedPage);
    await tenantsPage.goto();

    await tenantsPage.openCreateModal();

    const tenantName = `E2E Tenant ${Date.now()}`;
    await tenantsPage.fillCreateForm({
      fullName: tenantName,
      email: `e2e-${Date.now()}@tenant.test`,
      phone: "09171234567",
    });
    await tenantsPage.submitCreateForm();

    // Modal closes and table updates
    await expect(tenantsPage.getRowByName(tenantName)).toBeVisible({ timeout: 10_000 });
  });

  test("created tenant shows Active badge", async ({ authedPage }) => {
    const tenantsPage = new TenantsPage(authedPage);
    await tenantsPage.goto();

    await tenantsPage.openCreateModal();

    const tenantName = `Badge Tenant ${Date.now()}`;
    await tenantsPage.fillCreateForm({ fullName: tenantName });
    await tenantsPage.submitCreateForm();

    const row = tenantsPage.getRowByName(tenantName);
    await expect(row).toBeVisible({ timeout: 10_000 });
    await expect(row.locator("text=Active")).toBeVisible();
  });
});
```

- [ ] **Step 3: Run tenants spec**

Run: `cd tests/playwright && npx playwright test specs/tenants/`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add tests/playwright/pages/tenants.page.ts tests/playwright/specs/tenants/
git commit -m "test(e2e): add tenants page object and CRUD specs"
```

---

### Task 7: Properties Page Object + Spec

**Files:**
- Create: `tests/playwright/pages/properties.page.ts`
- Create: `tests/playwright/specs/properties/properties-crud.spec.ts`

- [ ] **Step 1: Create properties page object**

```typescript
import type { Page } from "@playwright/test";

export class PropertiesPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/properties");
    await this.page.waitForSelector("h1:has-text('Properties')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Property')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: {
    name: string;
    code: string;
    type?: string;
    size: string;
    rent?: string;
  }) {
    await this.page.fill("#prop-name", data.name);
    await this.page.fill("#prop-code", data.code);
    if (data.type) await this.page.selectOption("#prop-type", data.type);
    await this.page.fill("#prop-size", data.size);
    if (data.rent) await this.page.fill("#prop-rent", data.rent);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Property")');
  }

  getRowByName(name: string) {
    return this.page.locator(`tbody tr:has-text("${name}")`);
  }
}
```

- [ ] **Step 2: Create properties CRUD spec**

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { PropertiesPage } from "../../pages/properties.page";

test.describe("Properties", () => {
  test("displays properties page with heading and add button", async ({ authedPage }) => {
    const propsPage = new PropertiesPage(authedPage);
    await propsPage.goto();

    await expect(propsPage.getHeading()).toHaveText("Properties");
    await expect(propsPage.getAddButton()).toBeVisible();
  });

  test("create property via modal and verify it appears in table", async ({ authedPage }) => {
    const propsPage = new PropertiesPage(authedPage);
    await propsPage.goto();

    await propsPage.openCreateModal();

    const propName = `E2E Property ${Date.now()}`;
    const propCode = `E2E-${Date.now()}`;
    await propsPage.fillCreateForm({
      name: propName,
      code: propCode,
      type: "RESIDENTIAL",
      size: "100",
      rent: "5000",
    });
    await propsPage.submitCreateForm();

    const row = propsPage.getRowByName(propName);
    await expect(row).toBeVisible({ timeout: 10_000 });
    await expect(row.locator("text=RESIDENTIAL")).toBeVisible();
    await expect(row.locator("text=Active")).toBeVisible();
  });
});
```

- [ ] **Step 3: Run properties spec**

Run: `cd tests/playwright && npx playwright test specs/properties/`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add tests/playwright/pages/properties.page.ts tests/playwright/specs/properties/
git commit -m "test(e2e): add properties page object and CRUD specs"
```

---

### Task 8: Debts Page Object + Spec

**Files:**
- Create: `tests/playwright/pages/debts.page.ts`
- Create: `tests/playwright/specs/debts/debts-crud.spec.ts`

- [ ] **Step 1: Create debts page object**

```typescript
import type { Page } from "@playwright/test";

export class DebtsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/debts");
    await this.page.waitForSelector("h1:has-text('Debts')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Debt')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No debts yet");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: {
    tenantName: string;
    type?: string;
    description: string;
    amount: string;
    dueDate: string;
  }) {
    // Select tenant by visible text
    await this.page.selectOption("#debt-tenant", { label: data.tenantName });
    if (data.type) await this.page.selectOption("#debt-type", data.type);
    await this.page.fill("#debt-desc", data.description);
    await this.page.fill("#debt-amount", data.amount);
    await this.page.fill("#debt-due", data.dueDate);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Debt")');
  }

  getRowByDescription(desc: string) {
    return this.page.locator(`tbody tr:has-text("${desc}")`);
  }

  getStatusBadge(row: ReturnType<Page["locator"]>) {
    return row.locator("span").filter({ hasText: /PENDING|PARTIAL|PAID|OVERDUE|CANCELLED/ });
  }

  async clickPayButton(row: ReturnType<Page["locator"]>) {
    await row.locator('button:has-text("Pay")').click();
    await this.page.waitForSelector("dialog[open]");
  }

  async clickCancelButton(row: ReturnType<Page["locator"]>) {
    await row.locator('button:has-text("Cancel")').click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillPayForm(data: { amount?: string; method?: string }) {
    if (data.amount) {
      await this.page.fill("#pay-amount", "");
      await this.page.fill("#pay-amount", data.amount);
    }
    if (data.method) await this.page.selectOption("#pay-method", data.method);
  }

  async submitPayForm() {
    await this.page.click('button:has-text("Record Payment")');
  }

  async fillCancelReason(reason: string) {
    await this.page.fill("#cancel-reason", reason);
  }

  async submitCancelForm() {
    await this.page.click('button:has-text("Cancel Debt")');
  }
}
```

- [ ] **Step 2: Create debts CRUD spec**

This test creates a tenant via API first (test data setup), then exercises the debt lifecycle through the UI.

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { DebtsPage } from "../../pages/debts.page";
import { TestApiClient } from "../../helpers/api-client";
import { loadAuthState } from "../../fixtures/auth.fixture";

test.describe("Debts", () => {
  let apiClient: TestApiClient;
  let tenantName: string;

  test.beforeAll(async () => {
    apiClient = new TestApiClient();
    const authState = loadAuthState();
    await apiClient.login(authState.email, authState.password);

    tenantName = `Debt Tenant ${Date.now()}`;
    await apiClient.createTenant({ full_name: tenantName, email: `debt-${Date.now()}@tenant.test` });
  });

  test("create debt and verify PENDING status", async ({ authedPage }) => {
    const debtsPage = new DebtsPage(authedPage);
    await debtsPage.goto();

    await debtsPage.openCreateModal();

    const description = `E2E Rent ${Date.now()}`;
    const futureDate = new Date(Date.now() + 30 * 86400000).toISOString().split("T")[0];

    await debtsPage.fillCreateForm({
      tenantName,
      type: "RENT",
      description,
      amount: "5000",
      dueDate: futureDate,
    });
    await debtsPage.submitCreateForm();

    const row = debtsPage.getRowByDescription(description);
    await expect(row).toBeVisible({ timeout: 10_000 });
    await expect(debtsPage.getStatusBadge(row)).toHaveText("PENDING");
  });

  test("pay a debt and verify status changes to PAID", async ({ authedPage }) => {
    // Create debt via API for this test
    const description = `Pay Test ${Date.now()}`;
    const futureDate = new Date(Date.now() + 30 * 86400000).toISOString().split("T")[0];
    const tenantRes = await apiClient.createTenant({ full_name: `PayTenant ${Date.now()}` });
    await apiClient.createDebt({
      tenant_id: tenantRes.id,
      debt_type: "RENT",
      description,
      original_amount: { amount: "1000", currency: "PHP" },
      due_date: futureDate,
    });

    const debtsPage = new DebtsPage(authedPage);
    await debtsPage.goto();

    const row = debtsPage.getRowByDescription(description);
    await expect(row).toBeVisible({ timeout: 10_000 });

    await debtsPage.clickPayButton(row);
    await debtsPage.submitPayForm(); // Default: full balance

    // Wait for modal to close and status to update
    await authedPage.waitForTimeout(1000);
    await debtsPage.goto(); // Refresh to see updated status

    const updatedRow = debtsPage.getRowByDescription(description);
    await expect(debtsPage.getStatusBadge(updatedRow)).toHaveText("PAID");
  });

  test("cancel a debt and verify status changes to CANCELLED", async ({ authedPage }) => {
    const description = `Cancel Test ${Date.now()}`;
    const futureDate = new Date(Date.now() + 30 * 86400000).toISOString().split("T")[0];
    const tenantRes = await apiClient.createTenant({ full_name: `CancelTenant ${Date.now()}` });
    await apiClient.createDebt({
      tenant_id: tenantRes.id,
      debt_type: "RENT",
      description,
      original_amount: { amount: "2000", currency: "PHP" },
      due_date: futureDate,
    });

    const debtsPage = new DebtsPage(authedPage);
    await debtsPage.goto();

    const row = debtsPage.getRowByDescription(description);
    await expect(row).toBeVisible({ timeout: 10_000 });

    await debtsPage.clickCancelButton(row);
    await debtsPage.fillCancelReason("E2E test cancellation");
    await debtsPage.submitCancelForm();

    await authedPage.waitForTimeout(1000);
    await debtsPage.goto();

    const updatedRow = debtsPage.getRowByDescription(description);
    await expect(debtsPage.getStatusBadge(updatedRow)).toHaveText("CANCELLED");
  });
});
```

- [ ] **Step 3: Run debts spec**

Run: `cd tests/playwright && npx playwright test specs/debts/`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add tests/playwright/pages/debts.page.ts tests/playwright/specs/debts/
git commit -m "test(e2e): add debts page object and lifecycle specs (create, pay, cancel)"
```

---

### Task 9: Transactions + Audit Page Objects + Specs

**Files:**
- Create: `tests/playwright/pages/transactions.page.ts`
- Create: `tests/playwright/pages/audit.page.ts`
- Create: `tests/playwright/specs/transactions/transactions.spec.ts`
- Create: `tests/playwright/specs/audit/audit.spec.ts`

- [ ] **Step 1: Create transactions page object**

```typescript
import type { Page } from "@playwright/test";

export class TransactionsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/transactions");
    await this.page.waitForSelector("h1:has-text('Transactions')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No transactions yet");
  }
}
```

- [ ] **Step 2: Create audit page object**

```typescript
import type { Page } from "@playwright/test";

export class AuditPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/audit");
    await this.page.waitForSelector("h1:has-text('Audit Logs')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No audit entries found");
  }
}
```

- [ ] **Step 3: Create transactions spec**

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { TransactionsPage } from "../../pages/transactions.page";

test.describe("Transactions", () => {
  test("displays transactions page with heading", async ({ authedPage }) => {
    const txPage = new TransactionsPage(authedPage);
    await txPage.goto();

    await expect(txPage.getHeading()).toHaveText("Transactions");
  });

  test("shows table headers when transactions exist", async ({ authedPage }) => {
    const txPage = new TransactionsPage(authedPage);
    await txPage.goto();

    // If transactions exist from debt pay tests, verify table structure
    const rows = txPage.getTableRows();
    const count = await rows.count();
    if (count > 0) {
      // Verify table has expected columns
      await expect(authedPage.locator("th:has-text('Date')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('Type')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('Amount')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('Method')")).toBeVisible();
    } else {
      await expect(txPage.getEmptyMessage()).toBeVisible();
    }
  });
});
```

- [ ] **Step 4: Create audit spec**

```typescript
import { test, expect } from "../../fixtures/auth.fixture";
import { AuditPage } from "../../pages/audit.page";

test.describe("Audit Logs", () => {
  test("displays audit logs page with heading", async ({ authedPage }) => {
    const auditPage = new AuditPage(authedPage);
    await auditPage.goto();

    await expect(auditPage.getHeading()).toHaveText("Audit Logs");
  });

  test("shows table headers when audit entries exist", async ({ authedPage }) => {
    const auditPage = new AuditPage(authedPage);
    await auditPage.goto();

    const rows = auditPage.getTableRows();
    const count = await rows.count();
    if (count > 0) {
      await expect(authedPage.locator("th:has-text('Timestamp')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('User')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('Action')")).toBeVisible();
      await expect(authedPage.locator("th:has-text('Resource')")).toBeVisible();
    } else {
      await expect(auditPage.getEmptyMessage()).toBeVisible();
    }
  });
});
```

- [ ] **Step 5: Run both specs**

Run: `cd tests/playwright && npx playwright test specs/transactions/ specs/audit/`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add tests/playwright/pages/transactions.page.ts tests/playwright/pages/audit.page.ts tests/playwright/specs/transactions/ tests/playwright/specs/audit/
git commit -m "test(e2e): add transactions and audit page objects and specs"
```

---

### Task 10: Sidebar Navigation Spec

**Files:**
- Create: `tests/playwright/specs/navigation/sidebar.spec.ts`

- [ ] **Step 1: Create sidebar navigation spec**

Tests that the sidebar renders expected nav items for the admin user and highlights the active route.

```typescript
import { test, expect } from "../../fixtures/auth.fixture";

test.describe("Sidebar Navigation", () => {
  test("admin user sees all navigation items", async ({ authedPage }) => {
    const sidebar = authedPage.locator("aside");

    await expect(sidebar.locator("text=Dashboard")).toBeVisible();
    await expect(sidebar.locator("text=Tenants")).toBeVisible();
    await expect(sidebar.locator("text=Properties")).toBeVisible();
    await expect(sidebar.locator("text=Debts")).toBeVisible();
    await expect(sidebar.locator("text=Transactions")).toBeVisible();
    await expect(sidebar.locator("text=Programs")).toBeVisible();
    await expect(sidebar.locator("text=Audit Logs")).toBeVisible();
  });

  test("sidebar highlights active route", async ({ authedPage }) => {
    // We're on /dashboard from authedPage fixture
    const dashboardLink = authedPage.locator('aside a[href="/dashboard"]');
    await expect(dashboardLink).toHaveClass(/bg-primary-light/);

    // Navigate to tenants
    await authedPage.locator('aside a[href="/tenants"]').click();
    await authedPage.waitForURL("**/tenants");

    const tenantsLink = authedPage.locator('aside a[href="/tenants"]');
    await expect(tenantsLink).toHaveClass(/bg-primary-light/);
  });

  test("sidebar shows user info and sign out", async ({ authedPage }) => {
    const sidebar = authedPage.locator("aside");

    await expect(sidebar.locator("text=E2E Admin")).toBeVisible();
    await expect(sidebar.locator("text=e2e-admin@guimba.test")).toBeVisible();
    await expect(sidebar.locator("text=Sign out")).toBeVisible();
  });

  test("sign out redirects to login", async ({ authedPage }) => {
    await authedPage.locator("aside button:has-text('Sign out')").click();
    await authedPage.waitForURL("**/login");
    await expect(authedPage).toHaveURL(/\/login/);
  });
});
```

- [ ] **Step 2: Run sidebar spec**

Run: `cd tests/playwright && npx playwright test specs/navigation/`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add tests/playwright/specs/navigation/
git commit -m "test(e2e): add sidebar navigation and sign-out specs"
```

---

### Task 11: Full Test Suite Run + Cleanup

**Files:**
- Remove: `tests/playwright/fixtures/.gitkeep`, `tests/playwright/helpers/.gitkeep`, `tests/playwright/pages/.gitkeep`, `tests/playwright/specs/.gitkeep`, `tests/playwright/snapshots/.gitkeep`

- [ ] **Step 1: Remove old .gitkeep placeholder files**

```bash
rm tests/playwright/fixtures/.gitkeep tests/playwright/helpers/.gitkeep tests/playwright/pages/.gitkeep tests/playwright/specs/.gitkeep tests/playwright/snapshots/.gitkeep
```

- [ ] **Step 2: Run the entire E2E test suite**

Run: `cd tests/playwright && npx playwright test`
Expected: All specs pass (auth, dashboard, tenants, properties, debts, transactions, audit, navigation)

- [ ] **Step 3: Commit**

```bash
git add -A tests/playwright/
git commit -m "test(e2e): remove placeholder gitkeep files, all E2E specs passing"
```
