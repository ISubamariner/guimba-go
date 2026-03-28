import { TestApiClient } from "./api-client";
import * as fs from "fs";
import * as path from "path";

const ADMIN_EMAIL = "e2e-admin@guimba.test";
const ADMIN_NAME = "E2E Admin";
const ADMIN_PASSWORD = "TestPassword123!";
const AUTH_STATE_PATH = path.join(__dirname, "../.auth-state.json");

async function globalSetup() {
  const api = new TestApiClient();

  try {
    await api.register(ADMIN_EMAIL, ADMIN_NAME, ADMIN_PASSWORD);
  } catch {
    // User may already exist from a previous run — try login instead
    await api.login(ADMIN_EMAIL, ADMIN_PASSWORD);
  }

  const tokens = api.getTokens();
  fs.writeFileSync(AUTH_STATE_PATH, JSON.stringify({
    email: ADMIN_EMAIL,
    password: ADMIN_PASSWORD,
    fullName: ADMIN_NAME,
    tokens,
  }));
}

export default globalSetup;
