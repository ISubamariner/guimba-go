import { test, expect } from "@playwright/test";

test.describe("Register", () => {
  test("password mismatch shows client-side error", async ({ page }) => {
    await page.goto("/register");

    await page.fill("#fullName", "Test User");
    await page.fill("#email", "mismatch@guimba.test");
    await page.fill("#password", "Password123!");
    await page.fill("#confirmPassword", "DifferentPass123!");
    await page.click('button[type="submit"]');

    await expect(page.locator(".bg-danger-light")).toHaveText("Passwords do not match");
  });

  test("short password shows client-side error", async ({ page }) => {
    await page.goto("/register");

    await page.fill("#fullName", "Test User");
    await page.fill("#email", "short@guimba.test");
    await page.fill("#password", "short");
    await page.fill("#confirmPassword", "short");
    await page.click('button[type="submit"]');

    await expect(page.locator(".bg-danger-light")).toHaveText("Password must be at least 8 characters");
  });

  test("register page has link to login", async ({ page }) => {
    await page.goto("/register");

    const loginLink = page.locator('a[href="/login"]');
    await expect(loginLink).toBeVisible();
    await expect(loginLink).toHaveText("Sign in");
  });
});
