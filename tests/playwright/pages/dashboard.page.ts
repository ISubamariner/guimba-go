import type { Page } from "@playwright/test";

export class DashboardPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto("/dashboard");
    await this.page.waitForSelector("h1:has-text('Dashboard')");
  }

  getHeading() {
    return this.page.locator("h1");
  }

  getStatCard(label: string) {
    return this.page.locator(`p.text-muted:has-text("${label}")`).locator("..");
  }

  getStatValue(label: string) {
    return this.getStatCard(label).locator("p.text-3xl");
  }

  getActivitiesSection() {
    return this.page.locator('text="Recent Activities"').locator("..");
  }

  getActivityItems() {
    return this.page.locator(".border-b.border-border");
  }
}
