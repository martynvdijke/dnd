import { test, expect } from './fixtures.js';
import { login, waitLoadingDone } from './helpers.js';

// CI-only visual smoke (task 3.7): skipped outside CI so local runs stay fast.
// Fixed 1280x720 viewport keeps layout assertions deterministic.
test.describe('Visual CI smoke', () => {
  test.skip(!process.env.CI, 'Runs only in CI');

  test.beforeEach(async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    await login(page);
  });

  test('app shell and character list render at 1280x720', async ({ page }) => {
    await waitLoadingDone(page);
    await expect(page.locator('#appSidebar')).toBeVisible();
    await expect(page.locator('#charactersView')).toBeVisible();
    await expect(page.getByTestId('new-character')).toBeVisible();
  });

  test('login page renders the auth form', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByTestId('login-submit')).toBeVisible();
  });
});
