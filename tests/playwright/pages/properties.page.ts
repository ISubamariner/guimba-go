import type { Page } from "@playwright/test";

export class PropertiesPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/properties");
    await this.page.waitForSelector("h1:has-text('Properties')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getAddButton() {
    return this.page.locator("button:has-text('Add Property')");
  }

  getTableRows() {
    return this.page.locator("tbody tr");
  }

  async openCreateModal() {
    await this.getAddButton().click();
    await this.page.waitForSelector("dialog[open]");
  }

  async fillCreateForm(data: {
    name: string;
    code: string;
    type?: string;
    size: string;
    rent?: string;
  }) {
    await this.page.fill("#prop-name", data.name);
    await this.page.fill("#prop-code", data.code);
    if (data.type) await this.page.selectOption("#prop-type", data.type);
    await this.page.fill("#prop-size", data.size);
    if (data.rent) await this.page.fill("#prop-rent", data.rent);
  }

  async submitCreateForm() {
    await this.page.click('button:has-text("Create Property")');
  }

  getRowByName(name: string) {
    return this.page.locator(`tbody tr:has-text("${name}")`);
  }
}
