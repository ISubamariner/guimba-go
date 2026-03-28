import { test, expect } from "../../fixtures/auth.fixture";

test.describe("Sidebar Navigation", () => {
  test("admin user sees all navigation items", async ({ authedPage }) => {
    const sidebar = authedPage.locator("aside");

    await expect(sidebar.locator("text=Dashboard")).toBeVisible();
    await expect(sidebar.locator("text=Tenants")).toBeVisible();
    await expect(sidebar.locator("text=Properties")).toBeVisible();
    await expect(sidebar.locator("text=Debts")).toBeVisible();
    await expect(sidebar.locator("text=Transactions")).toBeVisible();
    await expect(sidebar.locator("text=Programs")).toBeVisible();
    await expect(sidebar.locator("text=Audit Logs")).toBeVisible();
  });

  test("sidebar highlights active route", async ({ authedPage }) => {
    // We're on /dashboard from authedPage fixture
    const dashboardLink = authedPage.locator('aside a[href="/dashboard"]');
    await expect(dashboardLink).toHaveClass(/bg-primary-light/);

    // Navigate to tenants
    await authedPage.locator('aside a[href="/tenants"]').click();
    await authedPage.waitForURL("**/tenants");

    const tenantsLink = authedPage.locator('aside a[href="/tenants"]');
    await expect(tenantsLink).toHaveClass(/bg-primary-light/);
  });

  test("sidebar shows user info and sign out", async ({ authedPage }) => {
    const sidebar = authedPage.locator("aside");

    await expect(sidebar.locator("text=E2E Admin")).toBeVisible();
    await expect(sidebar.locator("text=e2e-admin@guimba.test")).toBeVisible();
    await expect(sidebar.locator("text=Sign out")).toBeVisible();
  });

  test("sign out redirects to login", async ({ authedPage }) => {
    await authedPage.locator("aside button:has-text('Sign out')").click();
    await authedPage.waitForURL("**/login");
    await expect(authedPage).toHaveURL(/\/login/);
  });
});
