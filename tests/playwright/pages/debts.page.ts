import type { Page } from "@playwright/test";

export class DebtsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/debts");
    await this.page.waitForSelector("h1:has-text('Debts')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Debt')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No debts yet");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: {
    tenantName: string;
    type?: string;
    description: string;
    amount: string;
    dueDate: string;
  }) {
    // Select tenant by visible text
    await this.page.selectOption("#debt-tenant", { label: data.tenantName });
    if (data.type) await this.page.selectOption("#debt-type", data.type);
    await this.page.fill("#debt-desc", data.description);
    await this.page.fill("#debt-amount", data.amount);
    await this.page.fill("#debt-due", data.dueDate);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Debt")');
  }

  getRowByDescription(desc: string) {
    return this.page.locator(`tbody tr:has-text("${desc}")`);
  }

  getStatusBadge(row: ReturnType<Page["locator"]>) {
    return row.locator("span").filter({ hasText: /PENDING|PARTIAL|PAID|OVERDUE|CANCELLED/ });
  }

  async clickPayButton(row: ReturnType<Page["locator"]>) {
    await row.locator('button:has-text("Pay")').click();
    await this.page.waitForSelector("dialog[open]");
  }

  async clickCancelButton(row: ReturnType<Page["locator"]>) {
    await row.locator('button:has-text("Cancel")').click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillPayForm(data: { amount?: string; method?: string }) {
    if (data.amount) {
      await this.page.fill("#pay-amount", "");
      await this.page.fill("#pay-amount", data.amount);
    }
    if (data.method) await this.page.selectOption("#pay-method", data.method);
  }

  async submitPayForm() {
    await this.page.click('button:has-text("Record Payment")');
  }

  async fillCancelReason(reason: string) {
    await this.page.fill("#cancel-reason", reason);
  }

  async submitCancelForm() {
    await this.page.click('button:has-text("Cancel Debt")');
  }
}
