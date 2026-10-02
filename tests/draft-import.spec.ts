import { test, expect } from './fixtures.js';
import { clickSecondaryNavItem, login, NAV_TIMEOUT } from './helpers.js';

const uniqueTitle = () => `Wolves-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

/** Click the One-Shots nav link, handling mobile hamburger menu or mobile More→One-Shots */
async function navigateToOneShots(page: any) {
  await clickSecondaryNavItem(page, 'oneshots', 'moreNavOneshot', 'One-Shots');
  await page.waitForSelector('#oneshotSection', { state: 'visible', timeout: NAV_TIMEOUT });
}

test.describe('One-shot JSON import', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('imports a pasted JSON draft from the one-shot toolbar', async ({ page }) => {
    await navigateToOneShots(page);

    await expect(page.locator('#oneshotSection button:has-text("Import JSON")')).toBeVisible({ timeout: NAV_TIMEOUT });
    await page.locator('#oneshotSection button:has-text("Import JSON")').click();
    await expect(page.locator('#genericModal')).toBeVisible({ timeout: NAV_TIMEOUT });

    const title = uniqueTitle();
    const draft = JSON.stringify({
      title,
      premise: 'Sheep are disappearing from the fields at night.',
      hook: 'A frightened farmer pleads for help.',
      difficulty: 'easy',
      estimated_minutes: 180,
      notes: '',
      acts: [],
      npcs: [],
      locations: [],
      encounters: [],
      clues: [],
    });
    await page.locator('#genericModalBody #draftImportJson').fill(draft);
    await page.locator('#genericModalBody #draftImportApplyBtn').click();

    // Success notice names the created adventure, then Done closes the modal.
    await expect(page.locator('#genericModalBody #draftImportNotice')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#genericModalBody #draftImportNotice')).toContainText(title);
    await page.locator('#genericModalBody #draftImportDoneBtn').click();
    await expect(page.locator('#genericModal')).toBeHidden({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#oneshotSection')).toContainText(title, { timeout: NAV_TIMEOUT });
  });

  test('rejects invalid JSON without closing the modal', async ({ page }) => {
    await navigateToOneShots(page);

    await page.locator('#oneshotSection button:has-text("Import JSON")').click();
    await expect(page.locator('#genericModal')).toBeVisible({ timeout: NAV_TIMEOUT });

    await page.locator('#genericModalBody #draftImportJson').fill('this is not json');
    await page.locator('#genericModalBody #draftImportApplyBtn').click();

    await expect(page.locator('#genericModalBody #draftImportError')).toBeVisible({ timeout: NAV_TIMEOUT });
    await expect(page.locator('#genericModalBody #draftImportError')).toContainText('not a valid JSON object');
    await expect(page.locator('#genericModal')).toBeVisible();
  });
});
