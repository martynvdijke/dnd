import { test, expect } from '@playwright/test';

test.describe('File upload and media gallery', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('#username', 'admin');
    await page.fill('#password', 'testpassword123');
    await Promise.all([
      page.waitForURL('/', { timeout: 10000 }),
      page.click('button[type="submit"]'),
    ]);
  });

  test('campaign overview has media section', async ({ page }) => {
    // Campaign 1 is seeded during setup
    await page.goto('/campaign/1', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(1500);

    // Media card should be present
    const mediaHeader = page.locator('.card-header:has-text("Media")');
    await expect(mediaHeader).toBeVisible({ timeout: 5000 });

    // Media gallery should load via HTMX
    const mediaGallery = page.locator('[id^="mediaGallery-"]');
    await expect(mediaGallery).toBeVisible({ timeout: 8000 });
  });

  test('upload button is present in media gallery', async ({ page }) => {
    await page.goto('/campaign/1', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(1500);

    // Upload button should exist
    const uploadBtn = page.locator('[id^="mediaGallery-"] button:has-text("Upload")');
    await expect(uploadBtn).toBeVisible({ timeout: 8000 });
  });

  test('empty media gallery shows empty state', async ({ page }) => {
    // Direct HTMX load for a non-existent entity
    await page.goto('/htmx/media-gallery?owner_type=campaign&owner_id=99999', { waitUntil: 'domcontentloaded' });

    const emptyState = page.locator('text=No Media Yet');
    await expect(emptyState).toBeVisible({ timeout: 5000 });
  });
});
