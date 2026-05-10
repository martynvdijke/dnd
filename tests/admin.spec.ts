import { test, expect } from '@playwright/test';

test.describe('Admin panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/', { waitUntil: 'domcontentloaded' });
  });

  test('admin link is visible for admin users', async ({ page }) => {
    await expect(page.locator('#adminNavItem')).toBeVisible();
  });

  test('admin panel loads user list', async ({ page }) => {
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');
    const rows = page.locator('#userTable tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('admin can create a new user', async ({ page }) => {
    const name = `testuser-${Date.now()}`;
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await page.click('#adminUsers button:has-text("Add User")');
    await page.fill('#addUsername', name);
    await page.fill('#addPassword', 'testpass123');
    await page.fill('#addDisplay', 'Test User');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.locator('#userTable tbody')).toContainText(name);
  });

  test('admin can manage compendium', async ({ page }) => {
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await page.click('#adminTabs button:has-text("Compendium")');

    await expect(page.locator('#adminCompendium .card-header')).toContainText('Compendium Management');

    await page.click('#adminCompendium button:has-text("Add Entry")');
    await page.fill('#compName', 'Test Race');
    await page.fill('#compDesc', 'A test race for testing');
    await page.fill('#compSpeed', '30');
    await page.fill('#compSize', 'Medium');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.locator('#compEntries')).toContainText('Test Race');
  });

  test('backup tab works', async ({ page }) => {
    await page.goto('/admin', { waitUntil: 'domcontentloaded' });
    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
    await expect(page.locator('#backupEnabled')).toBeVisible();
  });
});
