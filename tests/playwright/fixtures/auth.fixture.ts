import { test as base, type Page } from "@playwright/test";
import * as fs from "fs";
import * as path from "path";

interface AuthState {
  email: string;
  password: string;
  fullName: string;
  tokens: {
    access_token: string;
    refresh_token: string;
  };
}

function loadAuthState(): AuthState {
  const filePath = path.join(__dirname, "../.auth-state.json");
  return JSON.parse(fs.readFileSync(filePath, "utf-8"));
}

export const test = base.extend<{ authedPage: Page }>({
  authedPage: async ({ page }, use) => {
    const authState = loadAuthState();

    // Navigate to app first so localStorage is on the correct origin
    await page.goto("/login");

    // Inject tokens into localStorage
    await page.evaluate((tokens) => {
      localStorage.setItem("access_token", tokens.access_token);
      localStorage.setItem("refresh_token", tokens.refresh_token);
    }, authState.tokens);

    // Navigate to dashboard — auth provider will pick up tokens
    await page.goto("/dashboard");
    await page.waitForSelector("h1:has-text('Dashboard')");

    await use(page);
  },
});

export { expect } from "@playwright/test";
export { loadAuthState };
