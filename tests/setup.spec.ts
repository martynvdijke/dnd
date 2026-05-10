import { test, expect } from '@playwright/test';

test.describe.serial('Setup and Login', () => {
  test('redirects to setup when no admin exists', async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    // Wait for the JS to run and redirect
    await page.waitForURL(/\/setup/, { timeout: 10000 });
    await expect(page.locator('h1')).toContainText('villum');
  });

  test('creates admin account and logs in', async ({ page }) => {
    await page.goto('/setup');
    await expect(page.locator('h1')).toContainText('villum');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.fill('#confirm', 'testpassword123');

    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);

    await expect(page.locator('#userName')).toContainText('admin');
  });

  test('can login with created account', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('h1')).toContainText('villum');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');

    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
  });

  test('shows error on invalid login', async ({ page }) => {
    await page.goto('/login');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'wrongpassword');

    await page.click('button[type="submit"]');
    await expect(page.locator('#error')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#error')).not.toHaveClass(/d-none/);
  });
});
