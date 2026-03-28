import type { Page } from "@playwright/test";

export class TenantsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/tenants");
    await this.page.waitForSelector("h1:has-text('Tenants')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Tenant')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  getEmptyMessage() {
    return this.page.locator("text=No tenants yet");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: { fullName: string; email?: string; phone?: string; notes?: string }) {
    await this.page.fill("#tenant-name", data.fullName);
    if (data.email) await this.page.fill("#tenant-email", data.email);
    if (data.phone) await this.page.fill("#tenant-phone", data.phone);
    if (data.notes) await this.page.fill("#tenant-notes", data.notes);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Tenant")');
  }

  getRowByName(name: string) {
    return this.page.locator(`tbody tr:has-text("${name}")`);
  }
}
