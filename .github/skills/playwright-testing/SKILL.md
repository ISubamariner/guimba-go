---
name: playwright-testing
description: "Manages Playwright E2E tests, full-stack validation, and visual regression testing. Use when user says 'write e2e test', 'playwright test', 'browser test', 'visual regression', 'screenshot test', 'create page object', 'test user flow', 'ui test', or when working with tests/playwright/ files."
---

# Playwright Testing Skill

Full-stack browser E2E testing and visual regression with Playwright.

## Project Location
All Playwright files live in `tests/playwright/` (separate from the frontend and Go test code).

## Architecture

### Page Object Model (POM)
Every page under test gets a corresponding class in `tests/playwright/pages/`:

```typescript
// tests/playwright/pages/login.page.ts
import { type Page, type Locator } from '@playwright/test';

export class LoginPage {
  readonly page: Page;
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly submitButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.emailInput = page.getByLabel('Email');
    this.passwordInput = page.getByLabel('Password');
    this.submitButton = page.getByRole('button', { name: 'Sign in' });
  }

  async goto() {
    await this.page.goto('/login');
  }

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.submitButton.click();
  }
}
```

### Test Fixtures
Custom fixtures in `tests/playwright/fixtures/` for auth state, seeded data, etc.:

```typescript
// tests/playwright/fixtures/auth.fixture.ts
import { test as base } from '@playwright/test';
import { LoginPage } from '../pages/login.page';

type Fixtures = {
  loginPage: LoginPage;
  authenticatedPage: Page;
};

export const test = base.extend<Fixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  authenticatedPage: async ({ page }, use) => {
    // Login via API (faster than UI login)
    const response = await page.request.post('/api/v1/auth/login', {
      data: { email: 'test@guimba.gov', password: 'testpass' }
    });
    const { token } = await response.json();
    await page.context().addCookies([{ name: 'auth', value: token, url: 'http://localhost:3000' }]);
    await use(page);
  },
});
```

### API Helpers for Test Setup
Use direct API calls to set up test state (faster than UI):

```typescript
// tests/playwright/helpers/api-client.ts
const BASE_URL = process.env.API_URL || 'http://localhost:8080/api/v1';

export async function createTestProgram(token: string) {
  const res = await fetch(`${BASE_URL}/programs`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Test Program', description: 'E2E test data' }),
  });
  return res.json();
}

export async function cleanupTestData(token: string) {
  await fetch(`${BASE_URL}/test/cleanup`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
  });
}
```

## Writing Specs

### UI Flow Test
```typescript
// tests/playwright/specs/programs/crud.spec.ts
import { test, expect } from '../../fixtures/auth.fixture';
import { ProgramsPage } from '../../pages/programs.page';

test.describe('Program CRUD', () => {
  test('can create a new program', async ({ authenticatedPage }) => {
    const programs = new ProgramsPage(authenticatedPage);
    await programs.goto();
    await programs.clickCreate();
    await programs.fillName('New Social Program');
    await programs.fillDescription('Helps communities');
    await programs.submit();

    await expect(authenticatedPage.getByText('New Social Program')).toBeVisible();
  });
});
```

### Full-Stack Validation Test (Browser + API + DB)
```typescript
// tests/playwright/specs/api-validation/program-api.spec.ts
import { test, expect } from '@playwright/test';

test('created program appears in API and UI', async ({ page, request }) => {
  // 1. Create via API
  const res = await request.post('/api/v1/programs', {
    data: { name: 'API-Created Program' },
  });
  expect(res.ok()).toBeTruthy();
  const { id } = await res.json();

  // 2. Verify via API
  const getRes = await request.get(`/api/v1/programs/${id}`);
  expect(getRes.ok()).toBeTruthy();

  // 3. Verify in browser
  await page.goto('/programs');
  await expect(page.getByText('API-Created Program')).toBeVisible();
});
```

### Visual Regression Test
```typescript
// tests/playwright/specs/programs/list.spec.ts
import { test, expect } from '@playwright/test';

test('programs list page matches snapshot', async ({ page }) => {
  await page.goto('/programs');
  await page.waitForLoadState('networkidle');

  // Full page screenshot comparison
  await expect(page).toHaveScreenshot('programs-list.png', {
    maxDiffPixelRatio: 0.01,
  });
});

test('empty state matches snapshot', async ({ page }) => {
  await page.goto('/programs?filter=nonexistent');
  await expect(page).toHaveScreenshot('programs-empty.png');
});
```

## Common Commands

### Run All Playwright Tests
```bash
cd tests/playwright && npx playwright test
```

### Run Specific Spec
```bash
npx playwright test specs/auth/login.spec.ts
```

### Run with UI Mode (interactive debugging)
```bash
npx playwright test --ui
```

### Update Visual Regression Baselines
```bash
npx playwright test --update-snapshots
```

### Run in Headed Mode (see the browser)
```bash
npx playwright test --headed
```

### Generate HTML Report
```bash
npx playwright show-report
```

### Run Only Tests Tagged with @smoke
```bash
npx playwright test --grep @smoke
```

## Troubleshooting

### Tests Fail with "Navigation timeout"
**Cause**: Frontend or backend not running
**Fix**: Start both services before running Playwright:
```bash
docker compose up -d && cd frontend && npm run dev & cd backend && go run cmd/server/main.go
```

### Visual Regression Fails on CI but Passes Locally
**Cause**: Font rendering, antialiasing differ across OS
**Fix**: Always run visual regression in Docker or use `maxDiffPixelRatio` tolerance:
```typescript
await expect(page).toHaveScreenshot('name.png', { maxDiffPixelRatio: 0.02 });
```

### Tests Are Slow
**Cause**: Logging in via UI for every test
**Fix**: Use API-based auth in fixtures (see `auth.fixture.ts` above). Only test login UI in `auth/login.spec.ts`.

### Flaky Tests
**Cause**: Race conditions, animations, network timing
**Fix**: Use Playwright's built-in auto-waiting. Avoid `page.waitForTimeout()`. Use `await expect(...).toBeVisible()` instead.
