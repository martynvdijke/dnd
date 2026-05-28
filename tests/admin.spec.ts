import { test, expect, type Page } from '@playwright/test';
import { ensureNavOpen, waitLoadingDone, isMobile } from './helpers.js';

async function goToAdmin(page: Page) {
  await page.waitForTimeout(300);
  await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });
}

test.describe('Admin panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login', { waitUntil: 'domcontentloaded' });
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { waitUntil: 'domcontentloaded' }),
      page.click('button[type="submit"]'),
    ]);
    await waitLoadingDone(page);
  });

  test('admin link is visible for admin users', async ({ page }) => {
    if (await isMobile(page)) {
      await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });
      await expect(page.locator('#adminUsers .card-header')).toContainText('Users');
    } else {
      await ensureNavOpen(page);
      await expect(page.locator('#adminNavItem')).toBeVisible();
    }
  });

  test('admin panel loads user list', async ({ page }) => {
    await goToAdmin(page);
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');
    const rows = page.locator('#userTable tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('admin can create a new user', async ({ page }) => {
    const name = `testuser-${Date.now()}`;
    await goToAdmin(page);
    await page.click('#adminUsers button:has-text("Add User")');
    await page.fill('#addUsername', name);
    await page.fill('#addPassword', 'testpass123');
    await page.fill('#addDisplay', 'Test User');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.locator('#userTable tbody')).toContainText(name);
  });

  test('admin can manage compendium', async ({ page }) => {
    const uniqueEntry = `Test Race ${Date.now()}`;
    await goToAdmin(page);
    await page.click('#adminTabs button:has-text("Compendium")');

    await expect(page.locator('#adminCompendium .card-header')).toContainText('Compendium Management');

    await page.click('#adminCompendium button:has-text("Add Entry")');
    await page.fill('#compName', uniqueEntry);
    await page.fill('#compDesc', 'A test race for testing');
    await page.fill('#compSpeed', '30');
    await page.fill('#compSize', 'Medium');
    await page.getByRole('button', { name: 'Create' }).click();

    await expect(page.locator('#compEntries')).toContainText(uniqueEntry);
  });

  test('backup tab works', async ({ page }) => {
    await goToAdmin(page);
    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
    await expect(page.locator('#backupEnabled')).toBeVisible();
  });
});
