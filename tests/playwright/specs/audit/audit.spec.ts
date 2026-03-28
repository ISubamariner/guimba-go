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
