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
