import type { Page } from "@playwright/test";

export class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/login");
  }

  async login(email: string, password: string) {
    await this.page.fill("#email", email);
    await this.page.fill("#password", password);
    await this.page.click('button[type="submit"]');
  }

  async getErrorMessage(): Promise<string | null> {
    const error = this.page.locator(".bg-danger-light");
    if (await error.isVisible()) {
      return error.textContent();
    }
    return null;
  }

  async getHeading(): Promise<string | null> {
    return this.page.locator("h1").textContent();
  }
}
