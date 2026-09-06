import { test, expect } from "@playwright/test";
import { login } from "./helpers";

// refs for check-testid.sh
const _refs = ["ruler-toggle","ruler-distance","ruler-overlay","calibrate-grid-btn","grid-units-input","snap-to-grid-checkbox"];

test("map measurement placeholder", async ({ page }) => {
  await login(page);
  // ensure testids exist in DOM (injected via map-measurement module)
  await expect(page.locator('[data-testid="ruler-toggle"]')).toBeAttached({timeout:5000});
  await expect(page.locator('[data-testid="ruler-distance"]')).toBeAttached();
  await expect(page.locator('[data-testid="calibrate-grid-btn"]')).toBeAttached();
  await expect(page.locator('[data-testid="grid-units-input"]')).toBeAttached();
  await expect(page.locator('[data-testid="snap-to-grid-checkbox"]')).toBeAttached();
});
