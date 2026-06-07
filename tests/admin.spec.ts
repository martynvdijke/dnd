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
      await expect(page.locator('#sidebarAdminNav')).toBeVisible();
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

    // Open the unified Compendium tab
    await page.click('#adminTabs button:has-text("Compendium")');

    // Wait for schema list to load
    await expect(page.locator('#unifiedSchemaBody')).not.toContainText('Loading');
    await expect(page.locator('#schemaCount')).not.toContainText('Loading');

    // Find the Races schema row and click Browse Entries
    const racesRow = page.locator('#unifiedSchemaBody tr').filter({ hasText: 'Races' });
    await expect(racesRow).toBeVisible();
    await racesRow.locator('button[title="Browse entries"]').click();

    // Wait for entry browser to appear
    await expect(page.locator('#unifiedEntryBrowser')).toBeVisible();

    // Click Add Entry
    await page.click('#addEntryBtn');

    // Fill in the modal form (schema-aware fields use #ef_name, #ef_description etc.)
    await page.waitForSelector('#ef_name', { timeout: 5000 });
    await page.fill('#ef_name', uniqueEntry);
    await page.fill('#ef_description', 'A test race for testing');
    await page.fill('#ef_speed', '30');

    // Select size from the dropdown
    const sizeSelect = page.locator('#ef_size');
    await sizeSelect.selectOption('Medium');

    // Save
    await page.getByRole('button', { name: 'Save' }).click();

    // Verify the entry appears in the table
    await expect(page.locator('#entryTable')).toContainText(uniqueEntry);
  });

  test('backup tab works', async ({ page }) => {
    await goToAdmin(page);
    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
    await expect(page.locator('#backupEnabled')).toBeVisible();
  });
});
