import { test, expect } from "../../fixtures/auth.fixture";
import { AuditPage } from "../../pages/audit.page";
import { TestApiClient } from "../../helpers/api-client";
import { loadAuthState } from "../../fixtures/auth.fixture";

test.describe("Audit Logs", () => {
  test.beforeAll(async () => {
    // Trigger an audit entry by creating a tenant via API
    const apiClient = new TestApiClient();
    const authState = loadAuthState();
    await apiClient.login(authState.email, authState.password);
    await apiClient.createTenant({ full_name: `AuditTenant ${Date.now()}` });
  });

  test("displays audit logs page with heading", async ({ authedPage }) => {
    const auditPage = new AuditPage(authedPage);
    await auditPage.goto();

    await expect(auditPage.getHeading()).toHaveText("Audit Logs");
  });

  test("shows table with expected columns", async ({ authedPage }) => {
    const auditPage = new AuditPage(authedPage);
    await auditPage.goto();

    await expect(authedPage.locator("th:has-text('Timestamp')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('User')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('Action')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('Resource')")).toBeVisible();
    await expect(auditPage.getTableRows().first()).toBeVisible();
  });
});
