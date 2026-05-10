import { test, expect } from '@playwright/test';

test.describe('Admin panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/');
  });

  test('admin link is visible for admin users', async ({ page }) => {
    await expect(page.locator('#adminLink')).toBeVisible();
  });

  test('admin panel loads user list', async ({ page }) => {
    await page.goto('/admin');
    await expect(page.locator('h1')).toContainText('Users');
    const rows = page.locator('#userTable tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('admin can create a new user', async ({ page }) => {
    await page.goto('/admin');
    await page.click('text=Add User');
    await page.fill('#addUsername', 'testuser');
    await page.fill('#addPassword', 'testpass123');
    await page.fill('#addDisplay', 'Test User');
    await page.click('text=Create');

    await expect(page.locator('#userTable tbody')).toContainText('testuser');
  });

  test('admin can manage compendium', async ({ page }) => {
    await page.goto('/admin');
    await page.click('text=Compendium');

    // Should be on compendium tab
    await expect(page.locator('h1')).toContainText('Compendium Management');

    // Test adding a race
    await page.click('text=Add');
    await page.fill('#compName', 'Test Race');
    await page.fill('#compDesc', 'A test race for testing');
    await page.fill('#compSpeed', '30');
    await page.fill('#compSize', 'Medium');
    await page.click('text=Create');

    await expect(page.locator('#compEntries')).toContainText('Test Race');
  });

  test('backup tab works', async ({ page }) => {
    await page.goto('/admin');
    await page.click('text=Backup');
    await expect(page.locator('h1')).toContainText('Backup');
    await expect(page.locator('#backupEnabled')).toBeVisible();
  });
});
