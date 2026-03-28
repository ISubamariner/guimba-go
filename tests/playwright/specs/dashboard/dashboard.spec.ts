import { test, expect } from "../../fixtures/auth.fixture";
import { DashboardPage } from "../../pages/dashboard.page";

test.describe("Dashboard", () => {
  test("displays stats cards with numeric values", async ({ authedPage }) => {
    const dashboard = new DashboardPage(authedPage);

    await expect(dashboard.getHeading()).toHaveText("Dashboard");
    await expect(dashboard.getStatValue("Total Tenants")).toBeVisible();
    await expect(dashboard.getStatValue("Total Properties")).toBeVisible();
    await expect(dashboard.getStatValue("Active Debts")).toBeVisible();
    await expect(dashboard.getStatValue("Overdue Debts")).toBeVisible();
  });

  test("displays recent activities section", async ({ authedPage }) => {
    const dashboard = new DashboardPage(authedPage);

    await expect(dashboard.getActivitiesSection()).toBeVisible();
  });
});
