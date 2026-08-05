import { test, expect } from './fixtures.js';
import { isMobile, login, NAV_TIMEOUT } from './helpers.js';

const uniqueName = () => `Search-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

async function ensureNavOpen(page) {
  const toggler = page.locator('.navbar-toggler');
  if (await toggler.isVisible()) {
    await toggler.click();
    await page.waitForTimeout(300);
  }
}

async function waitModalClosed(page) {
  await page.waitForFunction(() => {
    const modal = document.getElementById('genericModal');
    return !modal || !modal.classList.contains('show');
  }, { timeout: 10000 }).catch(() => {});
}

async function waitForSearchOverlay(page) {
  await page.waitForFunction(() => {
    const overlay = document.getElementById('searchOverlay');
    return overlay && overlay.style.display !== 'none';
  }, { timeout: NAV_TIMEOUT });
}

async function searchAndWait(page, query) {
  // Open command palette via Ctrl+K / Cmd+K shortcut (always wired up)
  await page.evaluate(() => {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true }));
  });
  await waitForSearchOverlay(page);
  // Execute search with query directly (avoids race with showSearchOverlay clearing #searchInput)
  await page.evaluate((q) => window.doSearch(q), query);
  await page.waitForTimeout(1000);
}

test.describe('Advanced Search', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
    await page.waitForTimeout(300);
  });

  test('search bar is visible in navbar', async ({ page }) => {
    test.slow();
    if (await isMobile(page)) return;
    await ensureNavOpen(page);
    await expect(page.locator('[data-testid="search-input"]')).toBeVisible();
    await expect(page.getByTestId('search-submit')).toBeVisible();
  });

  test('doSearch function exists on window', async ({ page }) => {
    test.slow();
    const exists = await page.evaluate(() => typeof window.doSearch);
    expect(exists).toBe('function');
  });

  test('search shows no results for nonsense query', async ({ page }) => {
    test.slow();
    await searchAndWait(page, 'xyznonexistent12345');
    await expect(page.locator('#cpResults')).toContainText('No Results');
  });

  test('search finds character by name', async ({ page }) => {
    test.slow();
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Elf');
    await page.fill('#newClass', 'Ranger');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    const searchTerm = name.slice(0, 10);
    await searchAndWait(page, searchTerm);
    await expect(page.locator('#cpResults')).toContainText(name);
  });

  test('search result navigates to character sheet', async ({ page }) => {
    test.slow();
    const name = uniqueName();
    await page.click('button:has-text("New Character")');
    await page.locator('#newName').waitFor({ state: 'visible', timeout: NAV_TIMEOUT });
    await page.fill('#newName', name);
    await page.fill('#newRace', 'Dwarf');
    await page.fill('#newClass', 'Fighter');
    await page.click('.modal button:has-text("Create")');
    await waitModalClosed(page);

    await searchAndWait(page, name);

    const result = page.locator('.cp-result-item').first();
    await expect(result).toBeVisible({ timeout: NAV_TIMEOUT });
    await result.click();

    await expect(page.locator('#sheetName')).toContainText(name, { timeout: 10000 });
  });
});
