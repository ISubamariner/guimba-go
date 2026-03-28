import { test, expect } from "../../fixtures/auth.fixture";
import { PropertiesPage } from "../../pages/properties.page";

test.describe("Properties", () => {
  test("displays properties page with heading and add button", async ({ authedPage }) => {
    const propsPage = new PropertiesPage(authedPage);
    await propsPage.goto();

    await expect(propsPage.getHeading()).toHaveText("Properties");
    await expect(propsPage.getAddButton()).toBeVisible();
  });

  test("create property via modal and verify it appears in table", async ({ authedPage }) => {
    const propsPage = new PropertiesPage(authedPage);
    await propsPage.goto();

    await propsPage.openCreateModal();

    const propName = `E2E Property ${Date.now()}`;
    const propCode = `E2E-${Date.now()}`;
    await propsPage.fillCreateForm({
      name: propName,
      code: propCode,
      type: "RESIDENTIAL",
      size: "100",
      rent: "5000",
    });
    await propsPage.submitCreateForm();

    const row = propsPage.getRowByName(propName);
    await expect(row).toBeVisible({ timeout: 10_000 });
    await expect(row.locator("text=RESIDENTIAL")).toBeVisible();
    await expect(row.locator("text=Active")).toBeVisible();
  });
});
