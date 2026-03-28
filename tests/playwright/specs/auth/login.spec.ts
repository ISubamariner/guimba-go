import { test, expect } from "@playwright/test";
import { LoginPage } from "../../pages/login.page";
import { loadAuthState } from "../../fixtures/auth.fixture";

test.describe("Login", () => {
  test("successful login redirects to dashboard", async ({ page }) => {
    const loginPage = new LoginPage(page);
    const authState = loadAuthState();

    await loginPage.goto();
    await loginPage.login(authState.email, authState.password);

    await page.waitForURL("**/dashboard");
    await expect(page.locator("h1")).toHaveText("Dashboard");
  });

  test("invalid credentials show error message", async ({ page }) => {
    const loginPage = new LoginPage(page);

    await loginPage.goto();
    await loginPage.login("wrong@example.com", "wrongpassword");

    await expect(page.locator(".bg-danger-light")).toBeVisible();
  });

  test("login page has link to register", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    const registerLink = page.locator('a[href="/register"]');
    await expect(registerLink).toBeVisible();
    await expect(registerLink).toHaveText("Register");
  });
});
