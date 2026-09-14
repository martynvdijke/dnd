import { test, expect } from './fixtures.js';
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
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

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
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

    await page.click('text=Write Entry');
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

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
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

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
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });

    const card = page.locator('.journal-entry-card').first();
    await expect(card).not.toHaveClass(/expanded/);

    await card.locator('.journal-entry-header').click();
    await expect(page.locator('body')).toBeVisible({ timeout: 2000 });
    await expect(card).toHaveClass(/expanded/);
  });

  test('journal entry persists after reload', async ({ page }) => {
    const name = uniqueName();
    const title = `Persisted Entry ${Date.now()}`;

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Rogue');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await page.waitForFunction(() => (window as any).currentChar?.id > 0);

    await page.click('#tabBar button:has-text("Journal")');
    await expect(page.locator('#journalSection')).toBeVisible();

    await page.click('text=Write Entry');
    await expect(page.locator('#journalTitle')).toBeVisible();
    await expect(page.locator('#journalEditor .ProseMirror, #journalEditor .tiptap').first()).toBeVisible({ timeout: 5000 });

    await page.fill('#journalTitle', title);
    const editor = page.locator('#journalEditor .ProseMirror, #journalEditor .tiptap').first();
    await editor.click();
    await editor.pressSequentially('Entry created via UI that should survive reload.', { delay: 10 });

    await page.locator('#genericModal button:has-text("Save")').click();
    await waitModalClosed(page);

    await expect(page.locator('#journalSection')).toContainText(title);

    await page.reload();
    await waitLoadingDone(page);

    // Re-open the same character after reload (character list is default view)
    const card = page.locator('.character-card').filter({ hasText: name });
    if (await card.isVisible().catch(() => false)) {
      await card.click();
      await waitLoadingDone(page);
    } else {
      // Fallback: sheet may have restored via hash; wait for name
      await expect(page.locator('#sheetName')).toContainText(name, { timeout: 5000 }).catch(() => {});
    }

    await page.click('#tabBar button:has-text("Journal")');
    await expect(page.locator('#journalSection')).toContainText(title);
  });

  test('notes tab creates and persists a note', async ({ page }) => {
    const name = uniqueName();
    const noteTitle = `Note ${Date.now()}`;

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Wizard');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await page.waitForFunction(() => (window as any).currentChar?.id > 0);

    await page.click('#tabBar button:has-text("Notes")');
    // HTMX loads notes_list.html into #notesSection
    await expect(page.locator('#notesSection')).toContainText('Notes', { timeout: 10000 });
    await expect(page.locator('#notesSection button:has-text("New Note")')).toBeVisible({ timeout: 10000 });

    await page.locator('#notesSection button:has-text("New Note")').click();
    // Form is hx-get into #genericModalBody
    await expect(page.locator('#genericModalBody input[name="title"]')).toBeVisible({ timeout: 10000 });

    await page.fill('#genericModalBody input[name="title"]', noteTitle);
    await page.fill('#genericModalBody textarea[name="content"]', 'Secret meeting at midnight');

    // Submit creates note via hx-post targeting #notesSection
    const respPromise = page.waitForResponse((r) => r.url().includes('/htmx/notes') && r.request().method() === 'POST', { timeout: 10000 }).catch(() => null);
    await page.locator('#genericModalBody button:has-text("Create Note")').click();
    await respPromise;
    // Modal auto-hides via script in notes_list.html; wait for notesSection swap
    await expect(page.locator('#notesSection')).toContainText(noteTitle, { timeout: 10000 });

    await page.reload();
    await waitLoadingDone(page);

    const card = page.locator('.character-card').filter({ hasText: name });
    if (await card.isVisible().catch(() => false)) {
      await card.click();
      await waitLoadingDone(page);
    }

    await page.click('#tabBar button:has-text("Notes")');
    await expect(page.locator('#notesSection')).toContainText(noteTitle, { timeout: 10000 });
  });

  test('deep link to sheet notes tab renders the notes section', async ({ page }) => {
    const name = uniqueName();
    const noteTitle = `DeepLink Note ${Date.now()}`;

    await page.click('text=New Character');
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Human');
    await page.fill('#newClass', 'Fighter');
    await page.click('text=Create');
    await waitModalClosed(page);

    await page.locator('.character-card').filter({ hasText: name }).click();
    await waitLoadingDone(page);
    await page.waitForFunction(() => (window as any).currentChar?.id > 0);
    const cid = await page.evaluate(() => (window as any).currentChar.id);

    // Seed a note via API so deep-link has content to show
    await page.evaluate(async (opts) => {
      const fd = new FormData();
      fd.append('character_id', String(opts.cid));
      fd.append('title', opts.title);
      fd.append('content', 'Deep link content');
      fd.append('category', 'general');
      fd.append('visibility', 'player');
      const token = (window as any).__apiToken || localStorage.getItem('villum-api-token-admin') || '';
      // Use HTMX endpoint via fetch with CSRF
      const csrf = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') || '';
      await fetch('/htmx/notes', { method: 'POST', body: fd, headers: { 'X-CSRF-Token': csrf, Authorization: token ? `Bearer ${token}` : '' }, credentials: 'include' });
    }, { cid, title: noteTitle });

    // Deep link via hash then reload — init() must re-open the character and tab.
    await page.evaluate((id) => { location.hash = `#/sheet/${id}/notes`; }, cid);
    await page.reload();
    await waitLoadingDone(page);

    // The character must be restored from the deep link (not just view-switched).
    await page.waitForFunction((id) => (window as any).currentChar?.id === id, cid, { timeout: 10000 });
    await expect(page.locator('#sheetView')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('#tabBar button:has-text("Notes")')).toHaveClass(/active/, { timeout: 10000 });
    await expect(page.locator('#notesSection')).toContainText(noteTitle, { timeout: 10000 });
  });
});
