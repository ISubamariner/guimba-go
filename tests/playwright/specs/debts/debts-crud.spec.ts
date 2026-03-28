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
