# Playwright Patterns Reference

## Playwright Config
```typescript
// tests/playwright/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './specs',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { open: 'never' }],
    ['list'],
  ],

  use: {
    baseURL: process.env.FRONTEND_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  // Visual regression snapshot settings
  expect: {
    toHaveScreenshot: {
      maxDiffPixelRatio: 0.01,
      animations: 'disabled',
    },
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
    { name: 'firefox', use: { ...devices['Desktop Firefox'] } },
    { name: 'webkit', use: { ...devices['Desktop Safari'] } },
    { name: 'mobile-chrome', use: { ...devices['Pixel 5'] } },
    { name: 'mobile-safari', use: { ...devices['iPhone 13'] } },
  ],

  // Start frontend + backend before tests
  webServer: [
    {
      command: 'cd ../../frontend && npm run dev',
      url: 'http://localhost:3000',
      reuseExistingServer: !process.env.CI,
    },
    {
      command: 'cd ../../backend && go run cmd/server/main.go',
      url: 'http://localhost:8080/api/v1/health',
      reuseExistingServer: !process.env.CI,
    },
  ],
});
```

## Page Object Model Best Practices

### Naming Convention
- File: `kebab-case.page.ts` (e.g., `program-list.page.ts`)
- Class: `PascalCase` + `Page` suffix (e.g., `ProgramListPage`)
- Methods: describe the user action, not the UI element (e.g., `createProgram()` not `clickButton()`)

### Composition for Shared Components
```typescript
// Shared navigation component
export class NavBar {
  constructor(private page: Page) {}

  async goToPrograms() {
    await this.page.getByRole('link', { name: 'Programs' }).click();
  }

  async logout() {
    await this.page.getByRole('button', { name: 'Logout' }).click();
  }
}

// Page that uses the shared component
export class DashboardPage {
  readonly nav: NavBar;

  constructor(private page: Page) {
    this.nav = new NavBar(page);
  }
}
```

### Locator Strategy Priority
1. `getByRole()` — most resilient (accessibility-based)
2. `getByLabel()` — for form fields
3. `getByText()` — for static text
4. `getByTestId()` — fallback when no semantic selector exists
5. Never use CSS selectors or XPath unless absolutely required

## Visual Regression Workflow

### Initial Baseline Setup
```bash
# Run once to create baseline screenshots
npx playwright test --update-snapshots

# Commit the snapshots/ directory
git add tests/playwright/snapshots/
git commit -m "chore: add visual regression baselines"
```

### When Tests Fail on Diff
```bash
# View the diff report
npx playwright show-report

# If the change is intentional, update baselines:
npx playwright test --update-snapshots
```

### CI Integration
```yaml
# In GitHub Actions
- name: Run Playwright tests
  run: |
    cd tests/playwright
    npx playwright install --with-deps
    npx playwright test
- uses: actions/upload-artifact@v4
  if: failure()
  with:
    name: playwright-report
    path: tests/playwright/playwright-report/
```

## Test Data Management

### Seeding via API (preferred)
```typescript
test.beforeAll(async ({ request }) => {
  await request.post('/api/v1/test/seed', {
    data: { fixture: 'programs' }
  });
});

test.afterAll(async ({ request }) => {
  await request.post('/api/v1/test/cleanup');
});
```

### Isolated Test State
Each test should be independent. Never depend on state from a previous test.
Use `test.beforeEach` or fixtures to set up required state.

## Tagging & Filtering
```typescript
// Tag a test
test('@smoke can view dashboard', async ({ page }) => { ... });
test('@regression program table pagination', async ({ page }) => { ... });

// Run only smoke tests
// npx playwright test --grep @smoke
```
