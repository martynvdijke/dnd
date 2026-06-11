import { test, expect } from '@playwright/test';
import { login, waitLoadingDone, waitModalClosed } from './helpers.js';

const uniqueName = () => `J-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test.describe('Character Journal', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('creates and displays journal entry via API', async ({ page }) => {
    const name = uniqueName();
    const title = 'The Adventure Begins';
    const content = '<p>We set out from Waterdeep at dawn.</p>';

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await expect(page.locator('#sheetName')).toBeVisible();

    await page.waitForFunction(() => (window as any).currentChar?.id > 0);
    const cid = await page.evaluate(() => (window as any).currentChar.id);

    await page.evaluate(async (opts) => {
      await window.api('POST', `/api/characters/${opts.cid}/journal`, {
        entry_date: '2026-05-01',
        title: opts.title,
        entry: opts.content,
      });
    }, { cid, title, content });

    await page.click('#tabBar button:has-text("Journal")');
    await page.waitForTimeout(500);

    await expect(page.locator('#journalSection')).toContainText(title);
    await expect(page.locator('#journalSection')).toContainText('We set out from Waterdeep at dawn.');
  });

  test('journal editor modal shows TipTap toolbar', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.click('#tabBar button:has-text("Journal")');
    await page.waitForTimeout(300);

    await page.click('text=Write Entry');
    await page.waitForTimeout(500);

    await expect(page.locator('#journalToolbar .editor-btn').first()).toBeVisible();

    const editorBtns = page.locator('.editor-btn');
    const count = await editorBtns.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });

  test('journal entries are date-grouped', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.waitForFunction(() => (window as any).currentChar?.id > 0);
    const cid = await page.evaluate(() => (window as any).currentChar.id);

    await page.evaluate(async (charId) => {
      await window.api('POST', `/api/characters/${charId}/journal`, { entry_date: '2026-05-01', title: 'May Entry', entry: '<p>May content</p>' });
      await window.api('POST', `/api/characters/${charId}/journal`, { entry_date: '2026-04-15', title: 'April Entry', entry: '<p>April content</p>' });
    }, cid);

    await page.click('#tabBar button:has-text("Journal")');
    await page.waitForTimeout(500);

    await expect(page.locator('.journal-month-header').first()).toBeVisible();
    await expect(page.locator('#journalSection')).toContainText('May Entry');
    await expect(page.locator('#journalSection')).toContainText('April Entry');
  });

  test('journal entry expands and collapses', async ({ page }) => {
    const name = uniqueName();

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Cleric');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);

    await page.waitForFunction(() => (window as any).currentChar?.id > 0);
    const cid = await page.evaluate(() => (window as any).currentChar.id);

    await page.evaluate(async (charId) => {
      await window.api('POST', `/api/characters/${charId}/journal`, {
        entry_date: '2026-05-01',
        title: 'Expandable Entry',
        entry: '<p>Long content that should be hidden until expanded.</p>',
      });
    }, cid);

    await page.click('#tabBar button:has-text("Journal")');
    await page.waitForTimeout(500);

    const card = page.locator('.journal-entry-card').first();
    await expect(card).not.toHaveClass(/expanded/);

    await card.locator('.journal-entry-header').click();
    await page.waitForTimeout(300);
    await expect(card).toHaveClass(/expanded/);
  });
});
