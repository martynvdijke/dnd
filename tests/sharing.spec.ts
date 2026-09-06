import { test, expect } from './fixtures.js';
import { login, waitLoadingDone, waitModalClosed, clickSecondaryNavItem } from './helpers.js';

const uniqueName = () => `S-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function createCharacter(page: import('@playwright/test').Page, name: string) {
  await page.click('text=New Character');
  await page.fill('#newName', name);
  await page.fill('#newRace', 'Human');
  await page.fill('#newClass', 'Fighter');
  await page.click('text=Create');
  await waitModalClosed(page);
}

test.describe('Public Sharing', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('shared links dialog lists and revokes a character link', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();

    // Share the character from the sheet (Details tab)
    await page.click('#tabBar button:has-text("Details")');
    await page.click('text=Share Character');
    await expect(page.locator('#shareUrl')).toBeVisible();
    const url = await page.locator('#shareUrl').inputValue();
    expect(url).toContain('/api/share/');
    await page.locator('button:has-text("Close")').click();
    await waitModalClosed(page);

    // Management dialog shows the new link and can revoke it
    await clickSecondaryNavItem(page, 'shared', 'more-nav', 'Shared Links');
    await expect(page.locator('[data-testid="shared-links-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="share-link-row"]')).toHaveCount(1);
    await page.locator('[data-testid="share-link-row"] button').click();
    await expect(page.locator('[data-testid="share-link-row"]')).toHaveCount(0);
    await expect(page.locator('#shareLinkList')).toContainText('No shared links yet');
  });

  test('note share button creates a public link', async ({ page }) => {
    const name = uniqueName();
    await createCharacter(page, name);
    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();
    await page.waitForFunction(() => (window as any).currentChar?.id > 0);
    const cid = await page.evaluate(() => (window as any).currentChar.id);

    // Seed a note via the API
    await page.evaluate(async (opts) => {
      await window.api('POST', '/api/notes', {
        character_id: opts.cid,
        title: 'Shared Note',
        content: 'note body',
        visibility: 'both',
        category: 'plot',
      });
    }, { cid });

    await page.click('#tabBar button:has-text("Notes")');
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('#notesSection')).toContainText('Shared Note');

    // Share the note from its card
    await page.locator('#notesSection button[title="Share note"]').click();
    await expect(page.locator('#shareUrl')).toBeVisible();
    const url = await page.locator('#shareUrl').inputValue();
    expect(url).toContain('/api/share/');

    // The shared note appears in the management dialog with a label
    await page.locator('button:has-text("Close")').click();
    await waitModalClosed(page);
    await clickSecondaryNavItem(page, 'shared', 'more-nav', 'Shared Links');
    await expect(page.locator('[data-testid="share-link-row"]').first()).toContainText('Shared Note');
  });
});
