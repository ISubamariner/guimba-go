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
