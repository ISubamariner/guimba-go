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
