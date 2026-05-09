import { test, expect } from '@playwright/test';

test.describe('Setup flow', () => {
  test('redirects to setup when no admin exists', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/setup/);
  });

  test('creates admin account and logs in', async ({ page }) => {
    await page.goto('/setup');
    await expect(page.locator('h1')).toContainText('Vellum');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.fill('#confirm', 'testpassword123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL(/\/app/);
    await expect(page.locator('#userName')).toContainText('admin');
  });
});

test.describe('Login flow', () => {
  test('shows login page after setup', async ({ page }) => {
    // First ensure admin exists
    await page.goto('/setup');
    const body = await page.locator('body').textContent();
    if (body?.includes('First-Time')) {
      await page.fill('#username', 'admin');
      await page.fill('#password', 'testpassword123');
      await page.fill('#confirm', 'testpassword123');
      await page.click('button[type="submit"]');
      await page.waitForURL(/\/app/);
    }

    await page.goto('/login');
    await expect(page.locator('h1')).toContainText('Vellum');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/\/app/);
  });

  test('shows error on invalid login', async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'wrongpassword');
    await page.click('button[type="submit"]');
    await expect(page.locator('#error')).toBeVisible();
  });
});
