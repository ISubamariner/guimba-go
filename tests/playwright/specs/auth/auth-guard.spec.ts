import { test, expect } from "@playwright/test";

test.describe("Auth Guard", () => {
  test("unauthenticated user is redirected to login from dashboard", async ({ page }) => {
    // Clear any existing tokens
    await page.goto("/login");
    await page.evaluate(() => {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
    });

    await page.goto("/dashboard");
    await page.waitForURL("**/login");
    await expect(page).toHaveURL(/\/login/);
  });

  test("unauthenticated user is redirected to login from tenants", async ({ page }) => {
    await page.goto("/login");
    await page.evaluate(() => {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
    });

    await page.goto("/tenants");
    await page.waitForURL("**/login");
    await expect(page).toHaveURL(/\/login/);
  });
});
