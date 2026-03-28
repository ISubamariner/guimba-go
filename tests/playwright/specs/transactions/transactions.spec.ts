import { test, expect } from "../../fixtures/auth.fixture";
import { TransactionsPage } from "../../pages/transactions.page";
import { TestApiClient } from "../../helpers/api-client";
import { loadAuthState } from "../../fixtures/auth.fixture";

test.describe("Transactions", () => {
  test.beforeAll(async () => {
    // Create a transaction via API so the list is not empty
    const apiClient = new TestApiClient();
    const authState = loadAuthState();
    await apiClient.login(authState.email, authState.password);

    const tenant = await apiClient.createTenant({ full_name: `TxTenant ${Date.now()}` });
    const debt = await apiClient.createDebt({
      tenant_id: tenant.id,
      debt_type: "RENT",
      description: `Tx Test ${Date.now()}`,
      original_amount: { amount: "500", currency: "PHP" },
      due_date: new Date(Date.now() + 30 * 86400000).toISOString().split("T")[0],
    });
    await apiClient.payDebt({
      debt_id: debt.id,
      tenant_id: tenant.id,
      amount: { amount: "500", currency: "PHP" },
      payment_method: "CASH",
      transaction_date: new Date().toISOString().split("T")[0],
      description: "E2E transaction test",
    });
  });

  test("displays transactions page with heading", async ({ authedPage }) => {
    const txPage = new TransactionsPage(authedPage);
    await txPage.goto();

    await expect(txPage.getHeading()).toHaveText("Transactions");
  });

  test("shows table with expected columns", async ({ authedPage }) => {
    const txPage = new TransactionsPage(authedPage);
    await txPage.goto();

    await expect(authedPage.locator("th:has-text('Date')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('Type')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('Amount')")).toBeVisible();
    await expect(authedPage.locator("th:has-text('Method')")).toBeVisible();
    await expect(txPage.getTableRows().first()).toBeVisible();
  });
});
