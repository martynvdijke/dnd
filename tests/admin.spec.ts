import { test, expect, type Page } from './fixtures.js';
import { ensureNavOpen, waitLoadingDone, isMobile, login, NAV_TIMEOUT } from './helpers.js';

async function goToAdmin(page: Page) {
  await page.waitForTimeout(300);
  await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });
}

test.describe('Admin panel', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('admin link is visible for admin users', async ({ page }) => {
    test.slow();
    if (await isMobile(page)) {
      await page.goto('/admin', { waitUntil: 'domcontentloaded', timeout: 10000 });
      await expect(page.locator('#adminUsers .card-header')).toContainText('Users');
    } else {
      await ensureNavOpen(page);
      await expect(page.locator('#sidebarAdminNav')).toBeVisible();
    }
  });

  test('admin panel loads user list', async ({ page }) => {
    test.slow();
    await goToAdmin(page);
    await expect(page.locator('#adminUsers .card-header')).toContainText('Users');
    const rows = page.locator('#userTable tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('admin can create a new user', async ({ page }) => {
    test.slow();
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
    test.slow();
    const uniqueEntry = `Test Race ${Date.now()}`;
    await goToAdmin(page);

    // Open the unified Compendium tab
    await page.locator('[data-testid="admin-tabs"] button:has-text("Compendium")').click();

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
    await page.waitForSelector('#ef_name', { timeout: NAV_TIMEOUT });
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

  test('import wizard: paste, detect schema, dry-run, import, verify log', async ({ page }) => {
    test.slow();
    const stamp = Date.now();
    const name1 = `ImportWiz A ${stamp}`;
    const name2 = `ImportWiz B ${stamp}`;
    await goToAdmin(page);

    // Open the Import tab
    await page.locator('[data-testid="admin-tabs"] button:has-text("Import")').click();
    await expect(page.locator('#importSchema')).toBeVisible();

    // Paste JSON source
    await page.click('#adminImport button:has-text("Paste JSON")');
    await page.fill('#importPasteText', JSON.stringify([
      { name: name1, description: 'imported via wizard', speed: 30, size: 'Medium' },
      { name: name2, description: 'imported via wizard too', speed: 25, size: 'Small' },
    ]));
    await page.click('#adminImport button:has-text("Use This JSON")');

    // Preview loads and the detect-schema / preview buttons enable
    await expect(page.locator('#importPreview')).toBeVisible();
    await expect(page.locator('#importRecordCount')).toContainText('2 records');
    await expect(page.locator('#detectSchemaBtn')).toBeEnabled();

    // Auto-detect schema (server-side field-overlap scoring)
    await page.click('#detectSchemaBtn');
    await expect(page.locator('#importResults')).toContainText('Schema detected');
    const schemaValue = await page.locator('#importSchema').inputValue();
    expect(schemaValue).not.toBe('');

    // Auto-detect field mapping
    await page.click('#importPreview button:has-text("Auto-Detect Field Mapping")');
    await expect(page.locator('#importMapping')).toBeVisible();
    await expect(page.locator('#importMappingTable tbody tr').first()).toBeVisible();

    // Dry-run preview plan (nothing written)
    await page.click('#importPreviewPlanBtn');
    await expect(page.locator('#importResults')).toContainText('Dry-run plan');
    await expect(page.locator('#importResults')).toContainText('create: 2');

    // Execute the import
    await page.click('#importStartBtn');
    await expect(page.locator('#importResults')).toContainText('Done!', { timeout: 30000 });
    await expect(page.locator('#importResults')).toContainText('Imported 2');

    // Import log row appears with completed status
    const logRow = page.locator('#importLogBody tr').first();
    await expect(logRow).toContainText('pasted.json');
    await expect(logRow).toContainText('completed');
  });

  test('backup tab works', async ({ page }) => {
    test.slow();
    await goToAdmin(page);
    await page.click('#adminTabs button:has-text("Backup")');
    await expect(page.locator('#adminBackup .card-header').first()).toContainText('Backup Settings');
    await expect(page.locator('#backupEnabled')).toBeVisible();
  });

  test('push tab loads VAPID settings and saves', async ({ page }) => {
    test.slow();
    await goToAdmin(page);
    await page.click('#adminTabs button:has-text("Push")');
    const publicKey = page.locator('[data-testid="push-public-key"]');
    await expect(publicKey).toBeVisible();
    await expect(page.locator('[data-testid="push-save-settings"]')).toBeVisible();
    await expect(page.locator('[data-testid="push-test-send"]')).toBeVisible();

    // Saving auto-generates VAPID keys when none exist; the public key is
    // echoed back into the (readonly) field either way.
    await page.locator('[data-testid="push-save-settings"]').click();
    await expect(publicKey).toHaveValue(/.+/, { timeout: NAV_TIMEOUT });
  });
});
