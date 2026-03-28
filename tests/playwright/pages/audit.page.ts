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
